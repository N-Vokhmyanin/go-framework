package provider

import (
	"context"
	"github.com/N-Vokhmyanin/go-framework/transactions"
	"github.com/N-Vokhmyanin/go-framework/transactions/gormtx"
)

type adapter struct {
	base gormtx.Service
}

func (a *adapter) Start(ctx context.Context) (transactions.Transaction, context.Context) {
	return a.base.Start(ctx)
}
