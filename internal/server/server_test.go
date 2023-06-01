package server

import (
	"befw/internal/ipfirewall"
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5/middleware"
)

func httpRequest(ts *httptest.Server, method, path string, body []byte) (int, string) {
	req, _ := http.NewRequest(method, ts.URL+path, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := http.DefaultClient.Do(req)
	respBody, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	return resp.StatusCode, string(respBody)
}

func TestContentEncodingContains(t *testing.T) {
	contentEncodingValues := []string{"deflate", "gzip", "bzip"}

	if !contentEncodingContains(contentEncodingValues, "deflate") {
		t.Errorf("Error: content contain deflate")
	}

	if !contentEncodingContains(contentEncodingValues, "gzip") {
		t.Errorf("Error: content contain gzip")
	}

	if contentEncodingContains(contentEncodingValues, "fake") {
		t.Errorf("Error: content not contain fake")
	}
}

func NewTestServer(ctx context.Context, cfg *Config) *HTTPServer {
	s := NewServer(ctx, cfg)
	s.chiRouter.Use(middleware.Compress(5, "text/html", "application/json"))
	s.chiRouter.Use(unzipBodyHandler)
	s.chiRouter.Get("/", s.defaultHandler)
	s.chiRouter.Post("/", s.defaultHandler)
	s.chiRouter.Post("/api/v1/rule", s.getRule)
	s.chiRouter.Post("/api/v1/rule/add", s.addRule)
	s.chiRouter.Post("/api/v1/rule/delete", s.deleteRule)
	s.chiRouter.Post("/api/v1/rule/stat", s.updateRuleStat)

	return s
}

// run test with postgres db
func TestHTTPServerWithDB(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := DefaultConfig()
	cfg.DatabaseDSN = "postgresql://postgres:postgres@postgres:5432/praktikum"
	s := NewTestServer(ctx, cfg)
	s.Storage.DB().InitDB()
	s.Storage.DB().Clear()

	ts := httptest.NewServer(s.chiRouter)
	defer ts.Close()
	defer s.Storage.DB().Close()

	tests := []struct {
		name               string
		action             string
		requestBody        []byte
		responseStatusCode int
		responseBody       string
		ipfwTest           ipfirewall.IPFirewall
	}{
		{
			name:   "add empty rule #1",
			action: "add",
			requestBody: []byte(`{
				"tenant": "",
				"rules": []
			}`),
			responseStatusCode: http.StatusForbidden,
			responseBody:       "",
		},
		{
			name:   "add empty rule #2",
			action: "add",
			requestBody: []byte(`{
				"tenant": "",
				"rules": [{},{"jump": "ACCEPT"}]
			}`),
			responseStatusCode: http.StatusForbidden,
			responseBody:       "",
		},
		{
			name:   "add rule without tenant",
			action: "add",
			requestBody: []byte(`{
				"tenant": "",
				"rules": [{"jump": "ACCEPT"}]
			}`),
			responseStatusCode: http.StatusForbidden,
			responseBody:       "",
		},
		{
			name:   "add rule #1",
			action: "add",
			requestBody: []byte(`{
				"tenant": "host1",
				"rules": [
					{
						"table": "filter",
						"chain": "BEFW",
						"src-address": "192.168.1.0/24",
						"protocol": "tcp",
						"dst-port": "443",
						"jump": "ACCEPT"
					}
				]
			}`),
			responseStatusCode: http.StatusOK,
			responseBody:       "",
		},
		{
			name:   "add rule #2",
			action: "add",
			requestBody: []byte(`{
				"tenant": "host1",
				"rules": [
					{
						"table": "filter",
						"chain": "BEFW",
						"src-address": "192.168.2.0/24",
						"protocol": "tcp",
						"dst-port": "8080",
						"jump": "ACCEPT"
					}
				]
			}`),
			responseStatusCode: http.StatusOK,
			responseBody:       "",
		},
		{
			name:   "add rule #3",
			action: "add",
			requestBody: []byte(`{
				"tenant": "host1",
				"rules": [
					{
						"table": "filter",
						"chain": "BEFW",
						"dst-address": "192.168.3.2",
						"protocol": "udp",
						"dst-port": "53",
						"jump": "ACCEPT"
					}
				]
			}`),
			responseStatusCode: http.StatusOK,
			responseBody:       "",
		},
		{
			name:   "delete rule #1",
			action: "delete",
			requestBody: []byte(`{
				"tenant": "host1",
				"rules": [
					{
						"table": "filter",
						"chain": "BEFW",
						"rulenum": 1
					}
				]
			}`),
			responseStatusCode: http.StatusOK,
			responseBody:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.action == "add" {
				statusCode, body := httpRequest(ts, "POST", "/api/v1/rule/add", tt.requestBody)
				if statusCode != tt.responseStatusCode || body != tt.responseBody {
					t.Errorf("Error")
				}
			}

			if tt.action == "get" {
				statusCode, body := httpRequest(ts, "POST", "/api/v1/rule", tt.requestBody)
				if statusCode != tt.responseStatusCode || body != tt.responseBody {
					t.Errorf("Error")
				}
			}

			if tt.action == "delete" {
				statusCode, body := httpRequest(ts, "POST", "/api/v1/rule/delete", tt.requestBody)
				if statusCode != tt.responseStatusCode || body != tt.responseBody {
					t.Errorf("Error")
				}
			}
		})
	}
}
