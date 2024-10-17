package migorm

import (
	"gorm.io/gorm"
	"time"
)

const DefaultMigrationsTableName = "migrations"

type migrationModel struct {
	Id        uint   `gorm:"primary_key;"`
	Name      string `gorm:"unique;not null"`
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

func (m migrationModel) TableName() string {
	return DefaultMigrationsTableName
}

func WithCustomMigrationsTableName(customTableName string) func(tx *gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Table(customTableName)
	}
}
