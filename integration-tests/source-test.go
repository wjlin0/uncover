package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	folderutil "github.com/projectdiscovery/utils/folder"
	"github.com/wjlin0/uncover/testutils"
)

var (
	ConfigFile = filepath.Join(folderutil.AppConfigDirOrDefault(".uncover-config", "uncover"), "provider-config.yaml")
)

type qianxunSpiderTestcases struct {
}

func (q qianxunSpiderTestcases) Execute() error {
	results, err := testutils.RunUncoverAndGetResults(debug, "-qianxun-spider", "baidu.com")
	if err != nil {
		return err
	}
	return expectResultsGreaterThanCount(results, 0)
}

type rapiddnsSpiderTestcases struct {
}

func (r rapiddnsSpiderTestcases) Execute() error {
	results, err := testutils.RunUncoverAndGetResults(debug, "-rapiddns-spider", "baidu.com")
	if err != nil {
		return err
	}
	return expectResultsGreaterThanCount(results, 0)
}

type baiduSpiderTestcases struct {
}

func (b baiduSpiderTestcases) Execute() error {
	results, err := testutils.RunUncoverAndGetResults(debug, "-baidu-spider", "baidu.com")
	if err != nil {
		return err
	}
	return expectResultsGreaterThanCount(results, 0)
}

type shodanidbTestcases struct {
}

func (s shodanidbTestcases) Execute() error {
	results, err := testutils.RunUncoverAndGetResults(debug, "-shodan-idb", "baidu.com")
	if err != nil {
		return err
	}
	return expectResultsGreaterThanCount(results, 0)
}

type yahooSpiderTestcases struct {
}

func (y yahooSpiderTestcases) Execute() error {
	results, err := testutils.RunUncoverAndGetResults(debug, "-yahoo-spider", "baidu.com")
	if err != nil {
		return err
	}
	return expectResultsGreaterThanCount(results, 0)
}

type sitedossierSpiderTestcases struct {
}

func (s sitedossierSpiderTestcases) Execute() error {
	results, err := testutils.RunUncoverAndGetResults(debug, "-sitedossier-spider", "baidu.com")
	if err != nil {
		return err
	}
	return expectResultsGreaterThanCount(results, 0)
}

type bingSpiderTestcases struct {
}

func (b bingSpiderTestcases) Execute() error {
	results, err := testutils.RunUncoverAndGetResults(debug, "-bing-spider", "baidu.com")
	if err != nil {
		return err
	}
	return expectResultsGreaterThanCount(results, 0)
}

type chinazSpiderTestcases struct {
}

func (c chinazSpiderTestcases) Execute() error {
	results, err := testutils.RunUncoverAndGetResults(debug, "-chinaz-spider", "baidu.com")
	if err != nil {
		return err
	}
	return expectResultsGreaterThanCount(results, 0)
}

type googleSpiderTestcases struct {
}

func (g googleSpiderTestcases) Execute() error {
	results, err := testutils.RunUncoverAndGetResults(debug, "-google-spider", "baidu.com")
	if err != nil {
		return err
	}
	return expectResultsGreaterThanCount(results, 0)
}

type ip138SpiderTestcases struct {
}

func (i ip138SpiderTestcases) Execute() error {
	results, err := testutils.RunUncoverAndGetResults(debug, "-ip138-spider", "baidu.com")
	if err != nil {
		return err
	}
	return expectResultsGreaterThanCount(results, 0)
}

type anubisSpiderTestcases struct {
}

func (a anubisSpiderTestcases) Execute() error {
	//TODO implement me
	results, err := testutils.RunUncoverAndGetResults(debug, "-anubis-spider", "baidu.com")
	if err != nil {
		return err
	}
	return expectResultsGreaterThanCount(results, 0)
}

type publicwwwTestcases struct {
}

func (p publicwwwTestcases) Execute() error {
	token := os.Getenv("PUBLICWWW_API_KEY")
	if token == "" {
		return errors.New("missing publicwww api key")
	}
	publicwwwToken := fmt.Sprintf(`publicwww: [%s]`, token)
	_ = os.WriteFile(ConfigFile, []byte(publicwwwToken), os.ModePerm)
	defer os.RemoveAll(ConfigFile)
	results, err := testutils.RunUncoverAndGetResults(debug, "-publicwww", "'Grafana'")
	if err != nil {
		return err
	}
	return expectResultsGreaterThanCount(results, 0)
}

