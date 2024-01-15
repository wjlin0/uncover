package runner

import (
	"fmt"
	"github.com/wjlin0/uncover"
	"github.com/wjlin0/uncover/utils/update"
	"os"
	"path/filepath"

	"errors"

	"github.com/projectdiscovery/goflags"
	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/gologger/formatter"
	"github.com/projectdiscovery/gologger/levels"
	errorutil "github.com/projectdiscovery/utils/errors"
	fileutil "github.com/projectdiscovery/utils/file"
	folderutil "github.com/projectdiscovery/utils/folder"
	genericutil "github.com/projectdiscovery/utils/generic"
	"github.com/wjlin0/uncover/sources"
)

var (
	// cli flags config file location
	defaultConfigLocation = filepath.Join(folderutil.AppConfigDirOrDefault(".uncover-config", "uncover"), "config.yaml")
)

// Options contains the configuration options for tuning the enumeration process.
type Options struct {
	Query             goflags.StringSlice
	Engine            goflags.StringSlice
	ConfigFile        string
	ProviderFile      string
	OutputFile        string
	OutputFields      string
	JSON              bool
	Raw               bool
	Limit             int
	Silent            bool
	Verbose           bool
	NoColor           bool
	Timeout           int
	RateLimit         int
	RateLimitMinute   int
	Retries           int
	Shodan            goflags.StringSlice
	ShodanIdb         goflags.StringSlice
	Fofa              goflags.StringSlice
	Censys            goflags.StringSlice
	Quake             goflags.StringSlice
	Netlas            goflags.StringSlice
	Hunter            goflags.StringSlice
	ZoomEye           goflags.StringSlice
	CriminalIP        goflags.StringSlice
	Publicwww         goflags.StringSlice
	HunterHow         goflags.StringSlice
	FullHunt          goflags.StringSlice
	FoFaSpider        goflags.StringSlice
	Binaryedge        goflags.StringSlice
	Zone              goflags.StringSlice
	GoogleSpider      goflags.StringSlice
	BingSpider        goflags.StringSlice
	ChinazSpider      goflags.StringSlice
	Ip138Spider       goflags.StringSlice
	RapidDNSSpider    goflags.StringSlice
	QianXunSpider     goflags.StringSlice
	SiteDossierSpider goflags.StringSlice
	AnubisSpider      goflags.StringSlice
	BaiduSpider       goflags.StringSlice
	Github            goflags.StringSlice
	YahooSpider       goflags.StringSlice

	DisableUpdateCheck bool
	Proxy              string
	ProxyAuth          string
	Location           string
}

