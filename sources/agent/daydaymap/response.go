package daydaymap

// DaydayMapResponse contains the fofa response
type DaydayMapResponse struct {
	Code    int    `json:"code"`
	Data    Data   `json:"data"`
	Message string `json:"msg"`
}

type Data struct {
	List []Result `json:"list"`
}
type Result struct {
	IP      string `json:"ip"`
	Port    int    `json:"port"`
	Domain  string `json:"domain"`
	Service string `json:"service"`
}
