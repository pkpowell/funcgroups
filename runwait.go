package funcgroups

import (
	"context"
	"errors"
	"log"
	"runtime"
	"strconv"
	"time"

	"github.com/goccy/go-reflect"
)

type Function func()
type FunctionErr func() (err error)

type Options struct {
	Timeout time.Duration
	Ctx     context.Context
	cancel  context.CancelFunc
	Debug   bool
}

type groupNoErr struct {
	f     Function
	fName string
}

type noErr struct {
	functions []groupNoErr
	*Options
}

func New(fns []Function, o *Options) *noErr {
	o = check(o)
	var noErr = &noErr{
		Options:   o,
		functions: make([]groupNoErr, len(fns)),
	}
	for i, f := range fns {
		noErr.functions[i] = groupNoErr{
			fName: runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name(),
			f:     f,
		}
	}
	return noErr
}

func NewWithErr(fns []FunctionErr, o *Options) *withErr {
	o = check(o)
	var withErr = &withErr{
		Options:   o,
		functions: make([]groupWithErr, len(fns)),
	}
	for i, f := range fns {
		withErr.functions[i] = groupWithErr{
			fName: runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name(),
			f:     f,
		}
	}
	return withErr
}

type groupWithErr struct {
	f     FunctionErr
	fName string
}

type withErr struct {
	functions []groupWithErr
	*Options
}

var timeout = 5 * time.Second
var ctx, cancel = context.WithTimeout(context.Background(), timeout)

// Defaults
func DefaultOptions() *Options {
	return &Options{
		Timeout: timeout,
		Ctx:     ctx,
		cancel:  cancel,
		Debug:   false,
	}
}

func check(o *Options) *Options {
	if o == nil {
		log.Println("Using default options")
		return DefaultOptions()
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
func (g *noErr) RunWait() {
	length := len(g.functions)
	count := length
	waitChan := make(chan struct{}, length)

	for _, fg := range g.functions {
		go func() {
			if g.Options.Debug {
				timer(fg)
			} else {
				fg.f()
			}

			waitChan <- struct{}{}
		}()
	}

	for {
		select {
		case <-g.Options.Ctx.Done():
			return

		case <-waitChan:
			count--
			if count == 0 {
				if g.Options.Debug {
					log.Println("All " + strconv.Itoa(length) + " jobs done")
				}
				if g.Options.Ctx.Err() != nil {
					log.Println("Context error ", g.Options.Ctx.Err().Error())
				}
				g.Options.cancel()
			}
		}
	}
}

// RunWaitErr executes the provided functions concurrently and waits for them all to complete.
// The functions are executed in separate goroutines. Errors are collected.
func (g *withErr) RunWaitErr() (errGroup error) {
	var err error
	length := len(g.functions)
	count := length
	waitChan := make(chan struct{}, length)

	for _, fu := range g.functions {
		go func() {
			if g.Options.Debug {
				err = timerWithErr(fu)
			} else {
				err = fu.f()
			}
			if err != nil {
				errGroup = errors.Join(errGroup, err)
			}
			waitChan <- struct{}{}
		}()
	}

	for {
		select {
		case <-g.Options.Ctx.Done():
			return

		case <-waitChan:
			count--
			if count == 0 {
				if g.Options.Debug {
					log.Println("All " + strconv.Itoa(length) + " jobs done")
				}

				g.Options.cancel()
			}
		}
	}
}

func timerWithErr(g groupWithErr) (err error) {
	start := time.Now()
	err = g.f()
	elapsed := time.Since(start)
	log.Println(g.fName, elapsed)

	return
}

func timer(g groupNoErr) {
	start := time.Now()
	g.f()
	elapsed := time.Since(start)
	log.Println(g.fName, elapsed)
}