// ParseOptions parses the command line flags provided by a user
func ParseOptions() *Options {
	options := &Options{}
	flagSet := goflags.NewFlagSet()
	flagSet.SetDescription(`quickly discover exposed assets on the internet using multiple search engines.`)

	flagSet.CreateGroup("input", "Input",
		flagSet.StringSliceVarP(&options.Query, "query", "q", nil, "search query, supports: stdin,file,config input (example: -q 'example query', -q 'query.txt')", goflags.FileStringSliceOptions),
		flagSet.StringSliceVarP(&options.Engine, "engine", "e", nil, fmt.Sprintf("search engine to query %v (default fofa)", uncover.AllAgents()), goflags.FileNormalizedStringSliceOptions),
	)

	flagSet.CreateGroup("search-engine", "Search-Engine",
		flagSet.StringSliceVarP(&options.Shodan, "shodan", "s", nil, "search query for shodan (example: -shodan 'query.txt')", goflags.FileStringSliceOptions),
		flagSet.StringSliceVarP(&options.ShodanIdb, "shodan-idb", "sd", nil, "search query for shodan-idb (example: -shodan-idb 'query.txt')", goflags.FileStringSliceOptions),
		flagSet.StringSliceVarP(&options.Fofa, "fofa", "ff", nil, "search query for fofa (example: -fofa 'query.txt')", goflags.FileStringSliceOptions),
		flagSet.StringSliceVarP(&options.Censys, "censys", "cs", nil, "search query for censys (example: -censys 'query.txt')", goflags.FileStringSliceOptions),
		flagSet.StringSliceVarP(&options.Quake, "quake", "qk", nil, "search query for quake (example: -quake 'query.txt')", goflags.FileStringSliceOptions),
		flagSet.StringSliceVarP(&options.Hunter, "hunter", "ht", nil, "search query for hunter (example: -hunter 'query.txt')", goflags.FileStringSliceOptions),
		flagSet.StringSliceVarP(&options.ZoomEye, "zoomeye", "ze", nil, "search query for zoomeye (example: -zoomeye 'query.txt')", goflags.FileStringSliceOptions),
		flagSet.StringSliceVarP(&options.Netlas, "netlas", "ne", nil, "search query for netlas (example: -netlas 'query.txt')", goflags.FileStringSliceOptions),
		flagSet.StringSliceVarP(&options.Binaryedge, "binaryedge", "be", nil, "search query for binaryedge (example: -binaryedge 'query.txt')", goflags.FileStringSliceOptions),
		flagSet.StringSliceVarP(&options.Zone, "zone0", "z0", nil, "search query for zone0 (example: -zone0 'query.txt')", goflags.FileStringSliceOptions),
		flagSet.StringSliceVarP(&options.CriminalIP, "criminalip", "cl", nil, "search query for criminalip (example: -criminalip 'query.txt')", goflags.FileStringSliceOptions),
		flagSet.StringSliceVarP(&options.Publicwww, "publicwww", "pw", nil, "search query for publicwww (example: -publicwww 'query.txt')", goflags.FileStringSliceOptions),
		flagSet.StringSliceVarP(&options.HunterHow, "hunterhow", "hh", nil, "search query for hunterhow (example: -hunterhow 'query.txt')", goflags.FileStringSliceOptions),
		flagSet.StringSliceVarP(&options.Github, "github", "gh", nil, "search query for github (example: -github 'query.txt')", goflags.FileStringSliceOptions),
		flagSet.StringSliceVarP(&options.FullHunt, "fullhunt", "fh", nil, "search query for fullhunt (example: -fullhunt 'query.txt')", goflags.FileStringSliceOptions),
		flagSet.StringSliceVarP(&options.FoFaSpider, "fofa-spider", "fs", nil, "search query for fofa-spider (example: -fofa-spider 'query.txt')", goflags.FileStringSliceOptions),
		flagSet.StringSliceVarP(&options.GoogleSpider, "google-spider", "gs", nil, "search query for google-spider (example: -google-spider 'query.txt')", goflags.FileStringSliceOptions),
		flagSet.StringSliceVarP(&options.BingSpider, "bing-spider", "bs", nil, "search query for bing-spider (example: -bing-spider 'query.txt')", goflags.FileStringSliceOptions),
		flagSet.StringSliceVarP(&options.ChinazSpider, "chinaz-spider", "czs", nil, "search query for chinaz-spider (example: -chinaz-spider 'query.txt')", goflags.FileStringSliceOptions),
		flagSet.StringSliceVarP(&options.Ip138Spider, "ip138-spider", "is", nil, "search query for ip138-spider (example: -ip138-spider 'query.txt')", goflags.FileStringSliceOptions),
		flagSet.StringSliceVarP(&options.RapidDNSSpider, "rapiddns-spider", "rs", nil, "search query for rapiddns-spider (example: -rapiddns-spider 'query.txt')", goflags.FileStringSliceOptions),
		flagSet.StringSliceVarP(&options.QianXunSpider, "qianxun-spider", "qs", nil, "search query for qianxun-spider (example: -qianxun-spider 'query.txt')", goflags.FileStringSliceOptions),
		flagSet.StringSliceVarP(&options.SiteDossierSpider, "sitedossier-spider", "sds", nil, "search query for sitedossier-spider (example: -sitedossier-spider 'query.txt')", goflags.FileStringSliceOptions),
		flagSet.StringSliceVarP(&options.AnubisSpider, "anubis-spider", "as", nil, "search query for anubis-spider (example: -anubis-spider 'query.txt')", goflags.FileStringSliceOptions),
		flagSet.StringSliceVarP(&options.BaiduSpider, "baidu-spider", "bus", nil, "search query for baidu-spider (example: -baidu-spider 'query.txt')", goflags.FileStringSliceOptions),
		flagSet.StringSliceVarP(&options.YahooSpider, "yahoo-spider", "ys", nil, "search query for yahoo-spider (example: -yahoo-spider 'query.txt')", goflags.FileStringSliceOptions),
	)

	flagSet.CreateGroup("config", "Config",
		flagSet.StringVarP(&options.ProviderFile, "provider", "pc", sources.DefaultProviderConfigLocation, "provider configuration file"),
		flagSet.StringVar(&options.ConfigFile, "config", defaultConfigLocation, "flag configuration file"),
		flagSet.IntVar(&options.Timeout, "timeout", 30, "timeout in seconds"),
		flagSet.IntVarP(&options.RateLimit, "rate-limit", "rl", 0, "maximum number of http requests to send per second"),
		flagSet.IntVarP(&options.RateLimitMinute, "rate-limit-minute", "rlm", 0, "maximum number of requests to send per minute"),
		flagSet.IntVar(&options.Retries, "retry", 2, "number of times to retry a failed request"),
		flagSet.StringVar(&options.Proxy, "proxy", "", "proxy to use for requests (example: http://localhost:1080"),
		flagSet.StringVar(&options.ProxyAuth, "proxy-auth", "", "proxy authentication in the format username:password"),
	)

	flagSet.CreateGroup("update", "Update",
		flagSet.CallbackVarP(GetUpdateCallback(), "update", "up", "update uncover to latest version"),
		flagSet.BoolVarP(&options.DisableUpdateCheck, "disable-update-check", "duc", false, "disable automatic uncover update check"),
	)

	flagSet.CreateGroup("output", "Output",
		flagSet.StringVarP(&options.OutputFile, "output", "o", "", "output file to write found results"),
		flagSet.StringVarP(&options.OutputFields, "field", "f", "ip:port", "field to display in output (ip,port,host)"),
		flagSet.BoolVarP(&options.JSON, "json", "j", false, "write output in JSONL(ines) format"),
		flagSet.BoolVarP(&options.Raw, "raw", "r", false, "write raw output as received by the remote api"),
		flagSet.IntVarP(&options.Limit, "limit", "l", 100, "limit the number of results to return"),
		flagSet.BoolVarP(&options.NoColor, "no-color", "nc", false, "disable colors in output"),
	)

	flagSet.CreateGroup("debug", "Debug",
		flagSet.BoolVar(&options.Silent, "silent", false, "show only results in output"),
		flagSet.CallbackVar(versionCallback, "version", "show version of the project"),
		flagSet.BoolVar(&options.Verbose, "v", false, "show verbose output"),
	)

	if err := flagSet.Parse(); err != nil {
		gologger.Fatal().Msg(err.Error())
	}

	options.configureOutput()
	showBanner()

	if !options.DisableUpdateCheck {
		latestVersion, err := update.CheckVersion("wjlin0", "uncover", version)
		if err != nil {
			if options.Verbose {
				gologger.Error().Msgf("uncover version check failed: %v", err.Error())
			}
		} else {
			gologger.Info().Msgf("Current uncover version %v %v", version, update.GetVersionDescription(version, latestVersion))
		}
	}

	if options.ConfigFile != defaultConfigLocation {
		_ = options.loadConfigFrom(options.ConfigFile)
	}

	if options.ProviderFile != sources.DefaultProviderConfigLocation {
		sources.DefaultProviderConfigLocation = options.ProviderFile
	}

	if genericutil.EqualsAll(0,
		len(options.Engine),
		len(options.Shodan),
		len(options.Censys),
		len(options.Quake),
		len(options.Fofa),
		len(options.ShodanIdb),
		len(options.Binaryedge),
		len(options.Zone),
		len(options.Hunter),
		len(options.ZoomEye),
		len(options.Netlas),
		len(options.CriminalIP),
		len(options.Publicwww),
		len(options.HunterHow),
		len(options.Github),
		len(options.FullHunt),
		len(options.FoFaSpider),
		len(options.GoogleSpider),
		len(options.BingSpider),
		len(options.ChinazSpider),
		len(options.Ip138Spider),
		len(options.RapidDNSSpider),
		len(options.QianXunSpider),
		len(options.SiteDossierSpider),
		len(options.AnubisSpider),
		len(options.BaiduSpider),
		len(options.YahooSpider),
	) {
		options.Engine = append(options.Engine, "fofa")
	}

	// we make the assumption that input queries aren't that much
	if fileutil.HasStdin() {
		stdchan, err := fileutil.ReadFileWithReader(os.Stdin)
		if err != nil {
			gologger.Fatal().Msgf("couldn't read stdin: %s\n", err)
		}
		for query := range stdchan {
			options.Query = append(options.Query, query)
		}
	}

	// Validate the options passed by the user and if any
	// invalid options have been used, exit.
	if err := options.validateOptions(); err != nil {
		gologger.Fatal().Msgf("Program exiting: %s\n", err)
	}

	return options
}

