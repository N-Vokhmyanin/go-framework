package contracts

import "context"

type Dispatcher interface {
	Listen(l interface{})
	Fire(ctx context.Context, e interface{})
}
