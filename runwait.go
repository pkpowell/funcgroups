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
type FunctionErr func() error

type Options struct {
	// Timeout time.Duration
	// Ctx     context.Context
	Debug bool
	// cancel  context.CancelFunc
}

type groupNoErr struct {
	fn   Function
	name string
}

type noErr struct {
	fns []groupNoErr
	*Options
	length int
	wait   chan struct{}
	ctx    context.Context
	cancel context.CancelFunc
}

func New(fns []Function, opts *Options) *noErr {
	opts = check(opts)
	log.Println("opts", opts)
	var noErr = &noErr{
		Options: opts,
		fns:     make([]groupNoErr, len(fns)),
		length:  len(fns),
		wait:    make(chan struct{}, len(fns)),
	}

	for i, fn := range fns {
		noErr.fns[i] = groupNoErr{
			name: runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name(),
			fn:   fn,
		}
	}
	return noErr
}

func NewWithErr(fns []FunctionErr, opts *Options) *withErr {
	opts = check(opts)
	log.Println("opts", opts)
	var withErr = &withErr{
		Options: opts,
		fns:     make([]funcWithErr, len(fns)),
		length:  len(fns),
		wait:    make(chan struct{}, len(fns)),
	}

	for i, fn := range fns {
		withErr.fns[i] = funcWithErr{
			name: runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name(),
			fn:   fn,
		}
	}
	return withErr
}

type funcWithErr struct {
	fn   FunctionErr
	name string
}

type withErr struct {
	fns []funcWithErr
	*Options
	length int
	wait   chan struct{}
	ctx    context.Context
	cancel context.CancelFunc
}

var timeout = 5 * time.Second

// var ctx, cancel = context.WithTimeout(context.Background(), timeout)

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

// RunWait executes the provided functions concurrently and waits for them all to complete.
// The functions are executed in separate goroutines. No errors are collected.
func (g *noErr) RunWait(pctx context.Context, secs time.Duration) {
	count := g.length
	if secs == 0 {
		timeout = time.Second * 10
	} else {
		timeout = time.Second * secs
	}

	if pctx == nil {
		g.ctx, g.cancel = context.WithTimeout(context.Background(), timeout)
	} else {
		g.ctx, g.cancel = context.WithTimeout(pctx, timeout)
	}

	for _, fg := range g.fns {
		go func() {
			if g.Options.Debug {
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
				if g.Options.Debug {
					log.Println(strconv.Itoa(g.length) + " jobs done. No errors")
				}
			default:
				log.Println("Context done error:", g.ctx.Err().Error())
			}

			return

		case <-g.wait:
			count--
			if count == 0 {
				if g.Options.Debug {
					log.Println("All " + strconv.Itoa(g.length) + " jobs done")
				}
				g.cancel()
			}
		}
	}
}

// RunWaitErr executes the provided functions concurrently and waits for them all to complete.
// The functions are executed in separate goroutines. Errors are collected.
func (g *withErr) RunWaitErr(pctx context.Context, secs time.Duration) (errGroup error) {
	var err error
	count := g.length
	if secs == 0 {
		timeout = time.Second * 10
	} else {
		timeout = time.Second * secs
	}

	if pctx == nil {
		g.ctx, g.cancel = context.WithTimeout(context.Background(), timeout)
	} else {
		g.ctx, g.cancel = context.WithTimeout(pctx, timeout)
	}

	for _, fg := range g.fns {
		go func() {
			if g.Options.Debug {
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
				if g.Options.Debug {
					log.Println("All " + strconv.Itoa(g.length) + " jobs done. No errors")
				}
			default:
				log.Println("Context done error:", g.ctx.Err().Error())
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
