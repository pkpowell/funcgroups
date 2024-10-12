package groups

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

type Function func()
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
	// ctx, cancel := context.WithTimeout(Opts.Ctx, to)

	for idx, fu := range functions {
		go func(f Function) {
			fmt.Printf("Starting job #%d\n", idx)
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
			fmt.Printf("length = %d\n", length)
			if length == 0 {
				fmt.Printf("All jobs done\n")
				if Opts.Ctx.Err() != nil {
					fmt.Printf("Context error %s\n", Opts.Ctx.Err())
				}
				Opts.cancel()
				return

			} else {
				fmt.Printf("Still waiting...\n")
			}
		}
	}
}

func CallerName(skip int) string {
	pc, _, _, ok := runtime.Caller(skip + 1)
	if !ok {
		return ""
	}
	f := runtime.FuncForPC(pc)
	if f == nil {
		return ""
	}
	return f.Name()
}
