package chinaz_spider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/projectdiscovery/gologger"
	"github.com/wjlin0/uncover/sources"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	URL    = "https://alexa.chinaz.com/%s"
	Source = "chinaz-spider"
)

type Agent struct {
	options *sources.Agent
}
type chinazRequest struct {
	Domain string `json:"domain"`
}

func (agent *Agent) Name() string {
	return Source
}

func (agent *Agent) Query(session *sources.Session, query *sources.Query) (chan sources.Result, error) {

	results := make(chan sources.Result)
	start := time.Now()
	go func() {
		defer close(results)
		chinaz := &chinazRequest{Domain: query.Query}
		sub := agent.query(URL, session, chinaz, results)
		gologger.Info().Label(agent.Name()).Msgf("query %s took %s seconds to enumerate %d results.", query.Query, time.Since(start).Round(time.Second).String(), len(sub))
	}()

	return results, nil
}

func (agent *Agent) queryURL(session *sources.Session, URL string, chinaz *chinazRequest) (*http.Response, error) {
	chinazURL := fmt.Sprintf(URL, chinaz.Domain)
	request, err := sources.NewHTTPRequest(http.MethodGet, chinazURL, nil)
	if err != nil {
		return nil, err
	}
	return session.Do(request, agent.Name())
}

func (agent *Agent) query(URL string, session *sources.Session, chinaz *chinazRequest, results chan sources.Result) (sub []string) {
	var shouldIgnoreErrors bool
	resp, err := agent.queryURL(session, URL, chinaz)
	if err != nil {
		results <- sources.Result{Source: agent.Name(), Error: err}
		return
	}
	defer resp.Body.Close()
	body := bytes.Buffer{}
	_, err = io.Copy(&body, resp.Body)
	if err != nil {
		if strings.ContainsAny(err.Error(), "tls: user canceled") {
			shouldIgnoreErrors = true
		}
		if !shouldIgnoreErrors {
			results <- sources.Result{Source: agent.Name(), Error: err}
			return
		}
	}
	sub = sources.MatchSubdomains(chinaz.Domain, body.String(), true)
	for _, ch := range sub {
		result := sources.Result{Source: agent.Name()}
		result.Host = ch
		raw, _ := json.Marshal(result)
		result.Raw = raw
		results <- result
	}
	return
}
