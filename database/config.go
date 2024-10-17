package database

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"strings"
	"time"
)

const (
	MysqlDriver      = "mysql"
	PostgresqlDriver = "postgresql"
	ClickhouseDriver = "clickhouse"
)

type Config struct {
	Name   string
	Driver string

	DBHost string
	DBPort string
	DBUser string
	DBPass string
	DBName string

	MigrationsTableName  string
	MigrationsDir        string
	MigrationsRunOnStart bool

	LogLevel      string
	SlowThreshold time.Duration
}

func (c *Config) IsDefault() bool {
	if c == nil {
		return false
	}
	return c.Name == ""
}

func (c *Config) UpperPrefix() string {
	if c == nil {
		return ""
	}
	if c.IsDefault() {
		return ""
	}
	return strings.ToUpper(c.Name + "_")
}

func (c *Config) LowerPrefix() string {
	if c == nil {
		return ""
	}
	if c.IsDefault() {
		return ""
	}
	return strings.ToLower(c.Name + "_")
}

func (c *Config) Dialector() gorm.Dialector {
	if c == nil {
		return nil
	}
	switch c.Driver {
	case MysqlDriver, "":
		dsn := fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=true&loc=Local",
			c.DBUser,
			c.DBPass,
			c.DBHost,
			c.DBPort,
			c.DBName,
		)
		return mysql.New(mysql.Config{
			DSN: dsn,
		})
	case PostgresqlDriver:
		dsn := fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Europe/Moscow",
			c.DBHost,
			c.DBUser,
			c.DBPass,
			c.DBName,
			c.DBPort,
		)
		return postgres.New(postgres.Config{
			DSN: dsn,
		})
	case ClickhouseDriver:
		dsn := fmt.Sprintf(
			"clickhouse://%s:%s@%s:%s/%s",
			c.DBUser,
			c.DBPass,
			c.DBHost,
			c.DBPort,
			c.DBName,
		)
		return NewClickhouseDialector(dsn)
	}
	return nil
}

// ============================

type Configs []*Config

func (c Configs) Default() *Config {
	for _, cfg := range c {
		if cfg.IsDefault() {
			return cfg
		}
	}
	return nil
}
