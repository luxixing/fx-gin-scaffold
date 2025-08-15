package domain

import (
	"sync"
)

var (
	tablePrefix string
	once        sync.Once
)

// SetTablePrefix sets the global table prefix for all models
func SetTablePrefix(prefix string) {
	once.Do(func() {
		tablePrefix = prefix
	})
}

// GetTablePrefix returns the current table prefix
func GetTablePrefix() string {
	return tablePrefix
}

// GetTableName returns the full table name with prefix
func GetTableName(tableName string) string {
	if tablePrefix == "" {
		return tableName
	}
	return tablePrefix + tableName
}