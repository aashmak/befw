package ipfirewall

import (
	"testing"
)

func TestRuleString(t *testing.T) {
	rule := Rule{
		Table:        "filter",
		Chain:        "BEFW_TEST",
		InInterface:  "eth0",
		OutInterface: "eth0",
		SrcAddress:   "192.168.1.1",
		DstAddress:   "192.168.1.100",
		Protocol:     "udp",
		SrcPort:      "10001",
		DstPort:      "53",
		Jump:         "DROP",
		Comment:      "test rule",
	}

	testRuleString := "-i eth0 -o eth0 -s 192.168.1.1 -d 192.168.1.100 -p udp --sport 10001 --dport 53 -j DROP -m comment --comment test rule"
	if rule.String() != testRuleString {
		t.Errorf("Error")
	}
}
