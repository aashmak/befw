package server

import (
	"befw/internal/ipfirewall"
	"befw/internal/logger"
	"befw/internal/storage/models"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (s HTTPServer) defaultHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusForbidden)
}

/* GET one and more Rules */
func (s HTTPServer) getRule(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		logger.Debug("content type is not application/json")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var ipfw ipfirewall.IPFirewall
	var err error
	var ruleJSON []byte

	if ruleJSON, err = io.ReadAll(r.Body); err == nil {
		err = json.Unmarshal(ruleJSON, &ipfw)
	}

	if err != nil {
		logger.Error("", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tenant := ipfw.Tenant
	if tenant == "" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var ResponseIpfw ipfirewall.IPFirewall
	var ResponseIpfwRules []ipfirewall.Rule
	var ResponseIpfwStats []ipfirewall.RuleStat

	for _, rule := range ipfw.Rules {
		table := rule.Table

		if table == "" {
			logger.Error("getRule Error: tenant or table is empty", err)
			w.WriteHeader(http.StatusForbidden)
			return
		}

		chain := rule.Chain
		rulenum := rule.RuleNumber

		var dbRules []models.DBRule
		dbRules, err = s.Storage.Rules().GetRule(tenant, table, chain, rulenum)
		if err == nil {
			for _, dbRule := range dbRules {
				var ipfwRule ipfirewall.Rule
				var ipfwStat ipfirewall.RuleStat

				json.Unmarshal([]byte(dbRule.Rulespec), &ipfwRule)

				ipfwRule.ID = dbRule.ID.String()
				ipfwRule.Table = dbRule.Ruletable
				ipfwRule.Chain = dbRule.Chain
				ipfwRule.RuleNumber = dbRule.Rulenum

				ipfwStat.ID = dbRule.ID.String()
				ipfwStat.Packets = dbRule.Packets
				ipfwStat.Bytes = dbRule.Bytes

				ResponseIpfwRules = append(ResponseIpfwRules, ipfwRule)
				ResponseIpfwStats = append(ResponseIpfwStats, ipfwStat)
			}
		}
	}

	ResponseIpfw.Rules = ResponseIpfwRules
	ResponseIpfw.Stats = ResponseIpfwStats

	ipfwJSON, err := json.Marshal(ResponseIpfw)
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write(ipfwJSON)
		return
	}

	logger.Debug("response status is Forbidden")
	w.WriteHeader(http.StatusForbidden)
}

/* ADD one and more Rule */
func (s HTTPServer) addRule(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		logger.Debug("content type is not application/json")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var ipfw ipfirewall.IPFirewall
	var err error
	var ruleJSON []byte

	if ruleJSON, err = io.ReadAll(r.Body); err == nil {
		err = json.Unmarshal(ruleJSON, &ipfw)
	}

	if err != nil {
		logger.Error("", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tenant := ipfw.Tenant
	if tenant == "" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	for _, rule := range ipfw.Rules {
		var table, chain string

		table = rule.Table
		chain = rule.Chain

		rule.Table = ""
		rule.Chain = ""
		rule.RuleNumber = 0

		rulespec, err := json.Marshal(rule)
		if err != nil {
			logger.Error("", err)
			break
		}

		if table == "" || chain == "" || string(rulespec) == "{}" {
			logger.Error("data is not be empty", nil)
			w.WriteHeader(http.StatusForbidden)
			return
		}
		logger.Debug(fmt.Sprintf("marshall succefull: %s", rulespec))

		err = s.Storage.Rules().AppendRule(tenant, table, chain, string(rulespec))
		if err != nil {
			logger.Error("", err)
			w.WriteHeader(http.StatusForbidden)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

/* DELETE one and more Rule */
func (s HTTPServer) deleteRule(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		logger.Debug("content type is not application/json")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	ruleJSON, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("server could not read request body", err)
	}
	logger.Debug(string(ruleJSON))

	var ipfw ipfirewall.IPFirewall
	if err := json.Unmarshal(ruleJSON, &ipfw); err != nil {
		logger.Error("", err)
		return
	}

	for _, rule := range ipfw.Rules {
		tenant := ipfw.Tenant
		ruleID := rule.ID
		table := rule.Table
		chain := rule.Chain
		rulenum := rule.RuleNumber

		if ruleID != "" {
			err = s.Storage.Rules().DeleteRuleByID(ruleID)
		} else {
			err = s.Storage.Rules().DeleteRule(tenant, table, chain, rulenum)
		}
		if err == nil {
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	logger.Debug("response status is Forbidden")
	w.WriteHeader(http.StatusForbidden)
}

/* UPDATE stat one and more Rule */
func (s HTTPServer) updateRuleStat(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		logger.Debug("content type is not application/json")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	ruleJSON, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("server could not read request body", err)
	}
	logger.Debug(string(ruleJSON))

	var ipfw ipfirewall.IPFirewall
	if err := json.Unmarshal(ruleJSON, &ipfw); err != nil {
		logger.Error("", err)
		return
	}

	for _, stat := range ipfw.Stats {
		ruleID := stat.ID
		packets := stat.Packets
		bytes := stat.Bytes

		err = s.Storage.Rules().UpdateRuleStat(ruleID, packets, bytes)
		if err != nil {
			logger.Error("", err)
		}
	}

	if err == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusForbidden)
}

func unzipBodyHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentEncodingValues := r.Header.Values("Content-Encoding")

		if contentEncodingContains(contentEncodingValues, "gzip") {
			reader, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(
					w,
					http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError,
				)
				return
			}
			defer reader.Close()

			r.Body = io.NopCloser(reader)
		}

		next.ServeHTTP(w, r)
	})
}
