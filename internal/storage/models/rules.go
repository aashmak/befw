package models

import "github.com/google/uuid"

type DBRule struct {
	ID        uuid.UUID `db:"id"`
	Tenant    string    `db:"tenant"`
	Ruletable string    `db:"ruletable"`
	Chain     string    `db:"chain"`
	Rulenum   int       `db:"rulenum"`
	Rulespec  string    `db:"rulespec"`
	Packets   uint64    `db:"packets"`
	Bytes     uint64    `db:"bytes"`
}
