package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/jessevdk/go-flags"
)

type Options struct {
	Address string `long:"address" short:"a" env:"ADDRESS" default:"http://127.0.0.1:8080/api/v1" description:"set address of firewall server"`
	Tenant  string `long:"tenant" short:"t" env:"TENANT" description:"set current tenant"`
}

var options Options
var parser = flags.NewParser(&options, flags.HelpFlag)

func main() {
	if _, err := parser.Parse(); err != nil {
		var e *flags.Error

		if errors.As(err, &e) {
			if e.Type == flags.ErrHelp {
				log.Printf("%s", e.Message)
				os.Exit(0)
			}
		}
		log.Fatalf("error parse arguments:%+v\n", err)
	}
}

func SendRequest(ctx context.Context, url string, body []byte) ([]byte, error) {
	client := &http.Client{}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("new request error: %s", err.Error())
	}

	request.Header.Add("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("http request error: %s", err.Error())
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("the request was not executed successfully")
	}

	result, _ := io.ReadAll(response.Body)

	return result, nil
}
