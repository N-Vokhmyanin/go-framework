package database

import (
	"encoding/json"
	"fmt"
	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
	"strings"
)

type clickhouseDialector struct {
	clickhouse.Dialector
}

var _ gorm.Dialector = (*clickhouseDialector)(nil)

func NewClickhouseDialector(dsn string) gorm.Dialector {
	return &clickhouseDialector{
		Dialector: clickhouse.Dialector{
			Config: &clickhouse.Config{
				DSN:                      dsn,
				DisableDatetimePrecision: true,
			},
		},
	}
}

func (d clickhouseDialector) DataTypeOf(field *schema.Field) string {
	switch field.DataType {
	case schema.Time:
		precision := ""
		if !d.Dialector.DisableDatetimePrecision {
			if field.Precision == 0 {
				field.Precision = 3
			}
			if field.Precision > 0 {
				precision = fmt.Sprintf("(%d)", field.Precision)
			}
		}
		return "DateTime" + precision
	}

	return d.Dialector.DataTypeOf(field)
}

func (d clickhouseDialector) Migrator(db *gorm.DB) gorm.Migrator {
	return clickhouse.Migrator{
		Migrator: migrator.Migrator{
			Config: migrator.Config{
				DB:        db,
				Dialector: &d,
			},
		},
		Dialector: d.Dialector,
	}
}

func (d clickhouseDialector) Explain(sql string, vars ...interface{}) string {
	jVars := make([]interface{}, len(vars))
	for i, v := range vars {
		jv, _ := json.Marshal(v)
		if string(jv) == "" {
			jv = []byte("?")
		}
		jVars[i] = jv
	}
	return fmt.Sprintf(strings.ReplaceAll(sql, "?", "%s"), jVars...)
}
