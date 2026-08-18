package event_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/Drathveloper/go-web-skeleton/pkg/event"
)

type testEvent struct {
	name string
}

func (e testEvent) GetName() string {
	return e.name
}

// handlerTimeout is how long a test waits for the workers before calling it a
// hang. Generous on purpose: the assertions are about behaviour, not speed.
const handlerTimeout = time.Second

func waitWithTimeout(t *testing.T, waitGroup *sync.WaitGroup) {
	t.Helper()

	done := make(chan struct{})
	go func() {
		waitGroup.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(handlerTimeout):
		t.Fatal("timed out waiting for the handlers to run")
	}
}

func newTestBus(t *testing.T, opts event.Options) *event.Bus {
	t.Helper()

	bus, err := event.NewBus(opts)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = bus.Shutdown(context.Background())
	})

	return bus
}

func TestNewBus_RejectsInvalidOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantErrIs  error
		name       string
		wantErrMsg string
		opts       event.Options
	}{
		{
			name:       "test new bus should reject a non positive channel buffer",
			opts:       event.Options{ChannelBuffer: 0, WorkerConcurrency: 1, DefaultTimeout: time.Second},
			wantErrIs:  event.ErrInvalidChannelBuffer,
			wantErrMsg: "invalid event bus options: invalid channel buffer",
		},
		{
			name:       "test new bus should reject a non positive worker concurrency",
			opts:       event.Options{ChannelBuffer: 1, WorkerConcurrency: 0, DefaultTimeout: time.Second},
			wantErrIs:  event.ErrInvalidWorkerConcurrency,
			wantErrMsg: "invalid event bus options: invalid worker concurrency",
		},
		{
			name:       "test new bus should reject a non positive default timeout",
			opts:       event.Options{ChannelBuffer: 1, WorkerConcurrency: 1, DefaultTimeout: 0},
			wantErrIs:  event.ErrInvalidDefaultTimeout,
			wantErrMsg: "invalid event bus options: invalid default timeout",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			bus, err := event.NewBus(tt.opts)

			require.Nil(t, bus)
			require.ErrorIs(t, err, tt.wantErrIs)
			require.Equal(t, tt.wantErrMsg, err.Error())
		})
	}
}

func TestBus_BasicPubSub(t *testing.T) {
	t.Parallel()

	bus := newTestBus(t, event.NewDefaultOptions())

	var waitGroup sync.WaitGroup
	waitGroup.Add(1)

	var received atomic.Value
	bus.Subscribe("ping", func(_ context.Context, e event.Event) error {
		defer waitGroup.Done()
		received.Store(e.GetName())

		return nil
	})

	bus.Publish(testEvent{name: "ping"})
	waitWithTimeout(t, &waitGroup)

	require.Equal(t, "ping", received.Load())
}

func TestBus_DeliversToEverySubscriber(t *testing.T) {
	t.Parallel()

	const subscribers = 5

	bus := newTestBus(t, event.NewDefaultOptions())

	var count atomic.Int32
	var waitGroup sync.WaitGroup
	waitGroup.Add(subscribers)

	for range subscribers {
		bus.Subscribe("tick", func(_ context.Context, _ event.Event) error {
			count.Add(1)
			waitGroup.Done()

			return nil
		})
	}

	bus.Publish(testEvent{name: "tick"})
	waitWithTimeout(t, &waitGroup)

	require.Equal(t, int32(subscribers), count.Load())
}

func TestBus_DoesNotDeliverToOtherEvents(t *testing.T) {
	t.Parallel()

	bus := newTestBus(t, event.NewDefaultOptions())

	var called atomic.Bool
	bus.Subscribe("expected", func(_ context.Context, _ event.Event) error {
		called.Store(true)

		return nil
	})

	// An event nobody subscribed to must not panic either.
	require.NotPanics(t, func() { bus.Publish(testEvent{name: "unrelated"}) })

	time.Sleep(50 * time.Millisecond)

	require.False(t, called.Load())
}

func TestBus_CancelFuncIsIdempotentAndStopsDelivery(t *testing.T) {
	t.Parallel()

	bus := newTestBus(t, event.NewDefaultOptions())

	var called atomic.Bool
	cancel := bus.Subscribe("x", func(_ context.Context, _ event.Event) error {
		called.Store(true)

		return nil
	})

	require.NotPanics(t, func() {
		for range 10 {
			cancel()
		}
	})

	bus.Publish(testEvent{name: "x"})
	time.Sleep(50 * time.Millisecond)

	require.False(t, called.Load(), "a cancelled subscription must stop receiving")
}

func TestBus_ErrorHandlerIsCalled(t *testing.T) {
	t.Parallel()

	handlerErr := errors.New("something went wrong")

	var errCount atomic.Int32
	var reported atomic.Value
	bus := newTestBus(t, event.Options{
		OnError: func(_ event.Event, err error) {
			reported.Store(err.Error())
			errCount.Add(1)
		},
		ChannelBuffer:     1,
		WorkerConcurrency: 1,
		DefaultTimeout:    10 * time.Second,
	})

	var waitGroup sync.WaitGroup
	waitGroup.Add(1)
	bus.Subscribe("fail", func(_ context.Context, _ event.Event) error {
		defer waitGroup.Done()

		return handlerErr
	})

	bus.Publish(testEvent{name: "fail"})
	waitWithTimeout(t, &waitGroup)
	require.NoError(t, bus.Shutdown(context.Background()))

	require.Equal(t, int32(1), errCount.Load())
	require.Equal(t, handlerErr.Error(), reported.Load())
}

