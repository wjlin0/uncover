package uncover

import (
	"context"
	sliceutil "github.com/projectdiscovery/utils/slice"
	anubis_spider "github.com/wjlin0/uncover/sources/agent/anubis-spider"
	baidu_spider "github.com/wjlin0/uncover/sources/agent/baidu-spider"
	"github.com/wjlin0/uncover/sources/agent/binaryedge"
	bing_spider "github.com/wjlin0/uncover/sources/agent/bing-spider"
	chinaz_spider "github.com/wjlin0/uncover/sources/agent/chinaz-spider"
	"github.com/wjlin0/uncover/sources/agent/daydaymap"
	fofa_spider "github.com/wjlin0/uncover/sources/agent/fofa-spider"
	"github.com/wjlin0/uncover/sources/agent/fullhunt"
	"github.com/wjlin0/uncover/sources/agent/github"
	google_spider "github.com/wjlin0/uncover/sources/agent/google-spider"
	ip138_spider "github.com/wjlin0/uncover/sources/agent/ip138-spider"
	qianxun_spider "github.com/wjlin0/uncover/sources/agent/qianxun-spider"
	rapiddns_spider "github.com/wjlin0/uncover/sources/agent/rapiddns-spider"
	"github.com/wjlin0/uncover/sources/agent/sitedossier-spider"
	yahoo_spider "github.com/wjlin0/uncover/sources/agent/yahoo-spider"
	"github.com/wjlin0/uncover/sources/agent/zone0"
	zoomeye_spider "github.com/wjlin0/uncover/sources/agent/zoomeye-spider"
	"github.com/wjlin0/uncover/utils/strings"
	"sync"
	"time"

	"github.com/projectdiscovery/gologger"
	errorutil "github.com/projectdiscovery/utils/errors"
	stringsutil "github.com/projectdiscovery/utils/strings"
	"github.com/wjlin0/uncover/sources"
	"github.com/wjlin0/uncover/sources/agent/censys"
	"github.com/wjlin0/uncover/sources/agent/criminalip"
	"github.com/wjlin0/uncover/sources/agent/fofa"
	"github.com/wjlin0/uncover/sources/agent/hunter"
	"github.com/wjlin0/uncover/sources/agent/hunterhow"
	"github.com/wjlin0/uncover/sources/agent/netlas"
	"github.com/wjlin0/uncover/sources/agent/publicwww"
	"github.com/wjlin0/uncover/sources/agent/quake"
	"github.com/wjlin0/uncover/sources/agent/shodan"
	"github.com/wjlin0/uncover/sources/agent/shodanidb"
	"github.com/wjlin0/uncover/sources/agent/zoomeye"
)

var DefaultChannelBuffSize = 32

var DefaultCallback = func(query string, agent string) string {
	return query
}

type Options struct {
	Agents   []string // Uncover Agents to use
	Queries  []string // Queries to pass to Agents
	Limit    int
	MaxRetry int
	Timeout  int
	// Note these ratelimits are used as fallback in case agent
	// ratelimit is not available in DefaultRateLimits
	RateLimit              uint          // default 30 req
	RateLimitUnit          time.Duration // default unit
	ProviderConfigLocation string
	Proxy                  string
	ProxyAuth              string
}

// Service handler of all uncover Agents
type Service struct {
	Options  *Options
	Agents   []sources.Agent
	Session  *sources.Session
	Provider *sources.Provider
	Keys     sources.Keys
}