// configureOutput configures the output on the screen
func (options *Options) configureOutput() {
	// If the user desires verbose output, show verbose output
	if options.Verbose {
		gologger.DefaultLogger.SetMaxLevel(levels.LevelVerbose)
	}
	if options.NoColor {
		gologger.DefaultLogger.SetFormatter(formatter.NewCLI(true))
	}
	if options.Silent {
		gologger.DefaultLogger.SetMaxLevel(levels.LevelSilent)
	}
}

func (Options *Options) loadConfigFrom(location string) error {
	if !fileutil.FileExists(location) {
		return errorutil.New("config file %s does not exist", location)
	}
	return fileutil.Unmarshal(fileutil.YAML, []byte(location), Options)
}

// validateOptions validates the configuration options passed
func (options *Options) validateOptions() error {
	// Check if domain, list of domains, or stdin info was provided.
	// If none was provided, then return.
	if genericutil.EqualsAll(0,
		len(options.Query),
		len(options.Shodan),
		len(options.Censys),
		len(options.Quake),
		len(options.Fofa),
		len(options.ShodanIdb),
		len(options.Hunter),
		len(options.Binaryedge),
		len(options.Zone),
		len(options.ZoomEye),
		len(options.Netlas),
		len(options.CriminalIP),
		len(options.Publicwww),
		len(options.HunterHow),
		len(options.Github),
		len(options.FullHunt),
		len(options.FoFaSpider),
		len(options.GoogleSpider),
		len(options.BingSpider),
		len(options.ChinazSpider),
		len(options.Ip138Spider),
		len(options.RapidDNSSpider),
		len(options.QianXunSpider),
		len(options.SiteDossierSpider),
		len(options.AnubisSpider),
		len(options.BaiduSpider),
		len(options.YahooSpider),
	) {
		return errors.New("no query provided")
	}

	// Both verbose and silent flags were used
	if options.Verbose && options.Silent {
		return errors.New("both verbose and silent mode specified")
	}

	// Validate threads and options
	if genericutil.EqualsAll(0,
		len(options.Engine),
		len(options.Shodan),
		len(options.Censys),
		len(options.Quake),
		len(options.Fofa),
		len(options.ShodanIdb),
		len(options.Hunter),
		len(options.ZoomEye),
		len(options.Netlas),
		len(options.Binaryedge),
		len(options.Zone),
		len(options.CriminalIP),
		len(options.Publicwww),
		len(options.HunterHow),
		len(options.Github),
		len(options.FullHunt),
		len(options.FoFaSpider),
		len(options.GoogleSpider),
		len(options.BingSpider),
		len(options.ChinazSpider),
		len(options.Ip138Spider),
		len(options.RapidDNSSpider),
		len(options.QianXunSpider),
		len(options.SiteDossierSpider),
		len(options.AnubisSpider),
		len(options.BaiduSpider),
		len(options.YahooSpider),
	) {
		return errors.New("no engine specified")
	}

	return nil
}

