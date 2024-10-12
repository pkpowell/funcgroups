package groups

import (
	"context"
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
	// Cancel:  cancel,
}

// RunWait executes the provided functions concurrently and waits for them all to complete.
// The functions are executed in separate goroutines, and a sync.WaitGroup is used to
// wait for all functions to finish before returning.
func RunWait(functions []Function, Opts *Options) {
	waitChan := make(chan struct{}, len(functions))
	length := len(functions)
	to := Opts.Timeout
	fmt.Printf("Starting %d jobs. Timeout = %s\n", length, to.String())
	if Opts.Ctx == nil {
		Opts.Ctx, Opts.cancel = context.WithTimeout(context.Background(), to)
	} else {
		Opts.Ctx, Opts.cancel = context.WithTimeout(Opts.Ctx, to)
	}

	for _, fu := range functions {
		go func(f Function) {
			f()
			waitChan <- struct{}{}
		}(fu)
	}

	for {
		select {
		case <-Opts.Ctx.Done():
			fmt.Printf("Canceling jobs %s\n", Opts.Ctx.Err())
			Opts.cancel()
			return
		case <-waitChan:
			length--

			if length == 0 {
				fmt.Printf("All %d jobs done\n", length)
				if Opts.Ctx.Err() != nil {
					fmt.Printf("Context error %s\n", Opts.Ctx.Err())
				}
				Opts.cancel()
				return
			}
		}
	}
}

// RunWaitErr executes the provided functions concurrently and waits for them all to complete.
// The functions are executed in separate goroutines, and a sync.WaitGroup is used to
// wait for all functions to finish before returning.
func RunWaitErr(functions []FunctionErr, Opts *Options) {
	waitChan := make(chan struct{}, len(functions))
	length := len(functions)
	to := Opts.Timeout
	fmt.Printf("Starting %d jobs. Timeout = %s\n", length, to.String())
	if Opts.Ctx == nil {
		Opts.Ctx, Opts.cancel = context.WithTimeout(context.Background(), to)
	} else {
		Opts.Ctx, Opts.cancel = context.WithTimeout(Opts.Ctx, to)
	}
	// ctx, cancel := context.WithTimeout(Opts.Ctx, to)
	var errGroup error

	for _, fu := range functions {
		go func(f FunctionErr) {
			// fmt.Printf("Starting job #%d\n", idx)
			err := f()
			if err != nil {
				errGroup = fmt.Errorf("%s %w", errGroup, err)
			}
			waitChan <- struct{}{}
		}(fu)
	}

	for {
		select {
		case <-Opts.Ctx.Done():
			fmt.Printf("Canceling jobs %s\n", Opts.Ctx.Err())
			Opts.cancel()
			return
		case <-waitChan:
			length--
			if length == 0 {
				fmt.Printf("All %d jobs done\n", len(functions))
				if errGroup != nil {
					fmt.Printf("Errors encountered %s\n", errGroup)
				}
				Opts.cancel()
				// return
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
