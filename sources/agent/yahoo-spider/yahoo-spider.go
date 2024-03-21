package yahoo_spider

import (
	"fmt"
	"github.com/wjlin0/uncover/sources"
	"net/http"
)

const (
	URL     = "https://search.yahoo.com/search?p=%s&b=%d&pz=%d"
	URLInit = "https://search.yahoo.com/"
	Source  = "yahoo-spider"
	Limit   = 500
)

type Agent struct {
	options *sources.Agent
}

type yahooRequest struct {
	Query   string `json:"p"`
	Page    int    `json:"b"`
	PerPage int    `json:"pz"`
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
func (agent *Agent) queryURL(session *sources.Session, URL string, cookies []*http.Cookie, yahooRequest *yahooRequest) (*http.Response, error) {
	yahooURL := fmt.Sprintf(URL, yahooRequest.Query, yahooRequest.Page, yahooRequest.PerPage)
	request, err := sources.NewHTTPRequest(http.MethodGet, yahooURL, nil)
	if err != nil {
		return nil, err
	}
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
	resp, err := session.Do(request, agent.Name())
	if err != nil {
		return nil, err
	}
	return resp.Cookies(), nil
}
