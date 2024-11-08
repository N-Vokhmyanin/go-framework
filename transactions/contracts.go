package transactions

import "context"

type SuccessCallback func(ctx context.Context)
type FailCallback func(ctx context.Context, err error)

type Transaction interface {
	OnSuccess(cb SuccessCallback)
	OnFail(cb FailCallback)
	End(ctx context.Context, err error) error
}

type Service interface {
	Start(ctx context.Context) (Transaction, context.Context)
}
