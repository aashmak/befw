package postgres

import (
	"befw/internal/storage/models"
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresDB(t *testing.T) {
	var db *pgxpool.Pool
	var err error

	dsn := "postgresql://postgres:postgres@postgres:5432/praktikum"
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	db, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	pg := NewPostgresDB(db)
	pg.InitDB()
	pg.Clear()
	defer pg.DB().Close()

	// === Test Add Rule ===
	err = pg.AppendRule("host1", "filter", "BEFW", `{"src-address": "10.43.0.0/16", "jump": "ACCEPT"}`)
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	err = pg.AppendRule("host1", "filter", "BEFW", `{"src-address": "10.44.0.0/16", "jump": "ACCEPT"}`)
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	err = pg.AppendRule("host1", "filter", "BEFW", `{"src-address": "10.45.0.0/16", "jump": "ACCEPT"}`)
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	// === Test Get Rules ===
	var rules []models.DBRule
	rules, err = pg.GetRule("host1", "filter", "BEFW", 0)
	if err != nil || len(rules) != 3 {
		t.Errorf("Error: %s", err)
	}

	if rules[0].Tenant != "host1" ||
		rules[0].Ruletable != "filter" ||
		rules[0].Chain != "BEFW" ||
		rules[0].Rulespec != `{"src-address": "10.43.0.0/16", "jump": "ACCEPT"}` {
		t.Errorf("Error: %s", err)
	}

	// === Test Delete Rule by number ===
	id0 := rules[0].ID.String()
	err = pg.DeleteRule("host1", "filter", "BEFW", 1)
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	_, err = pg.GetRuleByID(id0)
	if err == nil {
		t.Errorf("Error: %s", err)
	}

	// === Test Delete Rule by ID ===
	id1 := rules[1].ID.String()
	err = pg.DeleteRuleByID(id1)
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	_, err = pg.GetRuleByID(id1)
	if err == nil {
		t.Errorf("Error: %s", err)
	}

	// === Test Update stat ===
	id2 := rules[2].ID.String()
	err = pg.UpdateRuleStat(id2, 12, 1212)
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	var rule2 models.DBRule
	rule2, err = pg.GetRuleByID(id2)
	if err != nil || rule2.Packets != 12 || rule2.Bytes != 1212 {
		t.Errorf("Error: %s", err)
	}

	pg.Clear()
}
