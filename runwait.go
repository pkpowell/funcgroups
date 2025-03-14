package funcgroups

import (
	"context"
	"errors"
	"log"
	"time"

	"tlog.app/go/loc"
)

const defaultTimeout = 10

type Function func()
type FunctionErr func() error

type Options struct {
	Debug bool
}

type groupNoErr struct {
	fn   Function
	name string
}

type FuncsNoErrs struct {
	*Options
	fns     []groupNoErr
	length  int
	wait    chan struct{}
	ctx     context.Context
	cancel  context.CancelFunc
	timeout time.Duration
}

// New creates a new instance of FuncsNoErrs.
// The functions are executed concurrently, no errors are collected.
func New(fns []Function, opts *Options) *FuncsNoErrs {
	opts = check(opts)
	var noErr = &FuncsNoErrs{
		Options: opts,
		fns:     make([]groupNoErr, len(fns)),
		length:  len(fns),
		wait:    make(chan struct{}, len(fns)),
	}

	for i, fn := range fns {
		name, _, _ := loc.FuncEntryFromFunc(fn).NameFileLine()
		noErr.fns[i] = groupNoErr{
			name: name,
			fn:   fn,
		}
	}
	return noErr
}

// NewWithErr creates a new instance of FuncsWithErr.
// The functions are executed concurrently, errors are collected.
func NewWithErr(fns []FunctionErr, opts *Options) *FuncsWithErr {
	opts = check(opts)
	var withErr = &FuncsWithErr{
		Options: opts,
		fns:     make([]funcWithErr, len(fns)),
		length:  len(fns),
		wait:    make(chan struct{}, len(fns)),
	}

	for i, fn := range fns {
		name, _, _ := loc.FuncEntryFromFunc(fn).NameFileLine()
		withErr.fns[i] = funcWithErr{
			name: name,
			fn:   fn,
		}
	}
	return withErr
}

type funcWithErr struct {
	fn   FunctionErr
	name string
}

type FuncsWithErr struct {
	fns []funcWithErr
	*Options
	length  int
	wait    chan struct{}
	ctx     context.Context
	cancel  context.CancelFunc
	timeout time.Duration
}

// Defaults
func DefaultOptions() *Options {
	return &Options{
		Debug: false,
	}
}

func check(opts *Options) *Options {
	if opts == nil {
		log.Println("Using default options")
		return DefaultOptions()
	}

	return opts
}

// Run executes the provided functions concurrently and waits for them all to complete.
// The functions are executed in separate goroutines. No errors are collected.
func (g *FuncsNoErrs) Run(ctx context.Context, seconds time.Duration) {
	count := g.length
	if seconds == 0 {
		g.timeout = time.Second * defaultTimeout
	} else {
		g.timeout = time.Second * seconds
	}

	if ctx == nil {
		g.ctx, g.cancel = context.WithTimeout(context.Background(), g.timeout)
	} else {
		g.ctx, g.cancel = context.WithTimeout(ctx, g.timeout)
	}

	for _, fg := range g.fns {
		go func() {
			if g.Debug {
				timer(fg)
			} else {
				fg.fn()
			}

			g.wait <- struct{}{}
		}()
	}

	for {
		select {
		case <-g.ctx.Done():
			switch g.ctx.Err() {
			case nil, context.Canceled:
			default:
				log.Println("Context error:", g.ctx.Err().Error())
			}

			return

		case <-g.wait:
			count--
			if count == 0 {
				g.cancel()
			}
		}
	}
}

// RunErr executes the provided functions concurrently and waits for them all to complete.
// The functions are executed in separate goroutines. Errors are collected.
func (g *FuncsWithErr) RunErr(ctx context.Context, seconds time.Duration) (errGroup error) {
	var err error
	count := g.length
	if seconds == 0 {
		g.timeout = time.Second * defaultTimeout
	} else {
		g.timeout = time.Second * seconds
	}

	if ctx == nil {
		g.ctx, g.cancel = context.WithTimeout(context.Background(), g.timeout)
	} else {
		g.ctx, g.cancel = context.WithTimeout(ctx, g.timeout)
	}

	for _, fg := range g.fns {
		go func() {
			if g.Debug {
				err = timerWithErr(fg)
			} else {
				err = fg.fn()
			}
			if err != nil {
				errGroup = errors.Join(errGroup, err)
			}
			g.wait <- struct{}{}
		}()
	}

	for {
		select {
		case <-g.ctx.Done():
			switch g.ctx.Err() {
			case nil, context.Canceled:
			default:
				log.Println("Context error:", g.ctx.Err().Error())
			}

			return

		case <-g.wait:
			count--
			if count == 0 {
				g.cancel()
			}
		}
	}
}

func timerWithErr(g funcWithErr) (err error) {
	start := time.Now()
	err = g.fn()
	elapsed := time.Since(start)
	log.Println(g.name, elapsed)

	return
}

func timer(g groupNoErr) {
	start := time.Now()
	g.fn()
	elapsed := time.Since(start)
	log.Println(g.name, elapsed)
}

func Time(f Function, n string) {
	start := time.Now()
	f()
	log.Println(n, time.Since(start))
}

func TimeWithErr(f FunctionErr, n string) (err error) {
	start := time.Now()
	err = f()
	log.Println(n, time.Since(start))
	return
}