func versionCallback() {
	gologger.Info().Msgf("Current Version: %s\n", version)
	gologger.Info().Msgf("Uncover ConfigDir: %s\n", folderutil.AppConfigDirOrDefault(".uncover-config", "uncover"))
	os.Exit(0)
}

func appendQuery(options *Options, name string, queries ...string) {
	if len(queries) > 0 {
		options.Engine = append(options.Engine, name)
		options.Query = append(options.Query, queries...)
	}
}

func appendAllQueries(options *Options) {
	appendQuery(options, "shodan", options.Shodan...)
	appendQuery(options, "shodan-idb", options.ShodanIdb...)
	appendQuery(options, "fofa", options.Fofa...)
	appendQuery(options, "censys", options.Censys...)
	appendQuery(options, "quake", options.Quake...)
	appendQuery(options, "hunter", options.Hunter...)
	appendQuery(options, "zoomeye", options.ZoomEye...)
	appendQuery(options, "netlas", options.Netlas...)
	appendQuery(options, "criminalip", options.CriminalIP...)
	appendQuery(options, "publicwww", options.Publicwww...)
	appendQuery(options, "hunterhow", options.HunterHow...)
	appendQuery(options, "binaryedge", options.Binaryedge...)
	appendQuery(options, "github", options.Github...)
	appendQuery(options, "zone0", options.Zone...)
	appendQuery(options, "fullhunt", options.FullHunt...)
	appendQuery(options, "fofa-spider", options.FoFaSpider...)
	appendQuery(options, "google-spider", options.GoogleSpider...)
	appendQuery(options, "bing-spider", options.BingSpider...)
	appendQuery(options, "chinaz-spider", options.ChinazSpider...)
	appendQuery(options, "ip138-spider", options.Ip138Spider...)
	appendQuery(options, "rapiddns-spider", options.RapidDNSSpider...)
	appendQuery(options, "qianxun-spider", options.QianXunSpider...)
	appendQuery(options, "sitedossier-spider", options.SiteDossierSpider...)
	appendQuery(options, "anubis-spider", options.AnubisSpider...)
	appendQuery(options, "baidu-spider", options.BaiduSpider...)
	appendQuery(options, "yahoo-spider", options.YahooSpider...)
}
