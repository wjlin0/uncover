package baidu_spider

import (
	"fmt"
	"github.com/wjlin0/uncover/sources"
	"net/http"
)

const (
	URL     = "https://www.baidu.com/s?wd=%s&pn=%d&rn=%d"
	URLInit = "https://www.baidu.com/"
	Source  = "baidu-spider"
)

type Agent struct {
	options *sources.Agent
}

func (agent *Agent) Name() string {
	return Source
}
func (agent *Agent) Query(session *sources.Session, query *sources.Query) (chan sources.Result, error) {

	results := make(chan sources.Result)

	go func() {
		defer close(results)

		cookies, err := agent.queryCookies(session)
		if err != nil {
			results <- sources.Result{Source: agent.Name(), Error: err}
			return
		}
		newQuery(session, cookies, agent, query, results).run()

	}()

	return results, nil
}

func (agent *Agent) queryURL(session *sources.Session, URL string, cookies []*http.Cookie, baidu *baiduRequest) (*http.Response, error) {

	URL = fmt.Sprintf(URL, baidu.Wd, baidu.Pn, baidu.Rn)

	request, err := sources.NewHTTPRequest(http.MethodGet, URL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	request.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	request.Header.Add("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	//request.Header.Add("Accept-Encoding", "gzip, deflate, br")
	request.Header.Add("Connection", "close")
	request.Header.Add("Cache-Control", "max-age=0")
	request.Header.Add("Upgrade-Insecure-Requests", "1")
	request.Header.Add("Referer", URLInit)
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
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	request.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	request.Header.Add("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	//request.Header.Add("Accept-Encoding", "gzip, deflate, br")
	request.Header.Add("Connection", "close")
	request.Header.Add("Cache-Control", "max-age=0")
	request.Header.Add("Upgrade-Insecure-Requests", "1")
	request.Header.Set("Referer", URLInit)

	resp, err := session.Do(request, Source)
	if err != nil {
		return nil, err
	}
	return append(resp.Cookies(), &http.Cookie{
		Name:   "kleck",
		Value:  "6408666a6bc3e6a59bfa7b1ffcb4d094",
		Path:   "/",
		Domain: ".baidu.com",
		MaxAge: 86400,
	}), nil
}
