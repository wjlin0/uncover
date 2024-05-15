package sources

type Keys struct {
	CensysToken     string
	CensysSecret    string
	Shodan          string
	FofaEmail       string
	FofaKey         string
	QuakeToken      string
	HunterToken     string
	ZoomEyeToken    string
	NetlasToken     string
	CriminalIPToken string
	PublicwwwToken  string
	HunterHowToken  string
	BinaryedgeToken string
	GithubToken     string
	FullHuntToken   string
	Zone0Token      string
	DayDayMapToken  string
}

func (keys Keys) Empty() bool {
	return keys.CensysSecret == "" &&
		keys.CensysToken == "" &&
		keys.Shodan == "" &&
		keys.FofaEmail == "" &&
		keys.FofaKey == "" &&
		keys.QuakeToken == "" &&
		keys.HunterToken == "" &&
		keys.ZoomEyeToken == "" &&
		keys.NetlasToken == "" &&
		keys.CriminalIPToken == "" &&
		keys.PublicwwwToken == "" &&
		keys.HunterHowToken == "" &&
		keys.BinaryedgeToken == "" &&
		keys.GithubToken == "" &&
		keys.FullHuntToken == "" &&
		keys.Zone0Token == "" &&
		keys.DayDayMapToken == ""
}
