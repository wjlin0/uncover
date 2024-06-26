package sources

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/projectdiscovery/ratelimit"
	"github.com/projectdiscovery/retryablehttp-go"
	errorutil "github.com/projectdiscovery/utils/errors"
)

// DefaultRateLimits of all/most of uncover are hardcoded by default to improve performance
// engine is not present in default ratelimits then user given ratelimit from cli options is used
var DefaultRateLimits = map[string]*ratelimit.Options{
	"shodan":             {Key: "shodan", MaxCount: 1, Duration: time.Second},
	"shodan-idb":         {Key: "shodan-idb", MaxCount: 1, Duration: time.Second},
	"fofa":               {Key: "fofa", MaxCount: 1, Duration: time.Second},
	"censys":             {Key: "censys", MaxCount: 1, Duration: 3 * time.Second},
	"quake":              {Key: "quake", MaxCount: 1, Duration: time.Second},
	"hunter":             {Key: "hunter", MaxCount: 15, Duration: time.Second},
	"zoomeye":            {Key: "zoomeye", MaxCount: 1, Duration: time.Second},
	"netlas":             {Key: "netlas", MaxCount: 1, Duration: time.Second},
	"criminalip":         {Key: "criminalip", MaxCount: 1, Duration: time.Second},
	"publicwww":          {Key: "publicwww", MaxCount: 1, Duration: time.Minute},
	"hunterhow":          {Key: "hunterhow", MaxCount: 1, Duration: 3 * time.Second},
	"binaryedge":         {Key: "binaryedge", MaxCount: 1, Duration: time.Second},
	"fullhunt":           {Key: "fullhunt", MaxCount: 1, Duration: time.Second},
	"zone0":              {Key: "zone0", MaxCount: 1, Duration: time.Second},
	"daydaymap":          {Key: "daydaymap", MaxCount: 1, Duration: time.Second},
	"fofa-spider":        {Key: "fofa-spider", MaxCount: 5, Duration: time.Second},
	"bing-spider":        {Key: "bing-spider", MaxCount: 1, Duration: time.Second},
	"google-spider":      {Key: "google-spider", MaxCount: 1, Duration: time.Second},
	"chinaz-spider":      {Key: "chinaz-spider", MaxCount: 1, Duration: time.Second},
	"ip138-spider":       {Key: "ip138-spider", MaxCount: 1, Duration: time.Second},
	"qianxun-spider":     {Key: "qianxun-spider", MaxCount: 1, Duration: time.Second},
	"anubis-spider":      {Key: "anubis-spider", MaxCount: 1, Duration: time.Second},
	"baidu-spider":       {Key: "baidu-spider", MaxCount: 5, Duration: time.Second},
	"sitedossier-spider": {Key: "sitedossier-spider", MaxCount: 2, Duration: time.Second},
	"yahoo-spider":       {Key: "yahoo-spider", MaxCount: 3, Duration: time.Second},
	"zoomeye-spider":     {Key: "zoomeye-spider", MaxCount: 2, Duration: time.Second},
}

// Session handles session agent sessions
type Session struct {
	Keys       *Keys
	Client     *retryablehttp.Client
	RetryMax   int
	RateLimits *ratelimit.MultiLimiter
}

func ParseProxyAuth(auth string) (string, string, bool) {
	parts := strings.SplitN(auth, ":", 2)
	if len(parts) != 2 || parts[0] == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}

// GetProxyFunc 辅助函数：获取代理设置函数
func GetProxyFunc(proxy, auth string) (func(*http.Request) (*url.URL, error), error) {
	if proxy == "" {
		return nil, nil
	}
	proxyURL, err := url.Parse(proxy)
	if err != nil {
		return nil, fmt.Errorf("proxy error %s", err)
	}
	if auth != "" {
		username, password, ok := ParseProxyAuth(auth)
		if !ok {
			return nil, fmt.Errorf("proxy error")
		}
		proxyURL.User = url.UserPassword(username, password)
	}
	return http.ProxyURL(proxyURL), nil
}
func NewSession(keys *Keys, retryMax, timeout, rateLimit int, engines []string, duration time.Duration, proxy, proxyAuth string) (*Session, error) {
	var (
		proxyFunc func(*http.Request) (*url.URL, error)
		err       error
	)
	proxyFunc = http.ProxyFromEnvironment
	if proxy != "" {
		proxyFunc, err = GetProxyFunc(proxy, proxyAuth)
		if err != nil {
			return nil, err
		}
	}

	Transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		ResponseHeaderTimeout: time.Duration(timeout) * time.Second,
		Proxy:                 proxyFunc,
	}

	httpclient := &http.Client{
		Transport: Transport,
		Timeout:   time.Duration(timeout) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	options := retryablehttp.Options{RetryMax: retryMax}
	options.RetryWaitMax = time.Duration(timeout) * time.Second
	client := retryablehttp.NewWithHTTPClient(httpclient, options)

	session := &Session{
		Client:   client,
		Keys:     keys,
		RetryMax: retryMax,
	}

	var defaultRatelimit *ratelimit.Options
	switch {
	case rateLimit > 0:
		defaultRatelimit = &ratelimit.Options{Key: "default", MaxCount: uint(rateLimit), Duration: duration}
	default:
		defaultRatelimit = &ratelimit.Options{IsUnlimited: true, Key: "default"}
	}

	session.RateLimits, err = ratelimit.NewMultiLimiter(context.Background(), defaultRatelimit)
	if err != nil {
		return nil, err
	}

	// setup ratelimit of all engines
	for _, engine := range engines {
		rateLimitOpts := DefaultRateLimits[engine]
		if rateLimitOpts == nil {
			// fallback to using default ratelimit
			rateLimitOpts = defaultRatelimit
			rateLimitOpts.Key = engine
		}
		if err = session.RateLimits.Add(rateLimitOpts); err != nil {
			return nil, errorutil.NewWithErr(err).Msgf("failed to setup ratelimit of %v got %v", engine, err)
		}
	}

	return session, nil
}

func (s *Session) Do(request *retryablehttp.Request, source string) (*http.Response, error) {
	err := s.RateLimits.Take(source)
	if err != nil {
		return nil, err
	}
	// close request connection (does not reuse connections)
	request.Close = true
	resp, err := s.Client.Do(request)
	if err != nil {
		return nil, err
	}
	//if resp.StatusCode != http.StatusOK {
	//	requestURL, _ := url.QueryUnescape(request.URL.String())
	//	return resp, fmt.Errorf("unexpected status code %d received from %s", resp.StatusCode, requestURL)
	//}
	return resp, nil
}

func ReadBody(resp *http.Response) (*bytes.Buffer, error) {
	defer resp.Body.Close()
	body := bytes.Buffer{}
	_, err := io.Copy(&body, resp.Body)
	if err != nil {
		if !strings.ContainsAny(err.Error(), "tls: user canceled") {
			return nil, err
		}
	}
	return &body, nil
}
