package bing_spider

import (
	"fmt"
	"github.com/corpix/uarand"
	"github.com/wjlin0/uncover/sources"
	"net/http"
	"strings"
)

const (
	URL     = "https://www.bing.com/search?q=%s&first=%d"
	URLCN   = "https://cn.bing.com/search?q=%s&first=%d"
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

	go func() {
		defer close(results)

		cookies, isCN, err := agent.queryCookies(session)
		if err != nil {
			results <- sources.Result{Source: agent.Name(), Error: err}
			return
		}
		newQuery(session, cookies, agent, query, results, isCN).run()
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
	request.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.71 Safari/537.36ZaloTheme/light ZaloLanguage/en")
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}
	return session.Do(request, agent.Name())
}

func (agent *Agent) queryCookies(session *sources.Session) ([]*http.Cookie, bool, error) {
	var isCN bool
	request, err := sources.NewHTTPRequest(http.MethodGet, URLInit, nil)
	if err != nil {
		return nil, false, err
	}
	request.Header.Set("User-Agent", uarand.GetRandom())
	request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	resp, err := session.Do(request, agent.Name())
	if err != nil {
		return nil, false, err
	}
	if resp.StatusCode == 302 && strings.Contains(resp.Header.Get("Location"), "cn.bing.com") {
		isCN = true
	}

	return resp.Cookies(), isCN, nil
}
