package bing_spider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/wjlin0/uncover/sources"
	util "github.com/wjlin0/uncover/utils"
	stringsutls "github.com/wjlin0/uncover/utils/strings"
	"io"
	"net/http"
	"net/url"
	"strings"
)

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
	isCn       bool
}

func newQuery(session *sources.Session, cookies []*http.Cookie, agent *Agent, query2 *sources.Query, result chan sources.Result, isCN bool) *query {
	return &query{
		Domain:     query2.Query,
		Subdomains: map[string]struct{}{},
		PageNum:    0,
		session:    session,
		agent:      agent,
		result:     result,
		cookies:    cookies,
		query:      query2,
		isCn:       isCN,
	}

}

func (q *query) search(domain string, filteredSubdomain string) {
	q.PageNum = 0 // 二次搜索重新置0
	for {
		bingQuery := fmt.Sprintf("site:.%s%s", domain, filteredSubdomain)
		bing := &bingRequest{
			Q:     bingQuery,
			First: q.PageNum,
		}
		var U = URL
		if q.isCn {
			U = URLCN
		}
		resp, err := q.agent.queryURL(q.session, U, q.cookies, bing)
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

		decodeBody, err := url.QueryUnescape(body.String())
		if err != nil {
			decodeBody = body.String()
		}
		for _, sub := range sources.MatchSubdomains(domain, decodeBody, false) {
			_, host, _ := util.GetProtocolHostAndPort(sub)
			subdomains = append(subdomains, host)
		}
		if q.checkSubdomains(subdomains) {
			break
		}

		q.updates(subdomains)
		q.PageNum += 10
		if !strings.Contains(body.String(), "<div class=\"sw_next\">") {
			break
		}

		if q.PageNum > Limit || len(q.Subdomains) > q.query.Limit {
			break
		}
	}
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
		return true
	}
	// 在全搜索过程中发现搜索出的结果有完全重复的结果就停止搜索
	return stringsutls.AllStringsInMap(subdomains, q.Subdomains)
}

func (q *query) run() []string {
	q.search(q.Domain, "")
	// 排除同一子域搜索结果过多的子域以发现新的子域
	for _, statement := range Filter(q.Subdomains) {
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
func Filter(subdomains map[string]struct{}) []string {
	var (
		statementsList     []string
		subdomainsListTemp []string
	)

	for k, _ := range subdomains {
		subdomainsListTemp = append(subdomainsListTemp, k)
	}

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
