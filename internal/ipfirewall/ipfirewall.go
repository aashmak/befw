package ipfirewall

import (
	"befw/internal/logger"
	"errors"
	"fmt"
	"strings"

	goiptables "github.com/coreos/go-iptables/iptables"
)

type IPFirewall struct {
	Tenant       string        `json:"tenant"`
	Rules        []Rule        `json:"rules,omitempty"`
	AddressLists []AddressList `json:"address-lists,omitempty"`
	Stats        []RuleStat    `json:"stats"`
}

type Rule struct {
	ID           string `json:"id,omitempty"`
	Table        string `json:"table,omitempty"`
	Chain        string `json:"chain,omitempty"`
	RuleNumber   int    `json:"rulenum,omitempty"`
	InInterface  string `json:"in-interface,omitempty"`
	OutInterface string `json:"out-interface,omitempty"`
	SrcAddress   string `json:"src-address,omitempty"`
	DstAddress   string `json:"dst-address,omitempty"`
	Protocol     string `json:"protocol,omitempty"`
	SrcPort      string `json:"src-port,omitempty"`
	DstPort      string `json:"dst-port,omitempty"`
	Action       string `json:"action,omitempty"`
	Jump         string `json:"jump,omitempty"`
	Comment      string `json:"comment,omitempty"`
}

type AddressList struct {
	Name string `json:"name"`
	Host string `json:"host"`
}

type RuleSpec struct {
	Spec []string
}

type RuleStat struct {
	ID      string `json:"id"`
	Packets uint64 `json:"pkts"`
	Bytes   uint64 `json:"bytes"`
}

func (r *RuleSpec) AddParam(param string, value string) {
	if value != "" {
		if value[0] == '!' {
			r.Spec = append(r.Spec, "!", param, value[1:])
		} else {
			r.Spec = append(r.Spec, param, value)
		}
	}
}

func (r *RuleSpec) AddComment(value string) {
	if value != "" {
		r.Spec = append(r.Spec, "-m", "comment", "--comment", value)
	}
}

func (r *Rule) Spec() []string {
	var ruleSpec RuleSpec

	ruleSpec.AddParam("-i", r.InInterface)
	ruleSpec.AddParam("-o", r.OutInterface)
	ruleSpec.AddParam("-s", r.SrcAddress)
	ruleSpec.AddParam("-d", r.DstAddress)
	ruleSpec.AddParam("-p", r.Protocol)
	ruleSpec.AddParam("--sport", r.SrcPort)
	ruleSpec.AddParam("--dport", r.DstPort)
	ruleSpec.AddParam("-j", r.Jump)
	ruleSpec.AddComment(r.Comment)

	return ruleSpec.Spec
}

func (r *Rule) String() string {
	return strings.Join(r.Spec(), " ")
}

func (r *Rule) AddRule() error {
	ipt, err := goiptables.NewWithProtocol(goiptables.ProtocolIPv4)
	if err != nil {
		return err
	}

	switch r.Action {
	// inserts rulespec to specified table/chain (in specified position)
	case "insert":
		if r.RuleNumber != 0 {
			ipt.Insert(r.Table, r.Chain, r.RuleNumber, r.Spec()...)
		} else {
			return errors.New("to use insert action ,you must need to provides rule_number")
		}
	default:
		// appends rulespec to specified table/chain
		return ipt.AppendUnique(r.Table, r.Chain, r.Spec()...)
	}
	return nil
}

func CreateChainIfNotExist(table, chain string) error {
	ipt, err := goiptables.NewWithProtocol(goiptables.ProtocolIPv4)
	if err != nil {
		return err
	}

	if ok, _ := ipt.ChainExists(table, chain); !ok {
		err := ipt.NewChain(table, chain)
		if err != nil {
			logger.Error("", err)
			return err
		}
	}

	return nil
}

func ClearChain(table, chain string) error {
	ipt, err := goiptables.NewWithProtocol(goiptables.ProtocolIPv4)
	if err != nil {
		return err
	}

	err = ipt.ClearChain(table, chain)
	if err != nil {
		return err
	}

	return nil
}

func ApplyRules(rules []Rule) error {
	var chains map[string]struct{}
	var err error

	//create chains
	chains = make(map[string]struct{})
	for _, rule := range rules {
		if _, ok := chains[rule.Chain]; !ok {
			chains[rule.Chain] = struct{}{}
			CreateChainIfNotExist(rule.Table, rule.Chain)
			logger.Debug(fmt.Sprintf("new chain %s added successful", rule.Chain))
		}
	}

	chains = nil
	chains = make(map[string]struct{})
	for _, rule := range rules {
		if _, ok := chains[rule.Chain]; !ok {
			ClearChain(rule.Table, rule.Chain)
			chains[rule.Chain] = struct{}{}
			logger.Debug(fmt.Sprintf("chain %s cleared successful", rule.Chain))
		}

		err = rule.AddRule()
		if err != nil {
			logger.Error("", err)
			return err
		}
		logger.Debug(fmt.Sprintf("rule id %s in chain %s added successful", rule.ID, rule.Chain))
	}

	logger.Info("rules updated successful")
	return nil
}

func Stat(table, chain string) ([]RuleStat, error) {
	var stats []goiptables.Stat
	var ruleStats []RuleStat

	ipt, err := goiptables.NewWithProtocol(goiptables.ProtocolIPv4)
	if err != nil {
		return nil, err
	}

	stats, err = ipt.StructuredStats(table, chain)
	if err != nil {
		return nil, err
	}

	for _, stat := range stats {
		ruleStats = append(ruleStats, RuleStat{Packets: stat.Packets, Bytes: stat.Bytes})
	}

	return ruleStats, nil
}
