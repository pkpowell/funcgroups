package groups

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type Function func()
type FunctionErr func() (err error)

type Options struct {
	Timeout time.Duration
	Ctx     context.Context
	cancel  context.CancelFunc
}

var timeout = 5 * time.Second

// Defaults
var opts = Options{
	Timeout: timeout,
	Ctx:     context.Background(),
}

// RunWait executes the provided functions concurrently and waits for them all to complete.
// The functions are executed in separate goroutines. No errors are collected.
func RunWait(functions []Function, opts *Options) {
	length := len(functions)
	count := length
	waitChan := make(chan struct{}, length)

	if opts.Ctx == nil {
		opts.Ctx, opts.cancel = context.WithTimeout(context.Background(), opts.Timeout)
	} else {
		opts.Ctx, opts.cancel = context.WithTimeout(opts.Ctx, opts.Timeout)
	}
	fmt.Printf("Starting %d jobs. Timeout = %s\n", length, opts.Timeout.String())

	for _, fu := range functions {
		go func(f Function) {
			f()
			waitChan <- struct{}{}
		}(fu)
	}

	for {
		select {
		case <-opts.Ctx.Done():
			// fmt.Printf("Returning")
			return

		case <-waitChan:
			count--
			if count == 0 {
				fmt.Printf("All %d jobs done\n", length)
				if opts.Ctx.Err() != nil {
					fmt.Printf("Context error %s\n", opts.Ctx.Err())
				}
				opts.cancel()
			}
		}
	}
}

// RunWaitErr executes the provided functions concurrently and waits for them all to complete.
// The functions are executed in separate goroutines. Errors are collected.
func RunWaitErr(functions []FunctionErr, opts *Options) {
	length := len(functions)
	count := length
	waitChan := make(chan struct{}, length)

	fmt.Printf("Starting %d jobs and collecting errs. Timeout = %s\n", length, opts.Timeout.String())
	if opts.Ctx == nil {
		opts.Ctx, opts.cancel = context.WithTimeout(context.Background(), opts.Timeout)
	} else {
		opts.Ctx, opts.cancel = context.WithTimeout(opts.Ctx, opts.Timeout)
	}
	var errGroup error

	for _, fu := range functions {
		go func(f FunctionErr) {
			err := f()
			if err != nil {
				errGroup = errors.Join(errGroup, err)
			}
			waitChan <- struct{}{}
		}(fu)
	}

	for {
		select {
		case <-opts.Ctx.Done():
			// fmt.Printf("Returning")
			return

		case <-waitChan:
			count--
			if count == 0 {
				fmt.Printf("All %d jobs done\n", length)
				if errGroup != nil {
					fmt.Printf("Errors encountered %s\n", errGroup)
				}
				opts.cancel()
			}
		}
	}
}

// func CallerName(skip int) string {
// 	pc, _, _, ok := runtime.Caller(skip + 1)
// 	if !ok {
// 		return ""
// 	}
// 	f := runtime.FuncForPC(pc)
// 	if f == nil {
// 		return ""
// 	}
// 	return f.Name()
// }