// A panicking handler must not take the process down with it: the worker recovers
// and the bus keeps delivering.
func TestBus_SurvivesAPanickingHandler(t *testing.T) {
	t.Parallel()

	bus := newTestBus(t, event.NewDefaultOptions())

	var waitGroup sync.WaitGroup
	waitGroup.Add(2)

	bus.Subscribe("boom", func(_ context.Context, _ event.Event) error {
		defer waitGroup.Done()
		panic("handler exploded")
	})
	bus.Subscribe("boom", func(_ context.Context, _ event.Event) error {
		defer waitGroup.Done()

		return nil
	})

	bus.Publish(testEvent{name: "boom"})
	waitWithTimeout(t, &waitGroup)
}

// Publishing is non-blocking by design: when a subscriber's buffer is full the
// event is dropped and reported, rather than stalling the request that published it.
func TestBus_DroppedEventCallback(t *testing.T) {
	t.Parallel()

	var dropped atomic.Int32
	release := make(chan struct{})

	bus := newTestBus(t, event.Options{
		OnDropped:         func(_ event.Event) { dropped.Add(1) },
		OnError:           func(_ event.Event, _ error) {},
		ChannelBuffer:     1,
		WorkerConcurrency: 1,
		DefaultTimeout:    10 * time.Second,
	})

	bus.Subscribe("flood", func(_ context.Context, _ event.Event) error {
		<-release

		return nil
	})

	for range 10 {
		bus.Publish(testEvent{name: "flood"})
	}
	close(release)

	require.Positive(t, dropped.Load(), "a full buffer must drop rather than block the publisher")
}

func TestBus_ShutdownWaitsForInFlightHandlers(t *testing.T) {
	t.Parallel()

	bus, err := event.NewBus(event.NewDefaultOptions())
	require.NoError(t, err)

	started := make(chan struct{})
	var finished atomic.Bool
	bus.Subscribe("slow", func(_ context.Context, _ event.Event) error {
		close(started)
		time.Sleep(150 * time.Millisecond)
		finished.Store(true)

		return nil
	})

	bus.Publish(testEvent{name: "slow"})
	<-started

	require.NoError(t, bus.Shutdown(context.Background()))
	require.True(t, finished.Load(), "shutdown returned before the handler finished")
}

// Shutdown has to honour its context. The source's version waited forever, so a
// handler that never returned held the whole process open past its termination
// grace period and it got killed instead of stopping.
func TestBus_ShutdownHonoursItsContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		newContext func(t *testing.T) (context.Context, context.CancelFunc)
		wantErrIs  error
		name       string
		wantErrMsg string
	}{
		{
			name: "test shutdown should give up when the deadline passes",
			newContext: func(t *testing.T) (context.Context, context.CancelFunc) {
				t.Helper()

				return context.WithTimeout(context.Background(), 20*time.Millisecond)
			},
			wantErrIs:  context.DeadlineExceeded,
			wantErrMsg: "event bus shutdown failed: context deadline exceeded",
		},
		{
			name: "test shutdown should give up when the context is cancelled",
			newContext: func(t *testing.T) (context.Context, context.CancelFunc) {
				t.Helper()
				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				return ctx, func() {}
			},
			wantErrIs:  context.Canceled,
			wantErrMsg: "event bus shutdown failed: context canceled",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			bus, err := event.NewBus(event.NewDefaultOptions())
			require.NoError(t, err)

			started := make(chan struct{})
			release := make(chan struct{})
			defer close(release)

			bus.Subscribe("stuck", func(_ context.Context, _ event.Event) error {
				close(started)
				<-release

				return nil
			})

			bus.Publish(testEvent{name: "stuck"})
			<-started

			ctx, cancel := tt.newContext(t)
			defer cancel()

			err = bus.Shutdown(ctx)

			require.Error(t, err)
			require.ErrorIs(t, err, tt.wantErrIs)
			require.Equal(t, tt.wantErrMsg, err.Error())
		})
	}
}

// Subscribing to a bus that is already down is a no-op, not a panic and not a
// subscription nobody will ever serve: shutdown order between components is not
// something a caller can always control.
func TestBus_SubscribeAfterShutdown(t *testing.T) {
	t.Parallel()

	bus, err := event.NewBus(event.NewDefaultOptions())
	require.NoError(t, err)
	require.NoError(t, bus.Shutdown(context.Background()))

	var called atomic.Bool
	var cancel event.CancelFunc
	require.NotPanics(t, func() {
		cancel = bus.Subscribe("late", func(_ context.Context, _ event.Event) error {
			called.Store(true)

			return nil
		})
	})

	require.NotNil(t, cancel)
	require.NotPanics(t, func() { cancel() })
	require.NotPanics(t, func() { bus.Publish(testEvent{name: "late"}) })

	time.Sleep(50 * time.Millisecond)

	require.False(t, called.Load())
}

// Delivery happens under the read lock precisely because cancelling closes the
// subscription channel: publishing and cancelling at the same time used to send
// on a closed channel, which panics. Run with -race.
func TestBus_ConcurrentPublishSubscribeAndCancel(t *testing.T) {
	t.Parallel()

	const (
		publishers       = 20
		eventsPerRoutine = 50
		churningRoutines = 10
	)

	bus := newTestBus(t, event.NewDefaultOptions())

	for range 10 {
		bus.Subscribe("stress", func(_ context.Context, _ event.Event) error {
			return nil
		})
	}

	var waitGroup sync.WaitGroup
	for range publishers {
		waitGroup.Go(func() {
			for range eventsPerRoutine {
				bus.Publish(testEvent{name: "stress"})
			}
		})
	}
	for range churningRoutines {
		waitGroup.Go(func() {
			for range eventsPerRoutine {
				cancel := bus.Subscribe("stress", func(_ context.Context, _ event.Event) error {
					return nil
				})
				cancel()
			}
		})
	}

	require.NotPanics(t, waitGroup.Wait)
}
