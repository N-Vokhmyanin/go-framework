package gormopts

import (
	"github.com/N-Vokhmyanin/go-framework/database/dboptions"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func WithOptions(db *gorm.DB, opts *dboptions.Options) *gorm.DB {
	if opts == nil {
		return db
	}
	if opts.LockingForUpdate {
		db = db.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if !opts.SavingFields.IsEmpty() {
		db = db.Scopes(SelectScope(opts.SavingFields.ToCamelStrings()))
	}
	if !opts.WithRelations.IsEmpty() {
		for _, relation := range opts.WithRelations {
			db = db.Preload(relation.ToCamel())
		}
	}
	return db
}

func SelectScope(fields []string) func(query *gorm.DB) *gorm.DB {
	return func(query *gorm.DB) *gorm.DB {
		var dbFields []string
		if query.Statement.Schema == nil && query.Statement.Dest != nil {
			query.Error = query.Statement.Parse(query.Statement.Dest)
			if query.Error != nil {
				return query
			}
		}
		if query.Statement.Schema != nil {
			for _, field := range fields {
				for _, parsedField := range query.Statement.Schema.Fields {
					if len(parsedField.BindNames[0]) > 0 && parsedField.BindNames[0] == field {
						if parsedField.DBName != "" {
							dbFields = append(dbFields, parsedField.DBName)
						} else if parsedField.Name != "" {
							dbFields = append(dbFields, parsedField.Name)
						}
					}
				}
			}
		}
		if len(dbFields) > 0 {
			return query.Select(dbFields)
		}
		return query
	}
}

func OrdersScope(orders []dboptions.Order) func(query *gorm.DB) *gorm.DB {
	return func(query *gorm.DB) *gorm.DB {
		for _, order := range orders {
			query = query.Order(
				clause.OrderByColumn{
					Column: clause.Column{Name: order.Field.ToSnake()},
					Desc:   order.Desc,
				},
			)
		}
		return query
	}
}

func PaginationScope(pagination dboptions.Pagination) func(query *gorm.DB) *gorm.DB {
	return func(query *gorm.DB) *gorm.DB {
		limit := pagination.Limit()
		if limit > 0 {
			query = query.Limit(limit)
		}
		offset := pagination.Offset()
		if offset > 0 {
			query = query.Offset(offset)
		}
		return query
	}
}
