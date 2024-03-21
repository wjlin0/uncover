package censys

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/wjlin0/uncover/sources"
)

const (
	URL        = "https://search.censys.io/api/v2/hosts/search?q=%s&per_page=%d&virtual_hosts=INCLUDE"
	MaxPerPage = 100
	Source     = "censys"
)

type Agent struct{}

func (agent *Agent) Name() string {
	return Source
}

func (agent *Agent) Query(session *sources.Session, query *sources.Query) (chan sources.Result, error) {
	if session.Keys.CensysToken == "" || session.Keys.CensysSecret == "" {
		return nil, errors.New(fmt.Sprintf("empty %s keys please read docs %s on how to add keys ", Source, "https://github.com/wjlin0/uncover?tab=readme-ov-file#provider-configuration"))
	}
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		var numberOfResults int

		nextCursor := ""
		for {
			censysRequest := &CensysRequest{
				Query:   query.Query,
				PerPage: MaxPerPage,
				Cursor:  nextCursor,
			}
			censysResponse := agent.query(URL, session, censysRequest, results)
			if censysResponse == nil {
				break
			}
			nextCursor = censysResponse.Results.Links.Next
			if nextCursor == "" || numberOfResults > query.Limit || len(censysResponse.Results.Hits) == 0 {
				break
			}
			numberOfResults += len(censysResponse.Results.Hits)
		}
	}()

	return results, nil
}

func (agent *Agent) queryURL(session *sources.Session, URL string, censysRequest *CensysRequest) (*http.Response, error) {
	censysURL := fmt.Sprintf(URL, url.QueryEscape(censysRequest.Query), censysRequest.PerPage)
	if censysRequest.Cursor != "" {
		censysURL += fmt.Sprintf("&cursor=%s", censysRequest.Cursor)
	}
	request, err := sources.NewHTTPRequest(http.MethodGet, censysURL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/json")
	request.SetBasicAuth(session.Keys.CensysToken, session.Keys.CensysSecret)
	resp, err := session.Do(request, agent.Name())
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d received from %s", resp.StatusCode, censysURL)
	}
	return resp, nil
}

func (agent *Agent) query(URL string, session *sources.Session, censysRequest *CensysRequest, results chan sources.Result) *CensysResponse {
	// query certificates
	resp, err := agent.queryURL(session, URL, censysRequest)
	if err != nil {
		results <- sources.Result{Source: agent.Name(), Error: err}
		// httputil.DrainResponseBody(resp)
		return nil
	}

	censysResponse := &CensysResponse{}
	if err := json.NewDecoder(resp.Body).Decode(censysResponse); err != nil {
		results <- sources.Result{Source: agent.Name(), Error: err}
		return nil
	}

	for _, censysResult := range censysResponse.Results.Hits {
		result := sources.Result{Source: agent.Name()}
		if ip, ok := censysResult["ip"]; ok {
			result.IP = ip.(string)
		}
		if name, ok := censysResult["name"]; ok {
			result.Host = name.(string)
		}
		if services, ok := censysResult["services"]; ok {
			for _, serviceData := range services.([]interface{}) {
				if serviceData, ok := serviceData.(map[string]interface{}); ok {
					result.Port = int(serviceData["port"].(float64))
					raw, _ := json.Marshal(censysResult)
					result.Raw = raw
					results <- result
				}
			}
		} else {
			raw, _ := json.Marshal(censysResult)
			result.Raw = raw
			// only ip
			results <- result
		}
	}

	return censysResponse
}

type CensysRequest struct {
	Query   string
	PerPage int
	Cursor  string
}
