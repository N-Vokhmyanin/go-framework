package service

import (
	"context"
	"sync"

	"github.com/N-Vokhmyanin/go-framework/errors"
	"github.com/N-Vokhmyanin/go-framework/logger"
	"github.com/N-Vokhmyanin/go-framework/transactions"
	"github.com/N-Vokhmyanin/go-framework/transactions/gormtx"
	"gorm.io/gorm"
)

type transaction struct {
	sync.Mutex
	db        *gorm.DB
	newDB     func() *gorm.DB
	log       logger.Logger
	count     int
	onSuccess []transactions.SuccessCallback
	onFail    []transactions.FailCallback
}

func (t *transaction) Begin() {
	if t == nil {
		return
	}

	t.Lock()
	defer t.Unlock()

	if t.db == nil {
		// not started or finished
		t.db = t.newDB().Begin()
	}

	t.count += 1
}

func (t *transaction) OnSuccess(cb transactions.SuccessCallback) {
	if t == nil {
		return
	}

	t.Lock()
	defer t.Unlock()

	if t.db == nil {
		// not started or finished
		t.log.Error("incorrect use of transactions: using tx.OnSuccess when transaction is not started")
		return
	}

	t.onSuccess = append(t.onSuccess, cb)
}

func (t *transaction) OnFail(cb transactions.FailCallback) {
	if t == nil {
		return
	}

	t.Lock()
	defer t.Unlock()

	if t.db == nil {
		// not started or finished
		t.log.Error("incorrect use of transactions: using tx.OnFail when transaction is not started")
		return
	}

	t.onFail = append(t.onFail, cb)
}

func (t *transaction) End(ctx context.Context, inputErr error) (outErr error) {
	if t == nil {
		return nil
	}

	t.Lock()
	defer t.Unlock()

	if t.db == nil {
		// not started or finished
		t.log.Error("incorrect use of transactions: using tx.End when transaction is not started")
		return inputErr
	}

	t.count -= 1
	if t.count != 0 {
		return inputErr
	}

	defer func() {
		cbCtx := gormtx.InjectToContext(ctx, nil)
		if outErr == nil {
			for _, cb := range t.onSuccess {
				cb(cbCtx)
			}
		} else {
			for _, cb := range t.onFail {
				cb(cbCtx, inputErr)
			}
		}
		t.onSuccess = nil
		t.onFail = nil
		t.db = nil
	}()

	dbCtx := t.db.Statement.Context
	if dbCtx != nil && dbCtx.Err() != nil {
		return dbCtx.Err()
	}

	if inputErr == nil {
		outErr = t.db.Commit().Error
		if outErr != nil {
			return errors.WrapWith(outErr, "commit failed")
		}
	} else {
		outErr = t.db.Rollback().Error
		if outErr != nil {
			return errors.WrapWith(outErr, "rollback failed")
		}
	}

	return inputErr
}

func (t *transaction) DB() *gorm.DB {
	if t == nil {
		return nil
	}

	t.Lock()
	defer t.Unlock()

	if t.db != nil {
		return t.db
	}

	return t.newDB()
}
