package event

import (
	"errors"
	"log/slog"
	"time"
)

var ErrInvalidChannelBuffer = errors.New("invalid channel buffer")
var ErrInvalidWorkerConcurrency = errors.New("invalid worker concurrency")
var ErrInvalidDefaultTimeout = errors.New("invalid default timeout")

const (
	defaultChannelBuffer     = 100
	defaultWorkerConcurrency = 10
	defaultDefaultTimeout    = 30 * time.Second
)

type DropHandlerFunc func(Event)
type ErrorHandlerFunc func(Event, error)

type Options struct {
	OnError           ErrorHandlerFunc
	OnDropped         DropHandlerFunc
	ChannelBuffer     int
	WorkerConcurrency int
	DefaultTimeout    time.Duration
}

func NewDefaultOptions() Options {
	return Options{
		ChannelBuffer:     defaultChannelBuffer,
		WorkerConcurrency: defaultWorkerConcurrency,
		DefaultTimeout:    defaultDefaultTimeout,
		OnError:           defaultErrorHandler,
		OnDropped:         defaultDropHandler,
	}
}

func (o *Options) validate() error {
	if o.ChannelBuffer <= 0 {
		return ErrInvalidChannelBuffer
	}
	if o.WorkerConcurrency <= 0 {
		return ErrInvalidWorkerConcurrency
	}
	if o.DefaultTimeout <= 0 {
		return ErrInvalidDefaultTimeout
	}
	return nil
}

func (o *Options) onError() ErrorHandlerFunc {
	if o.OnError != nil {
		return o.OnError
	}
	return defaultErrorHandler
}

func (o *Options) onDropped() DropHandlerFunc {
	if o.OnDropped != nil {
		return o.OnDropped
	}
	return defaultDropHandler
}

func defaultErrorHandler(event Event, err error) {
	slog.Error("event handler error",
		slog.String("event", event.GetName()), slog.String("error", err.Error()))
}

func defaultDropHandler(event Event) {
	slog.Warn("event dropped", slog.String("event", event.GetName()))
}
