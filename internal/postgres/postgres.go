package postgres

import (
	"befw/internal/storage"
	"befw/internal/storage/models"
	"strings"

	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	db *pgxpool.Pool
}

func NewPostgresDB(db *pgxpool.Pool) *Postgres {
	return &Postgres{
		db: db,
	}
}

func (p *Postgres) DB() storage.DBRepository {
	return p
}

func (p *Postgres) Rules() storage.RulesRepository {
	return p
}

func (p *Postgres) InitDB() error {
	_, err := p.db.Exec(
		context.Background(),
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

		CREATE TABLE IF NOT EXISTS rules (
			id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
			tenant text not null,
			ruletable text not null,
			chain text not null,
			rulenum integer CHECK (rulenum > 0),
			rulespec text not null,
			packets integer,
			bytes integer
		);
	`)

	return err
}

func (p *Postgres) Clear() error {
	_, err := p.db.Exec(context.Background(), `TRUNCATE rules;`)

	return err
}

func (p *Postgres) DeleteRule(tenant, table, chain string, rulenum int) error {
	var ruleID uuid.UUID

	err := p.db.QueryRow(
		context.Background(),
		`SELECT id FROM rules WHERE tenant=$1 AND ruletable=$2 AND chain=$3 AND rulenum=$4 LIMIT 1;`,
		tenant,
		table,
		chain,
		rulenum).Scan(&ruleID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("rule not found")
		}

		return err
	}

	return p.DeleteRuleByID(ruleID.String())
}

//delete Rule by ID
func (p *Postgres) DeleteRuleByID(id string) error {
	ctx := context.Background()

	tx, err := p.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = p.db.Exec(ctx, `DELETE FROM rules WHERE id=$1`, id)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (p *Postgres) AppendRule(tenant, table, chain, rulespec string) error {
	ctx := context.Background()

	tx, err := p.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var rulenum int
	err = p.db.QueryRow(
		context.Background(),
		`SELECT coalesce(MAX(rulenum),0) FROM rules WHERE tenant=$1 AND ruletable=$2 AND chain=$3;`,
		tenant,
		table,
		chain).Scan(&rulenum)
	if err != nil {
		return err
	}

	queryStr := `INSERT INTO rules (tenant, ruletable, chain, rulenum, rulespec, packets, bytes) 
				VALUES ($1, $2, $3, $4, $5, 0, 0)`
	_, err = p.db.Exec(context.Background(), queryStr, tenant, table, chain, (rulenum + 1), rulespec)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (p *Postgres) GetRule(tenant, table, chain string, rulenum int) ([]models.DBRule, error) {
	var dbRules []models.DBRule

	queryRule := new(strings.Builder)
	queryRule.WriteString("SELECT id, tenant, ruletable, chain, rulenum, rulespec, packets, bytes FROM rules")

	if tenant != "" {
		queryRule.WriteString(fmt.Sprintf(" WHERE tenant='%s'", tenant))
	} else {
		return nil, fmt.Errorf("tenant is not be empty")
	}

	if table != "" {
		queryRule.WriteString(fmt.Sprintf(" AND ruletable='%s'", table))
	}

	if chain != "" {
		queryRule.WriteString(fmt.Sprintf(" AND chain='%s'", chain))
	}

	if rulenum > 0 {
		queryRule.WriteString(fmt.Sprintf(" AND rulenum='%d'", rulenum))
	}
	queryRule.WriteString(" ORDER BY chain, rulenum;")

	rows, err := p.db.Query(context.Background(), queryRule.String())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("rule not found")
		}
	}

	for rows.Next() {
		var dbRule models.DBRule
		err := rows.Scan(&dbRule.ID, &dbRule.Tenant, &dbRule.Ruletable, &dbRule.Chain, &dbRule.Rulenum, &dbRule.Rulespec, &dbRule.Packets, &dbRule.Bytes)
		if err != nil {
			return nil, err
		}

		dbRules = append(dbRules, dbRule)
	}

	return dbRules, err
}

func (p *Postgres) GetRuleByID(id string) (models.DBRule, error) {
	var dbRule models.DBRule

	err := p.db.QueryRow(
		context.Background(),
		`SELECT id, tenant, ruletable, chain, rulenum, rulespec, packets, bytes 
		FROM rules 
		WHERE id=$1 LIMIT 1;`,
		id).Scan(&dbRule.ID, &dbRule.Tenant, &dbRule.Ruletable, &dbRule.Chain, &dbRule.Rulenum, &dbRule.Rulespec, &dbRule.Packets, &dbRule.Bytes)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dbRule, fmt.Errorf("rule not found")
		}
		return dbRule, err
	}

	return dbRule, nil
}

func (p *Postgres) UpdateRuleStat(id string, packets, bytes uint64) error {
	dbRule, err := p.GetRuleByID(id)
	if err != nil {
		return err
	}

	packets = packets + dbRule.Packets
	bytes = bytes + dbRule.Bytes

	_, err = p.db.Exec(
		context.Background(),
		`UPDATE rules 
		SET packets=$1, bytes=$2
		WHERE id=$3;`,
		packets, bytes, id)

	if err != nil {
		return err
	}

	return nil
}

func (p *Postgres) Close() error {
	p.db.Close()

	return nil
}

func (p *Postgres) Ping() error {
	return p.db.Ping(context.Background())
}
