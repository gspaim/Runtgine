package event_test

import (
	"testing"
	"time"

	"github.com/gspaim/Runtgine/internal/core/event"
)

func TestNewFillsEnvelope(t *testing.T) {
	step := "s1"
	e, err := event.New(event.TypeStepStarted, "run-1", "task-1", &step, nil)
	if err != nil {
		t.Fatal(err)
	}
	if e.SchemaVersion != event.SchemaVersion {
		t.Fatalf("schema=%s", e.SchemaVersion)
	}
	if e.EventID == "" {
		t.Fatal("event_id is required")
	}
	if e.Type != event.TypeStepStarted || e.RunID != "run-1" || e.TaskID != "task-1" {
		t.Fatalf("envelope=%+v", e)
	}
	if e.StepID == nil || *e.StepID != "s1" {
		t.Fatalf("step_id=%v", e.StepID)
	}
	if e.TS.IsZero() {
		t.Fatal("ts is required")
	}
	if e.Payload == nil || len(e.Payload) != 0 {
		t.Fatalf("nil payload must become empty map, got %v", e.Payload)
	}
}

func TestPublishSubscribe(t *testing.T) {
	bus := event.NewMemoryBus()
	ch, unsub := bus.Subscribe(16)
	defer unsub()

	e1, _ := event.New(event.TypeRunStarted, "r", "t", nil, nil)
	e2, _ := event.New(event.TypeRunSucceeded, "r", "t", nil, nil)
	bus.Publish(e1)
	bus.Publish(e2)

	for _, want := range []event.Event{e1, e2} {
		select {
		case got := <-ch:
			if got.EventID != want.EventID {
				t.Fatalf("got %s want %s", got.EventID, want.EventID)
			}
		case <-time.After(time.Second):
			t.Fatalf("timed out waiting for %s", want.Type)
		}
	}
}

func TestUnsubscribeClosesAndStopsDelivery(t *testing.T) {
	bus := event.NewMemoryBus()
	ch, unsub := bus.Subscribe(4)
	unsub()

	e, _ := event.New(event.TypeRunStarted, "r", "t", nil, nil)
	bus.Publish(e)

	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("unexpected event after unsubscribe")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("channel must be closed by unsubscribe")
	}
}

func TestPublishDoesNotBlockOnSlowSubscriber(t *testing.T) {
	bus := event.NewMemoryBus()
	ch, unsub := bus.Subscribe(2)
	defer unsub()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 50; i++ {
			e, _ := event.New(event.TypeStepStarted, "r", "t", nil, nil)
			bus.Publish(e)
		}
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("publish blocked on slow subscriber")
	}

	// Buffer holds the first events; the rest were dropped (MVP behavior).
	for i := 0; i < 2; i++ {
		select {
		case <-ch:
		default:
			t.Fatalf("expected %d buffered events", 2-i)
		}
	}
	select {
	case e := <-ch:
		t.Fatalf("unexpected extra event %s", e.Type)
	default:
	}
}

func TestSubscribeSmallBufferUsesDefault(t *testing.T) {
	bus := event.NewMemoryBus()
	ch, unsub := bus.Subscribe(0)
	defer unsub()

	for i := 0; i < 10; i++ {
		e, _ := event.New(event.TypeStepSucceeded, "r", "t", nil, nil)
		bus.Publish(e)
	}
	count := 0
loop:
	for {
		select {
		case <-ch:
			count++
		default:
			break loop
		}
	}
	if count == 0 {
		t.Fatal("default buffer must still deliver events")
	}
}
