package zoomeye

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/wjlin0/uncover/sources"
)

const (
	URL = "https://api.zoomeye.org/web/search?query=%s&page=%d"
)

type Agent struct{}

func (agent *Agent) Name() string {
	return "zoomeye"
}

func (agent *Agent) Query(session *sources.Session, query *sources.Query) (chan sources.Result, error) {
	if session.Keys.ZoomEyeToken == "" {
		return nil, errors.New(fmt.Sprintf("empty zoomeye keys please read docs %s on how to add keys ", "https://github.com/wjlin0/uncover?tab=readme-ov-file#provider-configuration"))
	}
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		currentPage := 1
		var numberOfResults, totalResults int

		for {
			zoomeyeRequest := &ZoomEyeRequest{
				Query: query.Query,
				Page:  currentPage,
			}

			zoomeyeResponse := agent.query(URL, session, zoomeyeRequest, results)
			if zoomeyeResponse == nil {
				break
			}
			currentPage++
			numberOfResults += len(zoomeyeResponse.Results)
			if totalResults == 0 {
				totalResults = zoomeyeResponse.Total
			}

			// query certificates
			if numberOfResults > query.Limit || numberOfResults > totalResults || len(zoomeyeResponse.Results) == 0 {
				break
			}
		}
	}()

	return results, nil
}

func (agent *Agent) queryURL(session *sources.Session, URL string, zoomeyeRequest *ZoomEyeRequest) (*http.Response, error) {
	zoomeyeURL := fmt.Sprintf(URL, url.QueryEscape(zoomeyeRequest.Query), zoomeyeRequest.Page)

	request, err := sources.NewHTTPRequest(http.MethodGet, zoomeyeURL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("API-KEY", session.Keys.ZoomEyeToken)
	resp, err := session.Do(request, agent.Name())
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d received from %s", resp.StatusCode, zoomeyeURL)
	}
	return resp, nil
}

func (agent *Agent) query(URL string, session *sources.Session, zoomeyeRequest *ZoomEyeRequest, results chan sources.Result) *ZoomEyeResponse {
	// query certificates
	resp, err := agent.queryURL(session, URL, zoomeyeRequest)
	if err != nil {
		results <- sources.Result{Source: agent.Name(), Error: err}
		return nil
	}

	zoomeyeResponse := &ZoomEyeResponse{}
	if err := json.NewDecoder(resp.Body).Decode(zoomeyeResponse); err != nil {
		results <- sources.Result{Source: agent.Name(), Error: err}
		return nil
	}

	for _, zoomeyeResult := range zoomeyeResponse.Results {
		result := sources.Result{Source: agent.Name()}
		if site, ok := zoomeyeResult["site"]; ok {
			result.Host = site.(string)
		}
		if ip, ok := zoomeyeResult["ip"]; ok {
			if ips, ok := ip.([]interface{}); ok {
				if len(ips) > 0 {
					result.IP = ips[0].(string)
				}
			}
		}

		if result.Host == "" && result.IP == "" {
			continue
		}
		switch {
		case result.Host == "":
			result.Host = result.IP
		case result.IP == "":
			result.IP = result.Host
		}

		if portinfo, ok := zoomeyeResult["portinfo"]; ok {
			if port, ok := portinfo.(map[string]interface{}); ok {
				port_ := convertPortFromValue(port["port"])
				protocal := port["service"]
				if protocal == "https" && port_ == 0 {
					port_ = 443
				}
				if protocal == "http" && port_ == 0 {
					port_ = 80
				}
				result.Port = port_
				result.Url = fmt.Sprintf("%s://%s:%d", protocal, result.Host, result.Port)
				raw, _ := json.Marshal(result)
				result.Raw = raw
			}

		} else {
			result.Port = 80
			result.Url = fmt.Sprintf("http://%s:%d", result.Host, result.Port)
			raw, _ := json.Marshal(result)
			result.Raw = raw
		}
		results <- result
	}

	return zoomeyeResponse
}

type ZoomEyeRequest struct {
	Query string
	Page  int
}

func convertPortFromValue(value interface{}) int {
	switch v := value.(type) {
	case float64:
		return int(v)
	case string:
		parsed, _ := strconv.Atoi(v)
		return parsed
	default:
		return 0
	}
}