type binaryTestcases struct {
}

func (b binaryTestcases) Execute() error {
	token := os.Getenv("BINARYEDGE_API_KEY")
	if token == "" {
		return errors.New("missing binaryedge api key")
	}
	binaryToken := fmt.Sprintf(`binaryedge: [%s]`, token)
	_ = os.WriteFile(ConfigFile, []byte(binaryToken), os.ModePerm)
	defer os.RemoveAll(ConfigFile)
	//TODO implement me
	results, err := testutils.RunUncoverAndGetResults(debug, "-binaryedge", "baidu.com")
	if err != nil {
		return err
	}
	return expectResultsGreaterThanCount(results, 0)
}

type githubTestcases struct {
}

func (g githubTestcases) Execute() error {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return errors.New("missing github api key")
	}
	githubToken := fmt.Sprintf(`github: [%s]`, token)
	_ = os.WriteFile(ConfigFile, []byte(githubToken), os.ModePerm)
	defer os.RemoveAll(ConfigFile)
	results, err := testutils.RunUncoverAndGetResults(debug, "-github", "baidu.com")
	if err != nil {
		return err
	}
	return expectResultsGreaterThanCount(results, 0)
}

type fofaSpiderTestcases struct{}

func (f fofaSpiderTestcases) Execute() error {
	//TODO implement me
	results, err := testutils.RunUncoverAndGetResults(debug, "-fofa-spider", "domain=\"baidu.com\"")
	if err != nil {
		return err
	}
	return expectResultsGreaterThanCount(results, 0)
}

type censysTestcases struct{}

func (h censysTestcases) Execute() error {
	id := os.Getenv("CENSYS_API_ID")
	if id == "" {
		return errors.New("missing censys api id")
	}
	secret := os.Getenv("CENSYS_API_SECRET")
	if secret == "" {
		return errors.New("missing censys api secret")
	}
	censysToken := fmt.Sprintf(`censys: [%s:%S]`, id, secret)
	_ = os.WriteFile(ConfigFile, []byte(censysToken), os.ModePerm)
	defer os.RemoveAll(ConfigFile)
	results, err := testutils.RunUncoverAndGetResults(debug, "-censys", "'services.software.vendor=Grafana'")
	if err != nil {
		return err
	}
	return expectResultsGreaterThanCount(results, 0)
}

type shodanTestcases struct{}

func (h shodanTestcases) Execute() error {
	token := os.Getenv("SHODAN_API_KEY")
	if token == "" {
		return errors.New("missing shodan api key")
	}
	shodanToken := fmt.Sprintf(`shodan: [%s]`, token)
	_ = os.WriteFile(ConfigFile, []byte(shodanToken), os.ModePerm)
	defer os.RemoveAll(ConfigFile)
	results, err := testutils.RunUncoverAndGetResults(debug, "-shodan", "'title:\"Grafana\"'")
	if err != nil {
		return err
	}
	err = expectResultsGreaterThanCount(results, 0)
	if err != nil {
		return err
	}
	results, err = testutils.RunUncoverAndGetResults(debug, "-shodan", "'org:\"Something, Inc.\"'")
	if err != nil {
		return err
	}
	return expectResultsGreaterThanCount(results, 1)
}

type zoomeyeTestcases struct{}

func (h zoomeyeTestcases) Execute() error {
	token := os.Getenv("ZOOMEYE_API_KEY")
	if token == "" {
		return errors.New("missing zoomeye api key")
	}
	zoomeyeToken := fmt.Sprintf(`zoomeye: [%s]`, token)

	_ = os.WriteFile(ConfigFile, []byte(zoomeyeToken), os.ModePerm)
	defer os.RemoveAll(ConfigFile)
	results, err := testutils.RunUncoverAndGetResults(debug, "-zoomeye", "'app:\"Atlassian JIRA\"'")
	if err != nil {
		return err
	}
	return expectResultsGreaterThanCount(results, 0)
}

type fofaTestcases struct{}

func (h fofaTestcases) Execute() error {
	token := os.Getenv("FOFA_KEY")
	if token == "" {
		return errors.New("missing fofa api key")
	}
	email := os.Getenv("FOFA_EMAIL")
	if email == "" {
		return errors.New("missing fofa email")
	}
	fofaToken := fmt.Sprintf(`fofa: [%s:%s]`, token, email)
	_ = os.WriteFile(ConfigFile, []byte(fofaToken), os.ModePerm)
	defer os.RemoveAll(ConfigFile)
	results, err := testutils.RunUncoverAndGetResults(debug, "-fofa", "'app=Grafana'")
	if err != nil {
		return err
	}
	return expectResultsGreaterThanCount(results, 0)
}

