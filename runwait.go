package funcgroups

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
	Debug   *bool
}

var timeout = 5 * time.Second
var ctx, cancel = context.WithTimeout(context.Background(), timeout)

// Defaults
func DefaultOptions() *Options {
	return &Options{
		Timeout: timeout,
		Ctx:     ctx,
		cancel:  cancel,
		Debug:   BoolPointer(false),
	}
}

func BoolPointer(b bool) *bool {
	return &b
}

func check(o *Options) *Options {
	if o == nil {
		fmt.Println("Using default options")
		return DefaultOptions()
	}

	if o.Debug == nil {
		o.Debug = BoolPointer(false)
	}

	if o.Timeout == 0 {
		o.Timeout = timeout
	}
	if o.Ctx == nil {
		o.Ctx, o.cancel = context.WithTimeout(context.Background(), o.Timeout)
	} else {
		o.Ctx, o.cancel = context.WithTimeout(o.Ctx, o.Timeout)
	}
	return o
}

// RunWait executes the provided functions concurrently and waits for them all to complete.
// The functions are executed in separate goroutines. No errors are collected.
func RunWait(functions []Function, opts *Options) {
	opts = check(opts)

	length := len(functions)
	count := length
	waitChan := make(chan struct{}, length)

	if *opts.Debug {
		fmt.Printf("Starting %d jobs.\n", length)
	}

	for _, fu := range functions {
		go func(f Function) {
			f()

			waitChan <- struct{}{}
		}(fu)
	}

	for {
		select {
		case <-opts.Ctx.Done():
			return

		case <-waitChan:
			count--
			if count == 0 {
				if *opts.Debug {
					fmt.Printf("All %d jobs done\n", length)
				}
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
	opts = check(opts)
	length := len(functions)
	count := length
	waitChan := make(chan struct{}, length)

	if *opts.Debug {
		fmt.Printf("Starting %d jobs.\n", length)
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
			return

		case <-waitChan:
			count--
			if count == 0 {
				if *opts.Debug {
					fmt.Printf("All %d jobs done\n", length)
				}
				if errGroup != nil {
					fmt.Printf("Errors encountered %s\n", errGroup)
				}
				opts.cancel()
			}
		}
	}
}
