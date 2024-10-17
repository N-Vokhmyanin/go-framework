package migorm

import (
	"context"
	"gorm.io/gorm"
)

type Context interface {
	context.Context
	DB() *gorm.DB
	Log() Logger
}

type migrationContext struct {
	context.Context

	db  *gorm.DB
	log Logger
}

func (c *migrationContext) DB() *gorm.DB {
	return c.db
}

func (c *migrationContext) Log() Logger {
	return c.log
}
