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
	Timeout time.Duration
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
}

var timeout = 5 * time.Second
var ctx, cancel = context.WithTimeout(context.Background(), timeout)

// Defaults
func DefaultOptions() *Options {
	return &Options{
		Timeout: timeout,
		Debug:   false,
	}
}

func check(opts *Options) *Options {
	if opts == nil {
		log.Println("Using default options")
		return DefaultOptions()
	}

	if opts.Timeout == 0 {
		opts.Timeout = timeout
	}

	// if opts.Ctx == nil {
	// 	opts.Ctx, opts.cancel = context.WithTimeout(context.Background(), opts.Timeout)
	// } else {
	// 	opts.Ctx, opts.cancel = context.WithTimeout(opts.Ctx, opts.Timeout)
	// }

	return opts
}

// RunWait executes the provided functions concurrently and waits for them all to complete.
// The functions are executed in separate goroutines. No errors are collected.
func (g *noErr) RunWait(ctx context.Context, secs time.Duration) {
	count := g.length
	if secs != 0 {
		timeout = time.Second * secs
	}

	if ctx == nil {
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	} else {
		ctx, cancel = context.WithTimeout(ctx, timeout)
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
		case <-ctx.Done():
			if ctx.Err() != nil {
				log.Println("Context done. Reason:", ctx.Err().Error())
			}
			return

		case <-g.wait:
			count--
			if count == 0 {
				if g.Options.Debug {
					log.Println("All " + strconv.Itoa(g.length) + " jobs done")
				}
				cancel()
			}
		}
	}
}

// RunWaitErr executes the provided functions concurrently and waits for them all to complete.
// The functions are executed in separate goroutines. Errors are collected.
func (g *withErr) RunWaitErr(ctx context.Context, secs time.Duration) (errGroup error) {
	var err error
	count := g.length
	if secs != 0 {
		timeout = time.Second * secs
	}
	if ctx == nil {
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	} else {
		ctx, cancel = context.WithTimeout(ctx, timeout)
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
		case <-ctx.Done():
			if ctx.Err() != nil {
				log.Println("Context done. Reason:", ctx.Err().Error())
			}
			return

		case <-g.wait:
			count--
			if count == 0 {
				if g.Options.Debug {
					log.Println("All " + strconv.Itoa(g.length) + " jobs done")
				}
				cancel()
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
