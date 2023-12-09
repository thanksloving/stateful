package stateful

import (
	"context"
	"errors"
	"fmt"

	pkgErrors "github.com/pkg/errors"
)

type Future struct {
	ch  chan struct{}
	fn  Func
	err error
}

type Func func() error

func NewFuture(fn Func) *Future {
	f := &Future{
		ch: make(chan struct{}),
		fn: fn,
	}
	f.start()
	return f
}

func (f *Future) start() {
	go func() {
		defer func() {
			if rval := recover(); rval != nil {
				if err, ok := rval.(error); ok {
					f.err = pkgErrors.WithStack(err)
				} else {
					rvalStr := fmt.Sprint(rval)
					f.err = pkgErrors.WithStack(errors.New(rvalStr))
				}
			}

			close(f.ch)
		}()

		// 执行结果
		f.err = f.fn()
		return
	}()
}

func (f *Future) Get(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-f.ch:
		return f.err
	}
}

func (f *Future) Done() bool {
	select {
	case <-f.ch:
		return true
	default:
		return false
	}
}
