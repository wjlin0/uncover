package hunterhow

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/wjlin0/uncover/sources"
	"net/http"
)

const (
	baseURL      = "https://api.hunter.how/"
	baseEndpoint = "search"
	Size         = 100
	Source       = "hunterhow"
)

type Agent struct{}

func (agent *Agent) Name() string {
	return Source
}

func (agent *Agent) Query(session *sources.Session, query *sources.Query) (chan sources.Result, error) {
	if session.Keys.HunterHowToken == "" {
		return nil, errors.New(fmt.Sprintf("empty %s keys please read docs %s on how to add keys ", Source, "https://github.com/wjlin0/uncover?tab=readme-ov-file#provider-configuration"))
	}

	results := make(chan sources.Result)

	go func() {
		defer close(results)

		numberOfResults := 0

		pageQuery := 1

		for {
			hunterhowRequest := &Request{
				Query:    query.Query,
				PageSize: Size, // max size is 100
				Page:     pageQuery,
			}

			if numberOfResults > query.Limit {
				break
			}

			hunterhowResponse := agent.query(hunterhowRequest.buildURL(session.Keys.HunterHowToken), session, results)
			if hunterhowResponse == nil {
				break
			}

			if len(hunterhowResponse) == 0 {
				break
			}

			numberOfResults += len(hunterhowResponse)
			pageQuery += 1
		}
	}()

	return results, nil
}

func (agent *Agent) query(URL string, session *sources.Session, results chan sources.Result) []string {
	resp, err := agent.queryURL(session, URL)
	if err != nil {
		results <- sources.Result{Source: agent.Name(), Error: err}
		return nil
	}

	var apiResponse Response
	err = json.NewDecoder(resp.Body).Decode(&apiResponse)
	if err != nil {
		results <- sources.Result{Source: agent.Name(), Error: err}
		return nil
	}
	if apiResponse.Code != http.StatusOK {
		results <- sources.Result{Source: agent.Name(), Error: errors.New(apiResponse.Message)}
		return nil
	}

	var lines []string
	for _, data := range apiResponse.Data.List {
		result := sources.Result{Source: agent.Name()}
		result.Host = data.Domain
		result.IP = data.IP
		result.Port = data.Port
		raw, _ := json.Marshal(result)
		result.Raw = raw
		results <- result
		lines = append(lines, data.Domain)
	}

	return lines
}

func (agent *Agent) queryURL(session *sources.Session, URL string) (*http.Response, error) {
	request, err := sources.NewHTTPRequest(
		http.MethodGet,
		URL,
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := session.Do(request, agent.Name())
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d received from %s", resp.StatusCode, URL)
	}
	return resp, nil
}
