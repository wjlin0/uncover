package daydaymap

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/wjlin0/uncover/sources"
	"net/http"
)

const (
	URL    = "https://www.daydaymap.com/api/v1/raymap/search/all"
	Fields = "ip,port,domain,service"
	Size   = 100
	Source = "daydaymap"
)

type Agent struct{}

func (agent *Agent) Name() string {
	return Source
}

func (agent *Agent) Query(session *sources.Session, query *sources.Query) (chan sources.Result, error) {
	if session.Keys.DayDayMapToken == "" {
		return nil, errors.New(fmt.Sprintf("empty %s keys please read docs %s on how to add keys ", Source, "https://github.com/wjlin0/uncover?tab=readme-ov-file#provider-configuration"))
	}

	results := make(chan sources.Result)

	go func() {
		defer close(results)

		var numberOfResults int

		page := 1
		for {
			daymapRequest := &DayDayMapRequest{
				Keyword:  query.Query,
				Fields:   Fields,
				PageSize: Size,
				Page:     page,
			}
			if query.Limit > Size*5 {
				daymapRequest.PageSize = 500
			}
			daymapResponse := agent.query(URL, session, daymapRequest, results)
			if daymapResponse == nil {
				break
			}
			size := len(daymapResponse.Data.List)
			if size == 0 || numberOfResults > query.Limit || len(daymapResponse.Data.List) == 0 || numberOfResults > size {
				break
			}
			numberOfResults += size
			page++
		}
	}()

	return results, nil
}

func (agent *Agent) queryURL(session *sources.Session, URL string, daymapRequest *DayDayMapRequest) (*http.Response, error) {
	base64Query := base64.StdEncoding.EncodeToString([]byte(daymapRequest.Keyword))
	body := fmt.Sprintf(`{"keyword":"%s","fields":"%s","page":%d,"page_size":%d}`, base64Query, daymapRequest.Fields, daymapRequest.Page, daymapRequest.PageSize)
	buffer := bytes.NewBuffer([]byte(body))
	request, err := sources.NewHTTPRequest(http.MethodPost, URL, buffer)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("API-KEY", session.Keys.DayDayMapToken)
	resp, err := session.Do(request, agent.Name())
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d received from %s", resp.StatusCode, URL)
	}

	return resp, nil
}

func (agent *Agent) query(URL string, session *sources.Session, daymapRequest *DayDayMapRequest, results chan sources.Result) *DaydayMapResponse {
	resp, err := agent.queryURL(session, URL, daymapRequest)
	if err != nil {
		results <- sources.Result{Source: agent.Name(), Error: err}
		return nil
	}
	daymapResponse := &DaydayMapResponse{}

	if err := json.NewDecoder(resp.Body).Decode(daymapResponse); err != nil {
		results <- sources.Result{Source: agent.Name(), Error: err}
		return nil
	}
	if daymapResponse.Code != 200 {
		results <- sources.Result{Source: agent.Name(), Error: fmt.Errorf(daymapResponse.Message)}
		return nil
	}

	for _, daymapResult := range daymapResponse.Data.List {
		result := sources.Result{Source: agent.Name()}
		result.IP = daymapResult.IP
		result.Port = daymapResult.Port
		if daymapResult.Domain != "" {
			result.Host = daymapResult.Domain
		} else {
			result.Host = fmt.Sprintf("%s", daymapResult.IP)
		}
		protocal := "http"
		if daymapResult.Port == 443 {
			protocal = "https"
		}
		if daymapResult.Service != "" {
			protocal = daymapResult.Service
		}

		result.Url = fmt.Sprintf("%s://%s:%d", protocal, daymapResult.IP, daymapResult.Port)
		raw, _ := json.Marshal(result)
		result.Raw = raw
		results <- result
	}
	return daymapResponse
}

type DayDayMapRequest struct {
	Keyword  string `json:"keyword,omitempty"`
	Fields   string `json:"fields,omitempty"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page-size,omitempty"`
}
