package quake

type responseData struct {
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	Domain   string `json:"domain"`
}

type pagination struct {
	Count     int `json:"count"`
	PageIndex int `json:"page_index"`
	PageSize  int `json:"page_size"`
	Total     int `json:"total"`
}

type meta struct {
	Pagination pagination `json:"pagination"`
}

type Response struct {
	Data    []responseData `json:"data"`
	Message string         `json:"message"`
	Meta    meta           `json:"meta"`
}