type hunterTestcases struct{}

func (h hunterTestcases) Execute() error {
	results, err := testutils.RunUncoverAndGetResults(debug, "-hunter", "'Grafana'")
	if err != nil {
		return err
	}
	return expectResultsGreaterThanCount(results, 0)
}

type quakeTestcases struct{}

func (h quakeTestcases) Execute() error {
	token := os.Getenv("QUAKE_TOKEN")
	if token == "" {
		return errors.New("missing quake api key")
	}
	quakeToken := fmt.Sprintf(`quake: [%s]`, token)
	_ = os.WriteFile(ConfigFile, []byte(quakeToken), os.ModePerm)
	defer os.RemoveAll(ConfigFile)
	results, err := testutils.RunUncoverAndGetResults(debug, "-quake", "'Grafana'")
	if err != nil {
		return err
	}
	return expectResultsGreaterThanCount(results, 0)
}

type netlasTestcases struct{}

func (h netlasTestcases) Execute() error {
	token := os.Getenv("NETLAS_API_KEY")
	if token == "" {
		return errors.New("missing netlas api key")
	}
	netlasToken := fmt.Sprintf(`netlas: [%s]`, token)
	_ = os.WriteFile(ConfigFile, []byte(netlasToken), os.ModePerm)
	defer os.RemoveAll(ConfigFile)
	results, err := testutils.RunUncoverAndGetResults(debug, "-netlas", "'Grafana'")
	if err != nil {
		return err
	}
	return expectResultsGreaterThanCount(results, 0)
}

type criminalipTestcases struct{}

func (h criminalipTestcases) Execute() error {
	token := os.Getenv("CRIMINALIP_API_KEY")
	if token == "" {
		return errors.New("missing criminalip api key")
	}
	criminalipToken := fmt.Sprintf(`criminalip: [%s]`, token)
	_ = os.WriteFile(ConfigFile, []byte(criminalipToken), os.ModePerm)
	defer os.RemoveAll(ConfigFile)
	results, err := testutils.RunUncoverAndGetResults(debug, "-criminalip", "'Grafana'")
	if err != nil {
		return err
	}
	return expectResultsGreaterThanCount(results, 0)
}

type fullhuntTestcases struct {
}

func (h fullhuntTestcases) Execute() error {
	token := os.Getenv("FULLHUNT_API_KEY")
	if token == "" {
		return errors.New("missing fullhunt api key")
	}
	fullhuntToken := fmt.Sprintf(`fullhunt: [%s]`, token)
	_ = os.WriteFile(ConfigFile, []byte(fullhuntToken), os.ModePerm)
	defer os.RemoveAll(ConfigFile)
	results, err := testutils.RunUncoverAndGetResults(debug, "-fullhunt", "baidu.com")
	if err != nil {
		return err
	}
	return expectResultsGreaterThanCount(results, 0)
}

type hunterhowTestcases struct{}

func (h hunterhowTestcases) Execute() error {
	token := os.Getenv("HUNTERHOW_API_KEY")
	if token == "" {
		return errors.New("missing hunterhow api key")
	}
	hunterhowApiKey := fmt.Sprintf(`hunterhow: [%s]`, token)
	_ = os.WriteFile(ConfigFile, []byte(hunterhowApiKey), os.ModePerm)
	defer os.RemoveAll(ConfigFile)
	results, err := testutils.RunUncoverAndGetResults(debug, "-hunterhow", "'web.body=\"ElasticJob\"'")
	if err != nil {
		return err
	}
	return expectResultsGreaterThanCount(results, 0)
}

type outputTestcases struct{}

func (h outputTestcases) Execute() error {
	results, err := testutils.RunUncoverAndGetResults(debug, "-q", "'element'", "-j", "-silent")
	if err != nil {
		return err
	}
	err = expectResultsGreaterThanCount(results, 0)
	if err != nil {
		return err
	}
	results, err = testutils.RunUncoverAndGetResults(debug, "-q", "'element'", "-r", "-silent")
	if err != nil {
		return err
	}
	return expectResultsGreaterThanCount(results, 0)
}
