package storage

import "befw/internal/storage/models"

type Storage interface {
	DB() DBRepository
	Rules() RulesRepository
}

type DBRepository interface {
	InitDB() error
	Clear() error
	Close() error
}

type RulesRepository interface {
	GetRule(tenant, table, chain string, rulenum int) ([]models.DBRule, error)
	AppendRule(tenant, table, chain, rulespec string) error
	DeleteRule(tenant, table, chain string, rulenum int) error
	DeleteRuleByID(id string) error
	UpdateRuleStat(id string, packets, bytes uint64) error
}
