package bing_spider

import (
	"fmt"
	"github.com/corpix/uarand"
	"github.com/projectdiscovery/gologger"
	"github.com/wjlin0/uncover/sources"
	"net/http"
	"time"
)

const (
	URL     = "https://www.bing.com/search?q=%s&first=%d"
	URLInit = "https://www.bing.com/"
	Source  = "bing-spider"
	Limit   = 100
)

type Agent struct {
	options *sources.Agent
}

type bingRequest struct {
	Q     string `json:"q"`
	First int    `json:"first"`
}

func (agent *Agent) Name() string {
	return Source
}

func (agent *Agent) Query(session *sources.Session, query *sources.Query) (chan sources.Result, error) {

	results := make(chan sources.Result)

	start := time.Now()
	go func() {
		defer close(results)
		var (
			numberOfResults int
		)
		defer func() {
			gologger.Info().Label(agent.Name()).Msgf("query %s took %s seconds to enumerate %d results.", query.Query, time.Since(start).Round(time.Second).String(), numberOfResults)
		}()
		cookies, err := agent.queryCookies(session)
		if err != nil {
			results <- sources.Result{Source: agent.Name(), Error: err}
			return
		}
		numberOfResults = len(newQuery(session, cookies, agent, query, results).run())
	}()

	return results, nil
}
func (agent *Agent) queryURL(session *sources.Session, URL string, cookies []*http.Cookie, bingRequest *bingRequest) (*http.Response, error) {

	bingURL := fmt.Sprintf(URL, bingRequest.Q, bingRequest.First)
	request, err := sources.NewHTTPRequest(http.MethodGet, bingURL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Upgrade-Insecure-Requests", "1")
	request.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	// 使用随机生成的User-Agent
	request.Header.Set("User-Agent", uarand.GetRandom())
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}
	return session.Do(request, agent.Name())
}

func (agent *Agent) queryCookies(session *sources.Session) ([]*http.Cookie, error) {
	request, err := sources.NewHTTPRequest(http.MethodGet, URLInit, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", uarand.GetRandom())
	request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	resp, err := session.Do(request, agent.Name())
	if err != nil {
		return nil, err
	}
	return resp.Cookies(), nil
}
