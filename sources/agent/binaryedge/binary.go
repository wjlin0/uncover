package binaryedge

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/wjlin0/uncover/sources"
	util "github.com/wjlin0/uncover/utils"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Agent struct {
	options *sources.Agent
}

const (
	URL    = "https://api.binaryedge.io/v2/query/domains/subdomain/%s?page=%d&pageSize=%d"
	Size   = 100
	Source = "binaryedge"
)

func (agent *Agent) Name() string {
	return Source
}
func (agent *Agent) Query(session *sources.Session, query *sources.Query) (chan sources.Result, error) {
	if session.Keys.BinaryedgeToken == "" {
		return nil, errors.New(fmt.Sprintf("empty %s keys please read docs %s on how to add keys ", Source, "https://github.com/wjlin0/uncover?tab=readme-ov-file#provider-configuration"))
	}
	results := make(chan sources.Result)

	go func() {
		defer close(results)
		currentPage := 1
		var numberOfResults, totalResults int

		for {
			binaryRequest := &BinaryRequest{
				Query:    query.Query,
				Page:     currentPage,
				PageSize: Size,
			}
			if query.Limit > Size*5 {
				binaryRequest.PageSize = query.Limit / 5
			}
			binaryResponse := agent.query(session, binaryRequest, results)
			if binaryResponse == nil {
				break
			}
			currentPage++
			numberOfResults += len(binaryResponse.Data)
			if totalResults == 0 {
				totalResults = binaryResponse.Total
			}
			// query certificates
			if numberOfResults > query.Limit || numberOfResults > totalResults || len(binaryResponse.Data) == 0 {
				break
			}
		}
	}()
	return results, nil
}

func (agent *Agent) query(session *sources.Session, binaryRequest *BinaryRequest, results chan sources.Result) *Response {
	var (
		shouldIgnoreErrors bool
	)
	resp, err := agent.queryURL(session, URL, binaryRequest)
	if err != nil {
		results <- sources.Result{Source: agent.Name(), Error: err}
		return nil
	}
	binaryResponse := &Response{}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		if strings.ContainsAny(err.Error(), "tls: user canceled") {
			shouldIgnoreErrors = true
		}
		if !shouldIgnoreErrors {
			results <- sources.Result{Source: agent.Name(), Error: err}
			return nil
		}
	}
	err = json.Unmarshal(body, binaryResponse)
	if err != nil {
		results <- sources.Result{Source: agent.Name(), Error: err}
		return nil
	}
	for _, binaryResult := range binaryResponse.Data {
		result := sources.Result{Source: agent.Name()}
		_, result.Host, result.Port = util.GetProtocolHostAndPort(binaryResult)
		result.IP = result.Host
		raw, _ := json.Marshal(result)
		result.Raw = raw
		results <- result
	}

	return binaryResponse

}
func (agent *Agent) queryURL(session *sources.Session, URL string, binaryRequest *BinaryRequest) (*http.Response, error) {
	binaryURL := fmt.Sprintf(URL, url.QueryEscape(binaryRequest.Query), binaryRequest.Page, binaryRequest.PageSize)
	request, err := sources.NewHTTPRequest(http.MethodGet, binaryURL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Key", session.Keys.BinaryedgeToken)
	resp, err := session.Do(request, agent.Name())
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d received from %s", resp.StatusCode, binaryURL)
	}
	return resp, nil
}
