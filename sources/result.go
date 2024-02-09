package sources

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"net"
	"reflect"
	"strings"
)

type Result struct {
	Timestamp int64  `json:"timestamp" csv:"timestamp"`
	Source    string `json:"source" csv:"source"`
	IP        string `json:"ip" csv:"IP"`
	Port      int    `json:"port" csv:"port"`
	Host      string `json:"host" csv:"host"`
	Url       string `json:"url" csv:"url"`
	Raw       []byte `json:"-" csv:"-"`
	Error     error  `json:"-" csv:"-"`
}

func (result *Result) IpPort() string {
	return net.JoinHostPort(result.IP, fmt.Sprint(result.Port))
}

func (result *Result) HostPort() string {
	return net.JoinHostPort(result.Host, fmt.Sprint(result.Port))
}

func (result *Result) RawData() string {
	return string(result.Raw)
}

func (result *Result) JSON() string {
	data, _ := json.Marshal(result)
	return string(data)
}
func (result *Result) CSV() string {
	buffer := bytes.Buffer{}
	encoder := csv.NewWriter(&buffer)
	var fields []string
	vl := reflect.ValueOf(*result)
	ty := vl.Type()
	for i := 0; i < vl.NumField(); i++ {
		if ty.Field(i).Tag.Get("csv") != "-" {
			fields = append(fields, fmt.Sprint(vl.Field(i).Interface()))
		}

	}
	if err := encoder.Write(fields); err != nil {
		return ""
	}
	encoder.Flush()
	return strings.TrimSpace(buffer.String())
}

func (result *Result) CSVHeader() (string, error) {
	buffer := bytes.Buffer{}
	writer := csv.NewWriter(&buffer)
	ty := reflect.TypeOf(*result)
	var headers []string
	for i := 0; i < ty.NumField(); i++ {
		if ty.Field(i).Tag.Get("csv") != "-" {
			headers = append(headers, ty.Field(i).Tag.Get("csv"))
		}
	}

	if err := writer.Write(headers); err != nil {
		return "", errors.Wrap(err, "Could not write headers")
	}
	writer.Flush()
	return strings.TrimSpace(buffer.String()), nil
}