// New creates new uncover service instance
func New(opts *Options) (*Service, error) {
	s := &Service{Agents: []sources.Agent{}, Options: opts}
	for _, v := range opts.Agents {
		switch v {
		case "shodan":
			s.Agents = append(s.Agents, &shodan.Agent{})
		case "censys":
			s.Agents = append(s.Agents, &censys.Agent{})
		case "fofa":
			s.Agents = append(s.Agents, &fofa.Agent{})
		case "shodan-idb":
			s.Agents = append(s.Agents, &shodanidb.Agent{})
		case "quake":
			s.Agents = append(s.Agents, &quake.Agent{})
		case "hunter":
			s.Agents = append(s.Agents, &hunter.Agent{})
		case "zoomeye":
			s.Agents = append(s.Agents, &zoomeye.Agent{})
		case "netlas":
			s.Agents = append(s.Agents, &netlas.Agent{})
		case "criminalip":
			s.Agents = append(s.Agents, &criminalip.Agent{})
		case "publicwww":
			s.Agents = append(s.Agents, &publicwww.Agent{})
		case "hunterhow":
			s.Agents = append(s.Agents, &hunterhow.Agent{})
		case "binaryedge":
			s.Agents = append(s.Agents, &binaryedge.Agent{})
		case "zone0":
			s.Agents = append(s.Agents, &zone0.Agent{})
		case "daydaymap":
			s.Agents = append(s.Agents, &daydaymap.Agent{})
		case "github":
			s.Agents = append(s.Agents, &github.Agent{})
		case "fullhunt":
			s.Agents = append(s.Agents, &fullhunt.Agent{})
		case "anubis-spider":
			s.Agents = append(s.Agents, &anubis_spider.Agent{})
		case "sitedossier-spider":
			s.Agents = append(s.Agents, &sitedossier_spider.Agent{})
		case "fofa-spider":
			s.Agents = append(s.Agents, &fofa_spider.Agent{})
		case "bing-spider":
			s.Agents = append(s.Agents, &bing_spider.Agent{})
		case "chinaz-spider":
			s.Agents = append(s.Agents, &chinaz_spider.Agent{})
		case "google-spider":
			s.Agents = append(s.Agents, &google_spider.Agent{})
		case "ip138-spider":
			s.Agents = append(s.Agents, &ip138_spider.Agent{})
		case "qianxun-spider":
			s.Agents = append(s.Agents, &qianxun_spider.Agent{})
		case "rapiddns-spider":
			s.Agents = append(s.Agents, &rapiddns_spider.Agent{})
		case "baidu-spider":
			s.Agents = append(s.Agents, &baidu_spider.Agent{})
		case "yahoo-spider":
			s.Agents = append(s.Agents, &yahoo_spider.Agent{})
		case "zoomeye-spider":
			s.Agents = append(s.Agents, &zoomeye_spider.Agent{})
		}
	}
	s.Provider = sources.NewProvider(opts.ProviderConfigLocation)
	s.Keys = s.Provider.GetKeys()

	if opts.RateLimit == 0 {
		opts.RateLimit = 30
	}
	if opts.RateLimitUnit == 0 {
		opts.RateLimitUnit = time.Minute
	}

	var err error
	s.Session, err = sources.NewSession(&s.Keys, opts.MaxRetry, opts.Timeout, 10, opts.Agents, opts.RateLimitUnit, opts.Proxy, opts.ProxyAuth)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Service) Execute(ctx context.Context) (<-chan sources.Result, error) {

	// unlikely but as a precaution to handle random panics check all types
	if err := s.nilCheck(); err != nil {
		return nil, err
	}
	switch {
	case len(s.Agents) == 0:
		return nil, errorutil.NewWithTag("uncover", "no agent/source specified")
	case !s.hasAnyAnonymousProvider() && !s.Provider.HasKeys():
		return nil, errorutil.NewWithTag("uncover", "agents %v requires keys but no keys were found. please read docs %s on how to add keys", s.Options.Agents, "https://github.com/wjlin0/uncover?tab=readme-ov-file#provider-configuration")
	case s.hasOnlyDestructAgent():
		return nil, errorutil.NewWithTag("uncover", "destructive agents %v cannot be used with uncover. please use command show to see destruct agents (-da)", s.Options.Agents)
	}

	megaChan := make(chan sources.Result, DefaultChannelBuffSize)
	// iterate and run all sources
	wg := &sync.WaitGroup{}
	for _, q := range s.Options.Queries {
	agentLabel:
		for _, agent := range s.Agents {
			if strings.Contains(DestructAgents(), agent.Name()) {
				gologger.Warning().Msgf("destructive agent %s cannot be used with uncover", agent.Name())
				continue agentLabel
			}
			keys := s.Provider.GetKeys()
			if keys.Empty() && !(stringsutil.EqualFoldAny(agent.Name(), AnonymousAgents()...)) {
				gologger.Error().Msgf(agent.Name(), "agent given but keys not found")
				continue agentLabel
			}
			ch, err := agent.Query(s.Session, &sources.Query{
				Query: DefaultCallback(q, agent.Name()),
				Limit: s.Options.Limit,
			})
			if err != nil {
				gologger.Error().Msgf("%s\n", err)
				continue agentLabel
			}
			wg.Add(1)
			go func(source, relay chan sources.Result, ctx context.Context) {
				defer wg.Done()
				for {
					select {
					case <-ctx.Done():
						return
					case res, ok := <-source:
						res.Timestamp = time.Now().Unix()
						if !ok {
							return
						}
						relay <- res
					}
				}
			}(ch, megaChan, ctx)
		}
	}

	// close channel when all sources return
	go func(wg *sync.WaitGroup, megaChan chan sources.Result) {
		wg.Wait()
		defer close(megaChan)
	}(wg, megaChan)

	return megaChan, nil
}

