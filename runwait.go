package funcgroups

import (
	"context"
	"errors"
	"log"
	"strconv"
	"time"
)

type Function func()
type FunctionErr func() (err error)

type Options struct {
	Timeout time.Duration
	Ctx     context.Context
	cancel  context.CancelFunc
	Debug   bool
	Timer   bool
}

var timeout = 5 * time.Second
var ctx, cancel = context.WithTimeout(context.Background(), timeout)

// var pcsbuf [3]loc.PC
// var pcs loc.PCs

// Defaults
func DefaultOptions() *Options {
	return &Options{
		Timeout: timeout,
		Ctx:     ctx,
		cancel:  cancel,
		Debug:   false,
		Timer:   false,
	}
}

// func BoolPointer(b bool) *bool {
// 	return &b
// }

func check(o *Options) *Options {
	if o == nil {
		log.Println("Using default options")
		return DefaultOptions()
	}

	// if o.Debug == nil {
	// 	o.Debug = false
	// }

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

	if opts.Debug {
		log.Println("Starting " + strconv.Itoa(length) + " jobs")
	}

	for _, fu := range functions {
		go func(f Function) {
			if opts.Timer {
				timer(f)
			} else {
				f()
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
				if opts.Debug {
					log.Println("All " + strconv.Itoa(length) + " jobs done")
				}
				if opts.Ctx.Err() != nil {
					log.Println("Context error ", opts.Ctx.Err().Error())
				}
				opts.cancel()
			}
		}
	}
}

// RunWaitErr executes the provided functions concurrently and waits for them all to complete.
// The functions are executed in separate goroutines. Errors are collected.
func RunWaitErr(functions []FunctionErr, opts *Options) (errGroup error) {
	var err error
	opts = check(opts)
	length := len(functions)
	count := length
	waitChan := make(chan struct{}, length)

	if opts.Debug {
		log.Println("Starting " + strconv.Itoa(length) + " jobs.")
	}

	// var errGroup error

	for _, fu := range functions {
		go func(f FunctionErr) {
			if opts.Timer {
				err = timerWithErr(f)
			} else {
				err = f()
			}
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
				if opts.Debug {
					log.Println("All " + strconv.Itoa(length) + " jobs done")
				}

				opts.cancel()
			}
		}
	}
}

func timerWithErr(fn func() error) (err error) {
	start := time.Now()
	err = fn()
	// pcs = loc.CallersFill(2, pcsbuf[:])
	elapsed := time.Since(start)
	log.Printf("function : %p\n", fn)
	log.Println("duration", elapsed)

	return
}

func timer(fn func()) {
	start := time.Now()
	fn()
	// pcs = loc.CallersFill(2, pcsbuf[:])
	elapsed := time.Since(start)
	log.Printf("function : %p\n", fn)
	log.Println("duration", elapsed)
}
