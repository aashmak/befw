package main

import (
	"befw/internal/ipfirewall"
	"context"
	"encoding/json"
	"fmt"
)

type AddRuleCommand struct {
	Table        string `long:"table" description:"set table" default:"filter"`
	Chain        string `long:"chain" description:"set chain"`
	RuleNumber   int    `long:"rulenum" description:"set rulenum"`
	InInterface  string `long:"in-interface" description:"set in-interface"`
	OutInterface string `long:"out-interface" description:"set out-interface"`
	SrcAddress   string `long:"src-address" description:"set src-address"`
	DstAddress   string `long:"dst-address" description:"set dst-address"`
	Protocol     string `long:"protocol" description:"set protocol"`
	SrcPort      string `long:"src-port" description:"set src-port"`
	DstPort      string `long:"dst-port" description:"set dst-port"`
	Action       string `long:"action" description:"set action"`
	Jump         string `long:"jump" description:"set jump"`
	Comment      string `long:"comment" description:"set comment"`
}

func (x *AddRuleCommand) Execute(args []string) error {
	var ipfw ipfirewall.IPFirewall

	ipfw.Tenant = options.Tenant
	ipfw.Rules = append(ipfw.Rules, ipfirewall.Rule{
		Table:        x.Table,
		Chain:        x.Chain,
		RuleNumber:   x.RuleNumber,
		InInterface:  x.InInterface,
		OutInterface: x.OutInterface,
		SrcAddress:   x.SrcAddress,
		DstAddress:   x.DstAddress,
		Protocol:     x.Protocol,
		SrcPort:      x.SrcPort,
		DstPort:      x.DstPort,
		Action:       x.Action,
		Jump:         x.Jump,
		Comment:      x.Comment,
	})

	if ipfwJSON, err := json.Marshal(ipfw); err == nil {
		resultJSON, err := SendRequest(context.Background(), fmt.Sprintf("%s/rule/add", options.Address), ipfwJSON)
		if err == nil {
			fmt.Printf("%s", resultJSON)
			fmt.Printf("rule added successful\n")
		} else {
			fmt.Printf("add rule failed\n")
		}
	}

	return nil
}

type DeleteRuleCommand struct {
	ID      string `long:"id" description:"set ruleID"`
	Table   string `long:"table" description:"set table" default:"filter"`
	Chain   string `long:"chain" description:"set chain"`
	Rulenum int    `long:"rulenum" description:"set rulenum"`
}

func (x *DeleteRuleCommand) Execute(args []string) error {
	var ipfw ipfirewall.IPFirewall

	ipfw.Tenant = options.Tenant
	ipfw.Rules = append(ipfw.Rules, ipfirewall.Rule{
		ID:         x.ID,
		Table:      x.Table,
		Chain:      x.Chain,
		RuleNumber: x.Rulenum,
	})

	if ipfwJSON, err := json.Marshal(ipfw); err == nil {
		resultJSON, err := SendRequest(context.Background(), fmt.Sprintf("%s/rule/delete", options.Address), ipfwJSON)
		if err == nil {
			fmt.Printf("%s", resultJSON)
			fmt.Printf("rule deleted successful\n")
		} else {
			fmt.Printf("delete rule failed\n")
		}
	}

	return nil
}

type ShowRuleCommand struct {
	ID      string `long:"id" description:"set ruleID"`
	Table   string `long:"table" description:"set table" default:"filter"`
	Chain   string `long:"chain" description:"set chain"`
	Rulenum int    `long:"rulenum" description:"set rulenum"`
}

func (x *ShowRuleCommand) Execute(args []string) error {
	var ipfw ipfirewall.IPFirewall

	ipfw.Tenant = options.Tenant
	ipfw.Rules = append(ipfw.Rules, ipfirewall.Rule{
		Table:      x.Table,
		Chain:      x.Chain,
		RuleNumber: x.Rulenum,
	})

	if ipfwJSON, err := json.Marshal(ipfw); err == nil {
		resultJSON, err := SendRequest(context.Background(), fmt.Sprintf("%s/rule", options.Address), ipfwJSON)
		if err == nil {
			fmt.Printf("%s", resultJSON)
		} else {
			fmt.Printf("rule list is failed\n")
		}
	}

	return nil
}

var addRuleCmd AddRuleCommand
var deleteRuleCmd DeleteRuleCommand
var showRuleCmd ShowRuleCommand

func init() {
	parser.AddCommand("rule-show",
		"Show rule",
		"",
		&showRuleCmd)

	parser.AddCommand("rule-add",
		"Add rule",
		"",
		&addRuleCmd)

	parser.AddCommand("rule-del",
		"Delete rule",
		"",
		&deleteRuleCmd)
}
