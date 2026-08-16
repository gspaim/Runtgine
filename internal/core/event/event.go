package event

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

const SchemaVersion = "0.1.0"

type Event struct {
	SchemaVersion string         `json:"schema_version"`
	EventID       string         `json:"event_id"`
	Type          string         `json:"type"`
	TS            time.Time      `json:"ts"`
	RunID         string         `json:"run_id,omitempty"`
	TaskID        string         `json:"task_id,omitempty"`
	StepID        *string        `json:"step_id"`
	Payload       map[string]any `json:"payload"`
}

const (
	TypeTaskAccepted  = "task.accepted"
	TypeTaskRejected  = "task.rejected"
	TypeRunPlanned    = "run.planned"
	TypeRunStarted    = "run.started"
	TypeStepStarted   = "step.started"
	TypeStepSucceeded = "step.succeeded"
	TypeStepFailed    = "step.failed"
	TypeRunSucceeded  = "run.succeeded"
	TypeRunFailed     = "run.failed"
	TypeRunCancelled  = "run.cancelled"
)

func New(typ, runID, taskID string, stepID *string, payload map[string]any) (Event, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return Event{}, err
	}
	if payload == nil {
		payload = map[string]any{}
	}
	return Event{
		SchemaVersion: SchemaVersion,
		EventID:       id.String(),
		Type:          typ,
		TS:            time.Now().UTC(),
		RunID:         runID,
		TaskID:        taskID,
		StepID:        stepID,
		Payload:       payload,
	}, nil
}

// Bus is an in-process pub/sub (G-36: interface-ready for future NATS).
type Bus interface {
	Publish(e Event)
	Subscribe(buffer int) (ch <-chan Event, unsubscribe func())
}

type MemoryBus struct {
	mu   sync.RWMutex
	subs map[uint64]chan Event
	next uint64
}

func NewMemoryBus() *MemoryBus {
	return &MemoryBus{subs: map[uint64]chan Event{}}
}

func (b *MemoryBus) Publish(e Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.subs {
		select {
		case ch <- e:
		default:
			// drop if subscriber is slow (MVP)
		}
	}
}

func (b *MemoryBus) Subscribe(buffer int) (<-chan Event, func()) {
	if buffer < 1 {
		buffer = 64
	}
	ch := make(chan Event, buffer)
	b.mu.Lock()
	id := b.next
	b.next++
	b.subs[id] = ch
	b.mu.Unlock()
	unsub := func() {
		b.mu.Lock()
		if c, ok := b.subs[id]; ok {
			delete(b.subs, id)
			close(c)
		}
		b.mu.Unlock()
	}
	return ch, unsub
}
