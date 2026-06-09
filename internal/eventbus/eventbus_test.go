package eventbus

import (
	"sync"
	"testing"
	"time"
)

// testEvent is a simple event for testing.
type testEvent struct {
	payload string
}

func (e testEvent) Type() string { return "test" }

type otherEvent struct {
	value int
}

func (e otherEvent) Type() string { return "other" }

func TestSubscribeAndPublish(t *testing.T) {
	bus := New()

	var mu sync.Mutex
	var received []string
	done := make(chan struct{})

	bus.Subscribe("test", func(e Event) {
		mu.Lock()
		received = append(received, e.(testEvent).payload)
		mu.Unlock()
		close(done)
	})

	bus.Publish(testEvent{payload: "hello"})

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for handler")
	}

	mu.Lock()
	if len(received) != 1 || received[0] != "hello" {
		t.Errorf("expected [hello], got %v", received)
	}
	mu.Unlock()
}

func TestMultipleSubscribers(t *testing.T) {
	bus := New()

	count := 0
	done := make(chan struct{}, 2)

	for i := 0; i < 2; i++ {
		bus.Subscribe("test", func(e Event) {
			count++
			done <- struct{}{}
		})
	}

	bus.Publish(testEvent{payload: "multi"})

	for i := 0; i < 2; i++ {
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for handler")
		}
	}
}

func TestUnsubscribe(t *testing.T) {
	bus := New()

	called := false
	unsub := bus.Subscribe("test", func(e Event) {
		called = true
	})

	unsub()
	bus.Publish(testEvent{payload: "after-unsub"})

	// Give the goroutine a moment to run (it shouldn't)
	time.Sleep(50 * time.Millisecond)

	if called {
		t.Error("handler should not be called after unsubscribe")
	}
}

func TestDifferentEventTypes(t *testing.T) {
	bus := New()

	testCalled := false
	otherCalled := false

	bus.Subscribe("test", func(e Event) {
		testCalled = true
	})
	bus.Subscribe("other", func(e Event) {
		otherCalled = true
	})

	bus.Publish(testEvent{payload: "only-test"})

	time.Sleep(50 * time.Millisecond)

	if !testCalled {
		t.Error("test handler should be called")
	}
	if otherCalled {
		t.Error("other handler should NOT be called")
	}
}

func TestHasSubscribers(t *testing.T) {
	bus := New()

	if bus.HasSubscribers("test") {
		t.Error("expected no subscribers initially")
	}

	bus.Subscribe("test", func(e Event) {})

	if !bus.HasSubscribers("test") {
		t.Error("expected subscribers after subscribe")
	}
}

func TestNoSubscribersDoesNotPanic(t *testing.T) {
	bus := New()

	// Publishing with no subscribers should not panic.
	bus.Publish(testEvent{payload: "no-subscribers"})
}
