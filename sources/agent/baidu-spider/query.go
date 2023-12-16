package baidu_spider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/wjlin0/pathScan/pkg/util"
	"github.com/wjlin0/uncover/sources"
	stringsutls "github.com/wjlin0/uncover/utils/strings"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

var CommonSubNames = []string{"i", "w", "m", "en", "us", "zh", "w3", "app", "bbs",
	"web", "www", "job", "docs", "news", "blog", "data",
	"help", "live", "mall", "blogs", "files", "forum",
	"store", "mobile"}

type query struct {
	Domain     string              `json:"domain"`
	Subdomains map[string]struct{} `json:"subdomains"`
	PageNum    int                 `json:"page_num"`
	PerPageNum int                 `json:"per_page_num"`
	session    *sources.Session
	agent      *Agent
	cookies    []*http.Cookie
	result     chan sources.Result
	query      *sources.Query
}
type baiduRequest struct {
	Wd string `json:"wd"`
	Pn int    `json:"pn"`
	Rn int    `json:"rn"`
}

func newQuery(session *sources.Session, cookies []*http.Cookie, agent *Agent, query2 *sources.Query, result chan sources.Result) *query {
	return &query{
		Domain:     query2.Query,
		Subdomains: map[string]struct{}{},
		PageNum:    0,
		PerPageNum: 50,
		session:    session,
		agent:      agent,
		result:     result,
		cookies:    cookies,
		query:      query2,
	}

}

func (q *query) search(domain string, filteredSubdomain string) {
	q.PageNum = 0 // 二次搜索重新置0
	for {
		baiduQuery := fmt.Sprintf("site:.%s%s", domain, filteredSubdomain)
		baidu := &baiduRequest{
			Wd: baiduQuery,
			Pn: q.PageNum,
			Rn: q.PerPageNum,
		}
		resp, err := q.agent.queryURL(q.session, URL, q.cookies, baidu)
		if err != nil {
			break
		}
		body := bytes.Buffer{}
		_, err = io.Copy(&body, resp.Body)
		if err != nil && !strings.ContainsAny(err.Error(), "tls: user canceled") {
			break
		}
		resp.Body.Close()
		var subdomains []string
		if len(domain) > 12 {
			subdomains = q.redirectMatch(body)
		} else {
			decodeBody, err := url.QueryUnescape(body.String())
			if err != nil {
				decodeBody = body.String()
			}
			for _, sub := range sources.MatchSubdomains(domain, decodeBody, false) {
				_, host, _ := util.GetProtocolHostAndPort(sub)
				subdomains = append(subdomains, host)
			}
		}
		if q.checkSubdomains(subdomains) {
			break
		}
		q.updates(subdomains)
		q.PageNum += q.PerPageNum

		if !strings.Contains(body.String(), fmt.Sprintf("&pn=%d&", q.PageNum)) {
			break
		}

		if q.PageNum > q.query.Limit || len(q.Subdomains) > q.query.Limit {
			break
		}
	}
}

func (q *query) redirectMatch(body bytes.Buffer) []string {
	doc, err := htmlquery.Parse(&body)
	if err != nil {
		return nil
	}
	subdomains := make(map[string]struct{})
	wg := sync.WaitGroup{}
	lock := sync.Mutex{}
	notes := htmlquery.Find(doc, "//div[@class=\"c-row source_1Vdff OP_LOG_LINK c-gap-top-xsmall source_s_3aixw \"]/a/@href")
	for _, note := range notes {
		wg.Add(1)
		go func(sub string) {
			defer wg.Done()
			if sub == "" {
				return
			}

			for _, s := range q.matchLocation(sub) {
				_, host, _ := util.GetProtocolHostAndPort(s)
				lock.Lock()
				subdomains[host] = struct{}{}
				lock.Unlock()
			}
		}(htmlquery.InnerText(note))
	}
	wg.Wait()
	return func(subdomains map[string]struct{}) []string {
		var lists []string
		for subdomain, _ := range subdomains {
			lists = append(lists, subdomain)
		}
		return lists
	}(subdomains)
}

func (q *query) matchLocation(url string) []string {
	request, err := sources.NewHTTPRequest(http.MethodHead, url, nil)
	if err != nil {
		return nil
	}
	//start := time.Now()
	resp, err := q.session.Do(request, q.agent.Name())
	if err != nil {
		return nil
	}
	//gologger.Debug().Msgf("%s took %s seconds to enumerate %v results.", q.agent.Name(), time.Since(start).Round(time.Second).String(), url)
	if get := resp.Header.Get("Location"); get != "" {
		switch {
		case strings.HasPrefix(get, "/"):
			return sources.MatchSubdomains(q.Domain, url, true)
		default:
			sub := sources.MatchSubdomains(q.Domain, get, true)
			return sub
		}

	}
	return nil
}

func (q *query) updates(subdomains []string) {
	for _, subdomain := range subdomains {
		if subdomain == "" {
			continue
		}

		if _, ok := q.Subdomains[subdomain]; ok {
			continue
		}
		q.Subdomains[subdomain] = struct{}{}
		protocol, host, port := util.GetProtocolHostAndPort(subdomain)
		result := sources.Result{Source: q.agent.Name()}
		result.Host = host
		result.Port = port
		portStr := fmt.Sprintf("%d", port)
		result.Url = protocol + "://" + host + ":" + portStr
		raw, _ := json.Marshal(result)
		result.Raw = raw
		q.result <- result
	}
}

// checkSubdomains 检查是否已经爬取过
func (q *query) checkSubdomains(subdomains []string) bool {
	if len(subdomains) == 0 {
		return false
	}
	// 在全搜索过程中发现搜索出的结果有完全重复的结果就停止搜索
	return stringsutls.AllStringsInMap(subdomains, q.Subdomains)
}

func (q *query) run() []string {
	q.search(q.Domain, "")
	// 排除同一子域搜索结果过多的子域以发现新的子域
	for _, statement := range Filter(q.Domain, q.Subdomains) {
		q.search(q.Domain, statement)
	}
	return func(subdomains map[string]struct{}) (lists []string) {
		for subdomain, _ := range subdomains {
			lists = append(lists, subdomain)
		}
		return lists
	}(q.Subdomains)
}

// Filter 生成搜索过滤语句
// 使用搜索引擎支持的-site:语法过滤掉搜索页面较多的子域以发现新域
func Filter(domain string, subdomains map[string]struct{}) []string {
	var (
		statementsList     []string
		subdomainsListTemp []string
		subdomainsTemp     []string
	)

	for k, _ := range subdomains {
		subdomainsListTemp = append(subdomainsListTemp, k)
	}

	for _, sub := range CommonSubNames {
		subdomainsTemp = append(subdomainsTemp, fmt.Sprintf("%s.%s", sub, domain))
	}
	subdomainsListTemp = stringsutls.FindCommonStrings(subdomainsListTemp, subdomainsTemp)
	for i := 0; i < len(subdomainsListTemp); i += 2 {
		// 避免索引越界，检查 i+1 是否在范围内
		if i+1 < len(subdomainsListTemp) {
			statementsList = append(statementsList, fmt.Sprintf(" -site:%s -site:%s", subdomainsListTemp[i], subdomainsListTemp[i+1]))
		} else {
			// 如果 i+1 超出范围，只处理当前子域名
			statementsList = append(statementsList, fmt.Sprintf(" -site:%s", subdomainsListTemp[i]))
		}
	}

	return statementsList
}
