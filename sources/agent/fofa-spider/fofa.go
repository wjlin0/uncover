package fofa_spider

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/pkg/errors"
	"github.com/projectdiscovery/gologger"
	"github.com/wjlin0/pathScan/pkg/util"
	"github.com/wjlin0/uncover/sources"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const Source = "fofa-spider"

type Agent struct{}

func (agent *Agent) Name() string {
	return Source
}
func (agent *Agent) Query(session *sources.Session, query *sources.Query) (chan sources.Result, error) {

	results := make(chan sources.Result)
	start := time.Now()
	go func() {
		defer close(results)

		var numberOfResults int
		defer func() {
			gologger.Info().Label(agent.Name()).Msgf("query %s took %s seconds to enumerate %d results.", query.Query, time.Since(start).Round(time.Second).String(), numberOfResults)
		}()
		list, err := agent.queryStatsList(stats, session, query)
		if err != nil {
			results <- sources.Result{Source: agent.Name(), Error: errors.Wrap(err, "get fofa-spider error")}
			return
		}
		wg := sync.WaitGroup{}
		lock := sync.Mutex{}
		for _, c := range list.Data.Countries {
			for _, r := range c.Regions {
				wg.Add(1)
				go func(q string) {
					defer wg.Done()
					spiderResult := agent.query(session, q, URL, results)
					lock.Lock()
					numberOfResults += len(spiderResult)
					lock.Unlock()
					if numberOfResults > query.Limit {
						return
					}
				}(r.Code)
			}
		}
		wg.Wait()
	}()
	return results, nil
}

func (agent *Agent) queryStatsList(STATS string, session *sources.Session, query *sources.Query) (*foFaStatsResponse, error) {
	fofaResponse := &foFaStatsResponse{}
	qbase64 := base64.StdEncoding.EncodeToString([]byte(query.Query))
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	m := make(map[string]any)
	m["qbase64"] = qbase64
	m["full"] = false
	m["fields"] = ""
	m["ts"] = ts
	sign, err := signQuery(serialize(m))
	if err != nil {
		return nil, errors.Wrap(err, "sign error")
	}
	STATS = fmt.Sprintf(STATS, qbase64, false, "", ts, sign, appId)
	request, err := sources.NewHTTPRequest(http.MethodGet, STATS, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Referer", "https://fofa.info/")
	resp, err := session.Do(request, Source)
	if err := json.NewDecoder(resp.Body).Decode(fofaResponse); err != nil {
		return nil, err
	}
	if fofaResponse.Code == -9 {
		return nil, errors.New(fofaResponse.Message)
	}
	return fofaResponse, nil
}

func (agent *Agent) query(session *sources.Session, query string, URL string, result chan sources.Result) []sources.Result {
	var (
		shouldIgnoreErrors bool
		spiderResult       []sources.Result
	)
	resp, err := agent.queryURL(session, query, URL)
	if err != nil {
		result <- sources.Result{Source: agent.Name(), Error: err}
		return nil
	}
	defer resp.Body.Close()
	body := bytes.Buffer{}
	_, err = io.Copy(&body, resp.Body)
	if err != nil {
		if strings.ContainsAny(err.Error(), "tls: user canceled") {
			shouldIgnoreErrors = true
		}
		if !shouldIgnoreErrors {
			result <- sources.Result{Source: agent.Name(), Error: err}
			return nil
		}
	}

	doc, err := htmlquery.Parse(&body)
	if err != nil {
		return nil
	}
	notes := htmlquery.Find(doc, "//body//div[@class=\"hsxa-meta-data-list\"]/div[@class=\"el-checkbox-group\"]/div")
	if notes == nil {
		return nil
	}
	for _, n := range notes {
		r := sources.Result{}
		if portNote := htmlquery.FindOne(n, "//div[@class=\"hsxa-clearfix hsxa-meta-data-list-lv1\"]/div[@class=\"hsxa-fr\"]/a[@class=\"hsxa-port\"]/text()"); portNote != nil {
			r.Port, _ = strconv.Atoi(strings.TrimSpace(htmlquery.InnerText(portNote)))
		}
		if ipNote := htmlquery.FindOne(n, "//div[@class=\"hsxa-clearfix hsxa-pos-rel\"]/div[@class=\"hsxa-meta-data-list-main-left hsxa-fl\"]/p[2]/a[1]/text()"); ipNote != nil {
			r.IP = htmlquery.InnerText(ipNote)
		}
		if urlNote := htmlquery.FindOne(n, "//div[@class=\"hsxa-fl hsxa-meta-data-list-lv1-lf\"]//span[@class=\"hsxa-host\"]//a/@href"); urlNote != nil {
			url_ := htmlquery.InnerText(urlNote)
			_, host, port := util.GetProtocolHostAndPort(url_)
			r.Host = host
			if r.Port == 0 {
				r.Port = port
			}
			if url_ != "" {
				r.Url = url_
			}
		}
		raw, _ := json.Marshal(r)
		r.Raw = raw
		if r.Host != "" || r.IP != "" {
			result <- r
			spiderResult = append(spiderResult, r)
		}
	}
	return spiderResult
}
func (agent *Agent) queryURL(session *sources.Session, query string, URL string) (*http.Response, error) {

	spiderURL := fmt.Sprintf(URL, query, 1, 10)
	request, err := sources.NewHTTPRequest(http.MethodGet, spiderURL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Referer", URL)
	return session.Do(request, agent.Name())
}