package gormtx

import (
	"context"
	"github.com/N-Vokhmyanin/go-framework/transactions"
	"gorm.io/gorm"
)

var ctxTxKey = struct{}{}

type Transaction interface {
	transactions.Transaction
	DB() *gorm.DB
}

type Service interface {
	Start(ctx context.Context) (Transaction, context.Context)
	DB(ctx context.Context) *gorm.DB
}

func InjectToContext(ctx context.Context, tx Transaction) context.Context {
	return context.WithValue(ctx, ctxTxKey, tx)
}

func ExtractFromContext(ctx context.Context) Transaction {
	val := ctx.Value(ctxTxKey)
	tx, ok := val.(Transaction)
	if ok {
		return tx
	}
	return nil
}
