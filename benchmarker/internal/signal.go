package internal

import (
	"context"
	"os"
	"os/signal"
)

// https://github.com/golang/go/issues/60756
// https://go-review.googlesource.com/c/go/+/579375

// SignalError returned by context.Cause when a context is canceled by a signal.
//
// Example:
//
//	cause := context.Cause(ctx)
//	var cs *os.SignalError
//	if errors.As(cause, &cs) {
//		fmt.Println("Process terminating after receiving", cs.Signal())
//	}
type SignalError struct {
	// Signal cancelled.
	Signal os.Signal
}

// Error from the canceled signal.
func (e *SignalError) Error() string {
	return "canceled by " + e.Signal.String() + " signal"
}

func NotifyContext(parent context.Context, signals ...os.Signal) (ctx context.Context, stop context.CancelFunc) {
	ctx, cancel := context.WithCancelCause(parent)
	c := &signalCtx{
		Context: ctx,
		cancel:  cancel,
		signals: signals,
	}
	c.ch = make(chan os.Signal, 1)
	signal.Notify(c.ch, c.signals...)
	if ctx.Err() == nil {
		go func() {
			select {
			case sig := <-c.ch:
				c.cancel(&SignalError{Signal: sig})
			case <-c.Done():
			}
		}()
	}
	return c, c.stop
}

type signalCtx struct {
	context.Context

	cancel  context.CancelCauseFunc
	signals []os.Signal
	ch      chan os.Signal
}

func (c *signalCtx) stop() {
	c.cancel(context.Canceled)
	signal.Stop(c.ch)
}

type stringer interface {
	String() string
}

func (c *signalCtx) String() string {
	var buf []byte
	// We know that the type of c.Context is context.cancelCtx, and we know that the
	// String method of cancelCtx returns a string that ends with ".WithCancel".
	name := c.Context.(stringer).String()
	name = name[:len(name)-len(".WithCancel")]
	buf = append(buf, "signal.NotifyContext("+name...)
	if len(c.signals) != 0 {
		buf = append(buf, ", ["...)
		for i, s := range c.signals {
			buf = append(buf, s.String()...)
			if i != len(c.signals)-1 {
				buf = append(buf, ' ')
			}
		}
		buf = append(buf, ']')
	}
	buf = append(buf, ')')
	return string(buf)
}
