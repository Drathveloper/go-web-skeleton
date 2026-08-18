package event

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

type Handler func(ctx context.Context, event Event) error

type CancelFunc func()

type subscription struct {
	ch   chan Event
	once sync.Once
}

func (s *subscription) close() {
	s.once.Do(func() { close(s.ch) })
}

const (
	// defaultWrappedErrorTemplate is declared here on purpose: pkg is the reusable
	// layer and must not import anything from the application, not even constants.
	defaultWrappedErrorTemplate = "%s: %w"

	newEventBusErrMsg      = "invalid event bus options"
	shutdownEventBusErrMsg = "event bus shutdown failed"
)

type Bus struct {
	subscribers   map[string][]*subscription
	shutdownFuncs []CancelFunc
	opts          Options
	wg            sync.WaitGroup
	mu            sync.RWMutex
	closed        bool
}

func NewBus(opts Options) (*Bus, error) {
	if err := opts.validate(); err != nil {
		return nil, fmt.Errorf(defaultWrappedErrorTemplate, newEventBusErrMsg, err)
	}
	return &Bus{
		subscribers:   make(map[string][]*subscription),
		opts:          opts,
		shutdownFuncs: make([]CancelFunc, 0),
	}, nil
}

// Subscribe registers handler for eventName and starts its worker. Subscribing to a
// bus that has already been shut down is a no-op and returns a no-op CancelFunc.
func (eb *Bus) Subscribe(eventName string, handler Handler) CancelFunc {
	cancelFunc, ok := eb.addSubscription(eventName, handler)
	if !ok {
		slog.Warn("subscribe on an already shut down event bus", slog.String("event", eventName))
		return func() {}
	}
	return cancelFunc
}

func (eb *Bus) Publish(event Event) {
	slog.Debug("publishing event", slog.String("event", event.GetName()))
	eb.mu.RLock()
	subs, ok := eb.subscribers[event.GetName()]
	if !ok {
		slog.Debug("no subscribers for event", slog.String("event", event.GetName()))
		eb.mu.RUnlock()
		return
	}
	// The delivery happens under the read lock because removeSubscription closes the
	// subscription channel while holding the write lock: sending on a channel that is
	// being closed concurrently is a data race and panics.
	dropped := 0
	for _, sub := range subs {
		select {
		case sub.ch <- event:
		default:
			dropped++
		}
	}
	eb.mu.RUnlock()

	// Reported outside the lock: onDropped is caller-supplied code and must never run
	// while the bus is locked.
	onDropped := eb.opts.onDropped()
	for range dropped {
		onDropped(event)
	}
}

// Shutdown cancels every subscription and waits for the in-flight handlers to finish.
// It returns nil on a clean shutdown, and an error wrapping ctx.Err() when ctx is
// cancelled or its deadline passes before the handlers are done. In that case the
// remaining handlers keep running in the background until their own timeout expires.
// The bus is closed for new subscriptions as soon as Shutdown is called.
func (eb *Bus) Shutdown(ctx context.Context) error {
	eb.mu.Lock()
	eb.closed = true
	cancelFuncs := eb.shutdownFuncs
	eb.shutdownFuncs = nil
	eb.mu.Unlock()

	for _, cancelFunc := range cancelFuncs {
		cancelFunc()
	}

	done := make(chan struct{})
	go func() {
		eb.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
	}

	// ctx and the last worker may become ready at the same time: prefer reporting the
	// clean shutdown that actually happened.
	select {
	case <-done:
		return nil
	default:
		return fmt.Errorf(defaultWrappedErrorTemplate, shutdownEventBusErrMsg, ctx.Err())
	}
}

func (eb *Bus) runWorker(sub *subscription, handler Handler) {
	var sem chan struct{}
	if eb.opts.WorkerConcurrency > 0 {
		sem = make(chan struct{}, eb.opts.WorkerConcurrency)
	}

	var waitGroup sync.WaitGroup
	onError := eb.opts.onError()

	for event := range sub.ch {
		waitGroup.Add(1)
		if sem != nil {
			sem <- struct{}{}
		}
		go func(evt Event) {
			defer func() {
				if r := recover(); r != nil {
					slog.Error("event handler panic",
						slog.String("event", evt.GetName()), slog.String("panic", fmt.Sprint(r)))
				}
			}()
			ctx, cancelFunc := context.WithTimeout(context.Background(), eb.opts.DefaultTimeout)
			defer cancelFunc()
			defer waitGroup.Done()
			if sem != nil {
				defer func() { <-sem }()
			}
			if err := handler(ctx, evt); err != nil {
				onError(evt, err)
			}
		}(event)
	}
	waitGroup.Wait()
}

// addSubscription registers a subscription for eventName and starts its worker. The
// worker is started while holding the write lock so that no goroutine can be added to
// the wait group once Shutdown started waiting on it. It returns false when the bus is
// already shut down.
func (eb *Bus) addSubscription(eventName string, handler Handler) (CancelFunc, bool) {
	sub := &subscription{
		ch: make(chan Event, eb.opts.ChannelBuffer),
	}

	var cancelOnce sync.Once
	cancelFunc := func() {
		cancelOnce.Do(func() {
			slog.Debug("cancelling subscription", slog.String("event", eventName))
			eb.removeSubscription(eventName, sub)
		})
	}

	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.closed {
		return nil, false
	}

	eb.subscribers[eventName] = append(eb.subscribers[eventName], sub)
	eb.shutdownFuncs = append(eb.shutdownFuncs, cancelFunc)
	eb.wg.Go(func() {
		eb.runWorker(sub, handler)
	})

	return cancelFunc, true
}

func (eb *Bus) removeSubscription(eventName string, target *subscription) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	subs, ok := eb.subscribers[eventName]
	if !ok {
		return
	}

	newSubs := make([]*subscription, 0, len(subs)-1)
	for _, sub := range subs {
		if sub != target {
			newSubs = append(newSubs, sub)
		}
	}

	if len(newSubs) == 0 {
		delete(eb.subscribers, eventName)
	} else {
		eb.subscribers[eventName] = newSubs
	}

	target.close()
}
