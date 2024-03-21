package zoomeye_spider

type aggsResponse struct {
	Status  int        `json:"status,omitempty"`
	Country []*country `json:"country,omitempty"`
}
type country struct {
	Name   string `json:"name,omitempty"`
	Count  int    `json:"count,omitempty"`
	Label  string `json:"label,omitempty"`
	NameZh string `json:"name_zh,omitempty"`
	// 省份信息
	Subdivisions []subdivisions `json:"subdivisions_zh,omitempty"`
}
type subdivisions struct {
	Name  string  `json:"name,omitempty"`
	Count int     `json:"count,omitempty"`
	City  []*city `json:"city_zh,omitempty"`
}
type city struct {
	Name  string `json:"name,omitempty"`
	Count int    `json:"count,omitempty"`
}

type response struct {
	Status  int        `json:"status,omitempty"`
	Matches []*matches `json:"matches,omitempty"`
	Total   int        `json:"total,omitempty"`
}
type matches struct {
	Site     string      `json:"site,omitempty"`
	Ip       interface{} `json:"ip,omitempty"`
	Ssl      string      `json:"ssl,omitempty"`
	Type     string      `json:"type,omitempty"`
	PortInfo *portInfo   `json:"portinfo,omitempty"`
}
type portInfo struct {
	Port    int    `json:"port,omitempty"`
	Service string `json:"service,omitempty"`
}
