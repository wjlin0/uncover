package core

import (
	"context"
	"fmt"
	"github.com/projectdiscovery/gologger"
	stringsutil "github.com/projectdiscovery/utils/strings"
	"github.com/wjlin0/pathScan/pkg/writer"
	"github.com/wjlin0/uncover"
	"github.com/wjlin0/uncover/sources"
	"os"
	"strings"
)

func GetTarget(limit int, field string, csv bool, output string, engine, query []string, proxy, auth string, location string) (<-chan string, error) {
	var (
		err          error
		outputWriter *writer.OutputWriter
	)
	opts := uncover.Options{
		Agents:                 engine,
		Queries:                query,
		Limit:                  limit,
		MaxRetry:               2,
		Timeout:                20,
		Proxy:                  proxy,
		ProxyAuth:              auth,
		ProviderConfigLocation: location,
	}

	outputWriter, err = writer.NewOutputWriter()
	if err != nil {
		return nil, err
	}
	if output != "" {
		switch {
		case csv:
			csvWriter, err := writer.NewCSVWriter(output, sources.Result{})
			if err != nil {
				return nil, err
			}
			outputWriter.AddWriters(csvWriter)
		default:
			create, err := os.Create(output)
			if err != nil {
				return nil, err
			}
			outputWriter.AddWriters(create)
		}
	}
	u, err := uncover.New(&opts)
	if err != nil {
		return nil, err
	}
	ret := make(chan string)

	ch, err := u.Execute(context.Background())
	go func(ch <-chan sources.Result) {
		defer close(ret)
		for result := range ch {
			switch {
			case result.Error != nil:
				gologger.Warning().Msgf("Request %s sending error: %s", result.Source, result.Error)
			case csv:
				toString, err := writer.CSVToString(result)
				if err != nil {
					continue
				}
				outputWriter.Write(toString)

				replacer := strings.NewReplacer("ip", result.IP, "host", result.Host,
					"port", fmt.Sprint(result.Port),
				)
				port := fmt.Sprintf("%d", result.Port)
				if (result.IP == "" || port == "0") && stringsutil.ContainsAny(field, "ip", "port") {
					field = "host"
				}
				outData := replacer.Replace(field)
				ret <- outData
			default:
				replacer := strings.NewReplacer("ip", result.IP, "host", result.Host,
					"port", fmt.Sprint(result.Port),
				)
				port := fmt.Sprintf("%d", result.Port)
				if (result.IP == "" || port == "0") && stringsutil.ContainsAny(field, "ip", "port") {
					field = "host"
				}
				outData := replacer.Replace(field)
				ret <- outData
				outputWriter.WriteString(outData)

			}

		}
	}(ch)

	return ret, err
}
