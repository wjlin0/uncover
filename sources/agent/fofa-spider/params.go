package fofa_spider

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"sort"
)

const (
	private = "-----BEGIN RSA PRIVATE KEY-----\r\nMIIEogIBAAKCAQEAv0xjefuBTF6Ox940ZqLLUFFBDtTcB9dAfDjWgyZ2A55K+VdG\r\nc1L5LqJWuyRkhYGFTlI4K5hRiExvjXuwIEed1norp5cKdeTLJwmvPyFgaEh7Ow19\r\nTu9sTR5hHxThjT8ieArB2kNAdp8Xoo7O8KihmBmtbJ1umRv2XxG+mm2ByPZFlTdW\r\nRFU38oCPkGKlrl/RzOJKRYMv10s1MWBPY6oYkRiOX/EsAUVae6zKRqNR2Q4HzJV8\r\ngOYMPvqkau8hwN8i6r0z0jkDGCRJSW9djWk3Byi3R2oSdZ0IoS+91MFtKvWYdnNH\r\n2Ubhlnu1P+wbeuIFdp2u7ZQOtgPX0mtQ263e5QIDAQABAoIBAD67GwfeTMkxXNr3\r\n5/EcQ1XEP3RQoxLDKHdT4CxDyYFoQCfB0e1xcRs0ywI1be1FyuQjHB5Xpazve8lG\r\nnTwIoB68E2KyqhB9BY14pIosNMQduKNlygi/hKFJbAnYPBqocHIy/NzJHvOHOiXp\r\ndL0AX3VUPkWW3rTAsar9U6aqcFvorMJQ2NPjijcXA0p1MlZAZKODO2wqidfQ487h\r\nxy0ZkriYVi419j83a1cCK0QocXiUUeQM6zRNgQv7LCmrFo2X4JEzlujEveqvsDC4\r\nMBRgkK2lNH+AFuRwOEr4PIlk9rrpHA4O1V13P3hJpH5gxs5oLLM1CWWG9YWLL44G\r\nzD9Tm8ECgYEA8NStMXyAmHLYmd2h0u5jpNGbegf96z9s/RnCVbNHmIqh/pbXizcv\r\nmMeLR7a0BLs9eiCpjNf9hob/JCJTms6SmqJ5NyRMJtZghF6YJuCSO1MTxkI/6RUw\r\nmrygQTiF8RyVUlEoNJyhZCVWqCYjctAisEDaBRnUTpNn0mLvEXgf1pUCgYEAy1kE\r\nd0YqGh/z4c/D09crQMrR/lvTOD+LRMf9lH+SkScT0GzdNIT5yuscRwKsnE6SpC5G\r\nySJFVhCnCBsQqq+ohsrXt8a99G7ePTMSAGK3QtC7QS3liDmvPBk6mJiLrKiRAZos\r\nvgPg7nTP8VuF0ZIKzkdWbGoMyNxVFZXovQ8BYxECgYBvCR9xGX4Qy6KiDlV18wNu\r\nElYkxVqFBBE0AJRg/u+bnQ9jWhi2zxLa1eWZgtss80c876I8lbkGNWedOVZioatm\r\nMFLC4bFalqyZWyO7iP7i60LKvfDJfkOSlDUu3OikahFOiqyG1VBz4+M4U500alIU\r\nAVKD14zTTZMopQSkgUXsoQKBgHd8RgiD3Qde0SJVv97BZzP6OWw5rqI1jHMNBK72\r\nSzwpdxYYcd6DaHfYsNP0+VIbRUVdv9A95/oLbOpxZNi2wNL7a8gb6tAvOT1Cvggl\r\n+UM0fWNuQZpLMvGgbXLu59u7bQFBA5tfkhLr5qgOvFIJe3n8JwcrRXndJc26OXil\r\n0Y3RAoGAJOqYN2CD4vOs6CHdnQvyn7ICc41ila/H49fjsiJ70RUD1aD8nYuosOnj\r\nwbG6+eWekyLZ1RVEw3eRF+aMOEFNaK6xKjXGMhuWj3A9xVw9Fauv8a2KBU42Vmcd\r\nt4HRyaBPCQQsIoErdChZj8g7DdxWheuiKoN4gbfK4W1APCcuhUA=\r\n-----END RSA PRIVATE KEY-----"
	appId   = "9e9fb94330d97833acfbc041ee1a76793f1bc691"
	stats   = "https://api.fofa.info/v1/search/stats?qbase64=%s&full=%v&fields=%s&ts=%v&sign=%s&app_id=%s"
	URL     = "https://fofa.info/result?qbase64=%s&page=%d&page_size=%d"
)

type fofaRequest struct {
	Query   string `json:"query"`
	Page    int    `json:"page"`
	PageNum int    `json:"page_num"`
}

var rsaPrivateKey *rsa.PrivateKey

func init() {
	rsaPrivateKey, _ = parsePKCS1PemByPrivateKey([]byte(private))
}

type foFaStatsResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    data   `json:"data"`
}

type data struct {
	Size      int       `json:"size"`
	Page      int       `json:"page"`
	Countries []country `json:"countries"`
}

type country struct {
	Code    string   `json:"code"`
	Name    string   `json:"name"`
	Count   int      `json:"count"`
	Regions []region `json:"regions"`
}
type region struct {
	Code  string `json:"code"`
	Count int    `json:"count"`
	Name  string `json:"name"`
}

func serialize(h map[string]any) string {
	var keys []string
	for k, _ := range h {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	s := ""
	for _, k := range keys {
		switch h[k].(type) {
		case string:
			if h[k] == "" {
				continue
			}
		}
		s += fmt.Sprintf("%s%v", k, h[k])
	}
	return s
}

func hashWithSha256(b []byte) (hashed []byte) {
	myHash := sha256.New()
	myHash.Write(b)
	return myHash.Sum(nil)
}
func parsePKCS1PemByPrivateKey(b []byte) (*rsa.PrivateKey, error) {
	p, _ := pem.Decode(b)
	if p == nil {
		return nil, errors.New("pem格式错误")
	}
	private, err := x509.ParsePKCS1PrivateKey(p.Bytes)
	if err != nil {
		return nil, err
	}
	return private, nil
}

func signQuery(str string) (string, error) {
	hashed := hashWithSha256([]byte(str))
	v15, err := signPKCS1v15(rsaPrivateKey, crypto.SHA256, hashed)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(v15), nil
}
func signPKCS1v15(private *rsa.PrivateKey, hash crypto.Hash, hashed []byte) (sign []byte, err error) {
	return rsa.SignPKCS1v15(rand.Reader, private, hash, hashed)
}
