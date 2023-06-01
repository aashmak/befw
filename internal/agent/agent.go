package agent

import (
	"befw/internal/ipfirewall"
	"befw/internal/logger"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type FirewallAgent struct {
	ModifyLock sync.RWMutex

	Tenant         string
	ServerURL      string
	PollInterval   int
	ReportInterval int

	IPFirewall ipfirewall.IPFirewall
}

func New() *FirewallAgent {
	return &FirewallAgent{
		ServerURL:      "http://127.0.0.1:8080/api/v1",
		PollInterval:   60,
		ReportInterval: 10,
	}
}

func (f *FirewallAgent) Watcher(ctx context.Context) {
	var interval = time.Duration(f.PollInterval) * time.Second

	client := &http.Client{
		Timeout: interval,
	}

	//find all rules for the tenant and table
	matchRule := fmt.Sprintf(`{
		"tenant": "%s",
		"rules": [
			{
				"table": "filter"
			}
		]
	}`, f.Tenant)

	for {
		select {
		case <-ctx.Done():
			return

		default:
			go func(ctx context.Context, client *http.Client, url string, body []byte) {
				var resultJSON []byte
				var err error

				if resultJSON, err = MakeRequest(ctx, client, url, body); err == nil {
					ipfwTmp := ipfirewall.IPFirewall{}

					if err = json.Unmarshal(resultJSON, &ipfwTmp); err == nil {
						f.ModifyLock.Lock()

						f.IPFirewall.Rules = ipfwTmp.Rules
						f.IPFirewall.AddressLists = ipfwTmp.AddressLists

						f.ModifyLock.Unlock()

						ipfirewall.ApplyRules(f.IPFirewall.Rules)
					}
				}

				if err != nil {
					logger.Error("", err)
				}
			}(ctx, client, f.ServerURL, []byte(matchRule))
		}

		<-time.After(interval)
	}
}

func (f *FirewallAgent) SendStat(ctx context.Context) {
	var interval = time.Duration(f.ReportInterval) * time.Second

	client := &http.Client{
		Timeout: interval,
	}

	requestQueue := make(chan []byte, 50)

	//Create worker pool
	for i := 0; i < 3; i++ {
		workerID := i + 1
		go func(workerID int, ctx context.Context, client *http.Client, url string, requestQueue <-chan []byte) {
			for req := range requestQueue {
				_, err := MakeRequest(ctx, client, url, req)
				if err != nil {
					logger.Error(fmt.Sprintf("[Worker #%d]", workerID), err)
				} else {
					logger.Debug(fmt.Sprintf("[Worker #%d] the stats sent successfully", workerID))
				}
			}
		}(workerID, ctx, client, fmt.Sprintf("%s/stat", f.ServerURL), requestQueue)
	}

	for {
		select {
		case <-ctx.Done():
			close(requestQueue)
			return

		default:
			chains := make(map[string]struct{})
			var ruleStats []ipfirewall.RuleStat

			f.ModifyLock.Lock()

			for _, rule := range f.IPFirewall.Rules {
				if _, ok := chains[rule.Chain]; !ok {
					chains[rule.Chain] = struct{}{}

					if stat, err := ipfirewall.Stat(rule.Table, rule.Chain); err == nil {
						ruleStats = append(ruleStats, stat...)
					}
				}
			}

			if len(f.IPFirewall.Rules) == len(ruleStats) {
				var ipfwTmp ipfirewall.IPFirewall
				ipfwTmp.Tenant = f.Tenant

				for i, rule := range f.IPFirewall.Rules {

					ruleStats[i].Packets = 1
					ruleStats[i].Bytes = 10

					ipfwTmp.Stats = append(
						ipfwTmp.Stats,
						ipfirewall.RuleStat{
							ID:      rule.ID,
							Packets: ruleStats[i].Packets,
							Bytes:   ruleStats[i].Bytes,
						})
				}

				if ipfwJSON, err := json.Marshal(ipfwTmp); err == nil {
					requestQueue <- ipfwJSON
				}
			} else {
				logger.Debug("invalid rules stats")
			}

			f.ModifyLock.Unlock()
		}

		<-time.After(interval)
	}
}

func MakeRequest(ctx context.Context, client *http.Client, url string, body []byte) ([]byte, error) {
	var b bytes.Buffer

	writer := gzip.NewWriter(&b)
	_, err := writer.Write(body)
	if err != nil {
		return nil, fmt.Errorf("failed init compress writer: %v", err.Error())
	}
	writer.Close()

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(b.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("new request error: %s", err.Error())
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Set("Content-Encoding", "gzip")
	request.Header.Set("Accept-Encoding", "gzip")

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("http request error: %s", err.Error())
	}
	defer response.Body.Close()

	if response.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(response.Body)
		if err != nil {
			return nil, fmt.Errorf("failed init compress reader: %s", err.Error())
		}
		defer reader.Close()

		response.Body = io.NopCloser(reader)
	}

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("the request was not executed successfully")
	}

	result, _ := io.ReadAll(response.Body)

	return result, nil
}