// ExecuteWithCallback ExecuteWithWriters writes output to writer along with stdout
func (s *Service) ExecuteWithCallback(ctx context.Context, callback func(result sources.Result)) error {
	ch, err := s.Execute(ctx)
	if err != nil {
		return err
	}
	if callback == nil {
		return errorutil.NewWithTag("uncover", "result callback cannot be nil")
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case result, ok := <-ch:
			if !ok {
				return nil
			}
			callback(result)
		}
	}
}

// AllAgents returns all supported uncover Agents
func (s *Service) AllAgents() []string {
	return []string{
		"shodan", "censys", "fofa", "quake", "hunter", "zoomeye", "netlas", "criminalip", "publicwww", "hunterhow", "binaryedge", "github", "fullhunt", "zone0", "daydaymap",
		"shodan-idb", "anubis-spider", "sitedossier-spider", "fofa-spider", "google-spider", "bing-spider", "chinaz-spider", "ip138-spider", "qianxun-spider", "rapiddns-spider", "baidu-spider", "yahoo-spider", "zoomeye-spider",
	}
}
func DestructAgents() []string {
	return []string{
		"qianxun-spider",
	}
}
func UncoverAgents() []string {
	return []string{
		"shodan", "censys", "fofa", "quake", "hunter", "zoomeye", "netlas", "criminalip", "publicwww", "hunterhow", "binaryedge", "github", "fullhunt", "zone0", "daydaymap",
	}
}
func AnonymousAgents() []string {
	return []string{
		"shodan-idb", "sitedossier-spider", "fofa-spider", "bing-spider", "chinaz-spider", "google-spider", "ip138-spider", "qianxun-spider", "rapiddns-spider", "anubis-spider", "baidu-spider", "yahoo-spider", "zoomeye-spider",
	}
}
func (s *Service) nilCheck() error {
	if s.Provider == nil {
		return errorutil.NewWithTag("uncover", "provider cannot be nil")
	}
	if s.Options == nil {
		return errorutil.NewWithTag("uncover", "options cannot be nil")
	}
	if s.Session == nil {
		return errorutil.NewWithTag("uncover", "session cannot be nil")
	}
	return nil
}

func (s *Service) hasAnyAnonymousProvider() bool {
	return func() bool {
		for _, str := range AnonymousAgents() {
			if stringsutil.EqualFoldAny(str, s.Options.Agents...) {
				return true
			}
		}
		return false
	}()
}
func AllAgents() []string {
	return []string{
		"shodan", "censys", "fofa", "quake", "hunter", "zoomeye", "netlas", "criminalip", "publicwww", "hunterhow", "binaryedge", "github", "fullhunt", "zone0", "daydaymap",
		"shodan-idb", "anubis-spider", "sitedossier-spider", "fofa-spider", "bing-spider", "chinaz-spider", "google-spider", "ip138-spider", "qianxun-spider", "rapiddns-spider", "baidu-spider", "yahoo-spider", "zoomeye-spider",
	}
}
func (s *Service) hasOnlyDestructAgent() bool {
	return sliceutil.ContainsItems(DestructAgents(), s.Options.Agents)
}
