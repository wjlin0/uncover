package zoomeye_spider

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/wjlin0/uncover/sources"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	URL    = "https://www.zoomeye.org/api/search?q=%s&page=%d&pageSize=%d"
	aggs   = "https://www.zoomeye.org/api/aggs?q=%s"
	Source = "zoomeye-spider"
)

type Agent struct{}

type Request struct {
	Query    string `json:"q"`
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
}

func (agent *Agent) Name() string {
	return Source
}
func (agent *Agent) Query(session *sources.Session, query *sources.Query) (chan sources.Result, error) {

	results := make(chan sources.Result)

	go func() {
		defer close(results)
		var (
			numberOfResults int
		)

		list, err := agent.queryAggsList(aggs, session, query)
		if err != nil {
			results <- sources.Result{Source: agent.Name(), Error: errors.Wrap(err, "get fofa-spider error")}
			return
		}
		for _, c := range list.Country {
			// 省级行政单位
			for _, s := range c.Subdivisions {
				decodedString, err := strconv.Unquote(`"` + s.Name + `"`)
				if err != nil {
					continue
				}
				s.Name = decodedString
				for _, cc := range s.City {
					decodedString, err = strconv.Unquote(`"` + cc.Name + `"`)
					if err != nil {
						continue
					}
					cc.Name = decodedString
					subdivisions_ := s.Name
					city_ := cc.Name

					q := fmt.Sprintf("%s%%20+subdivisions:\"%s\"%%20+city:\"%s\"", query.Query, subdivisions_, city_)
					// url 编码 q
					q = url.QueryEscape(q)

					spiderResult, err := agent.query(session, q, URL, query.Limit, results, &numberOfResults)
					if err != nil {
						results <- sources.Result{Source: agent.Name(), Error: errors.Wrap(err, "get zoomeye-spider error")}
						return
					}
					numberOfResults += len(spiderResult)
					if numberOfResults > query.Limit {
						return
					}

				}
			}
		}
	}()
	return results, nil

}
func (agent *Agent) queryURL(session *sources.Session, zoomeyeRequest *Request, URL string) (*http.Response, error) {

	spiderURL := fmt.Sprintf(URL, zoomeyeRequest.Query, zoomeyeRequest.Page, zoomeyeRequest.PageSize)
	request, err := sources.NewHTTPRequest(http.MethodGet, spiderURL, nil)
	if err != nil {
		return nil, err
	}
	if cookies := os.Getenv("ZOOMEYE_COOKIE"); cookies != "" {
		request.Header.Set("Cookie", cookies)
	}
	request.Header.Set("Referer", spiderURL)
	return session.Do(request, agent.Name())
}

func (agent *Agent) queryAggsList(URL string, session *sources.Session, query *sources.Query) (*aggsResponse, error) {
	aggsRes := &aggsResponse{}

	URL = fmt.Sprintf(URL, url.QueryEscape(url.QueryEscape(query.Query)))
	request, err := sources.NewHTTPRequest(http.MethodGet, URL, nil)
	if err != nil {
		return nil, err
	}
	if cookies := os.Getenv("ZOOMEYE_COOKIE"); cookies != "" {
		request.Header.Set("Cookie", cookies)
	}
	request.Header.Set("Referer", URL)
	resp, err := session.Do(request, Source)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err = json.NewDecoder(resp.Body).Decode(aggsRes); err != nil {
		return nil, err
	}
	if aggsRes.Status != 200 {
		return nil, fmt.Errorf("zoomeye aggs status code: %d", aggsRes.Status)
	}
	for _, co := range aggsRes.Country {
		for _, _sub := range co.Subdivisions {
			sort.Slice(_sub.City, func(i, j int) bool {
				return _sub.City[i].Count < _sub.City[j].Count
			})
		}
		sort.Slice(co.Subdivisions, func(i, j int) bool {
			return co.Subdivisions[i].Count < co.Subdivisions[j].Count
		})
	}
	var temp = &aggsResponse{
		Status: aggsRes.Status,
	}
	for _co, co := range aggsRes.Country {
		if _co > 4 {
			continue
		}
		tempCountry := &country{
			Name:   co.Name,
			Count:  co.Count,
			Label:  co.Label,
			NameZh: co.NameZh,
		}
		for i, sub := range co.Subdivisions {
			if i > 4 {
				continue
			}
			tempSub := subdivisions{
				Name:  sub.Name,
				Count: sub.Count,
			}
			if len(sub.City) > 4 {
				tempSub.City = sub.City[:4]
			} else {
				tempSub.City = sub.City
			}
			tempCountry.Subdivisions = append(tempCountry.Subdivisions, tempSub)
		}

		temp.Country = append(temp.Country, tempCountry)
	}
	return temp, nil
}

func (agent *Agent) query(session *sources.Session, q string, url string, limit int, results chan sources.Result, num *int) ([]sources.Result, error) {
	var (
		spiderResult []sources.Result
	)
	page := 1
	for {
		zoomeye := &Request{
			Query:    q,
			Page:     page,
			PageSize: 40,
		}
		if *num > limit {
			break
		}

		resp, err := agent.queryURL(session, zoomeye, url)
		if err != nil {
			continue
		}
		body, err := sources.ReadBody(resp)
		if err != nil || body == nil {
			continue
		}
		responseJson := &response{}
		// json 序列化
		if err = json.NewDecoder(body).Decode(responseJson); err != nil {
			continue
		}
		if responseJson.Status == 429 {
			return spiderResult, fmt.Errorf("zoomeye api rate limit")
		}
		if responseJson.Status != 200 {
			break
		}
		if page*40 > limit || len(spiderResult) > limit || len(responseJson.Matches) == 0 {
			break
		}

		var ss []sources.Result
		for _, matcher := range responseJson.Matches {
			s := sources.Result{Source: Source, Timestamp: time.Now().Unix()}
			switch matcher.Type {
			case "web":
				s.Host = matcher.Site
				if ips, ok := matcher.Ip.([]string); ok && len(ips) > 0 {
					s.IP = ips[0]
				}
				if matcher.Ssl != "" {
					s.Port = 443
				} else {
					s.Port = 80
				}
			case "host":
				var (
					ip string
					ok bool
				)
				if ip, ok = matcher.Ip.(string); !ok {
					continue
				}
				s.IP = ip
				s.Host = ip
				if matcher.PortInfo == nil {
					continue
				}
				s.Port = matcher.PortInfo.Port

				if strings.HasPrefix(matcher.PortInfo.Service, "http") {
					s.Url = fmt.Sprintf("%s://%s:%d", matcher.PortInfo.Service, ip, s.Port)
				} else {
					s.Url = fmt.Sprintf("http://%s", ip)
				}
			default:
				continue
			}
			ss = append(ss, s)
		}
		if len(ss) == 0 {
			break
		}
	outerLoop:
		for _, s := range ss {
			for _, spider := range spiderResult {
				if spider.Host == s.Host && spider.Port == s.Port && spider.IP == s.IP {
					continue outerLoop
				}
			}
			results <- s
			spiderResult = append(spiderResult, s)
		}

		page++
	}

	return spiderResult, nil

}
