package desktop

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/gspaim/Runtgine/internal/core/api"
	"github.com/gspaim/Runtgine/internal/core/event"
	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/core/runner"
	"github.com/gspaim/Runtgine/internal/core/task"
)

const (
	entryPoint = "wails"
	sourceRef  = "intent"
	eventName  = "runtgine:event"
	callTimeout = 8 * time.Second
)

// CoreAPI is the Core surface the desktop Entry Point may call.
// It must not reach Players.
type CoreAPI interface {
	CompileIntent(context.Context, string, string, string) (task.Task, string, error)
	SubmitIntent(context.Context, string, string, string) (string, string, error)
	SubmitTask(context.Context, task.Task) (string, error)
	GetRun(context.Context, string) (api.RunSnapshot, error)
	ListRuns(context.Context, int) ([]api.RunSummary, error)
	Subscribe(int) (<-chan event.Event, func())
	CancelRun(string) error
	ApproveRun(string, string) error
}

// Emitter sends Core events to the Wails frontend.
type Emitter interface {
	Emit(name string, data any)
}

// Service is the Wails v3 service facade over api.Core (G-162).
type Service struct {
	core CoreAPI
	emit Emitter

	mu     sync.Mutex
	unsub  func()
	cancel context.CancelFunc
}

func NewService(core CoreAPI) *Service {
	return &Service{core: core}
}

func (s *Service) SetEmitter(emit Emitter) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.emit = emit
}

func (s *Service) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.unsub != nil {
		return
	}
	ch, unsub := s.core.Subscribe(256)
	ctx, cancel := context.WithCancel(context.Background())
	s.unsub = unsub
	s.cancel = cancel
	go s.forward(ctx, ch)
}

func (s *Service) Stop() {
	s.mu.Lock()
	unsub := s.unsub
	cancel := s.cancel
	s.unsub = nil
	s.cancel = nil
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if unsub != nil {
		unsub()
	}
}

func (s *Service) forward(ctx context.Context, ch <-chan event.Event) {
	for {
		select {
		case <-ctx.Done():
			return
		case e, ok := <-ch:
			if !ok {
				return
			}
			s.mu.Lock()
			emit := s.emit
			s.mu.Unlock()
			if emit != nil {
				emit.Emit(eventName, e)
			}
		}
	}
}

type IntentPreview struct {
	Task   json.RawMessage `json:"task"`
	Method string          `json:"method"`
}

type IntentSubmit struct {
	RunID  string          `json:"run_id"`
	Method string          `json:"method"`
	Task   json.RawMessage `json:"task"`
}

func (s *Service) CompileIntent(text string) (IntentPreview, error) {
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()
	tk, method, err := s.core.CompileIntent(ctx, text, entryPoint, sourceRef)
	if err != nil {
		return IntentPreview{}, mapErr(err)
	}
	raw, err := json.Marshal(tk)
	if err != nil {
		return IntentPreview{}, mapErr(err)
	}
	return IntentPreview{Task: raw, Method: method}, nil
}

func (s *Service) CompileTaskJSON(raw string) (IntentPreview, error) {
	trimmed := strings.TrimSpace(raw)
	if err := task.ValidateDocument([]byte(trimmed)); err != nil {
		return IntentPreview{}, mapErr(err)
	}
	tk, err := task.Parse([]byte(trimmed))
	if err != nil {
		return IntentPreview{}, mapErr(err)
	}
	if strings.TrimSpace(tk.Source.EntryPoint) == "" {
		tk.Source.EntryPoint = entryPoint
	}
	pretty, err := json.Marshal(tk)
	if err != nil {
		return IntentPreview{}, mapErr(err)
	}
	return IntentPreview{Task: pretty, Method: "json"}, nil
}

func (s *Service) SubmitIntent(text string) (IntentSubmit, error) {
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()
	runID, method, err := s.core.SubmitIntent(ctx, text, entryPoint, sourceRef)
	if err != nil {
		return IntentSubmit{}, mapErr(err)
	}
	tk, _, _ := s.core.CompileIntent(ctx, text, entryPoint, sourceRef)
	raw, _ := json.Marshal(tk)
	return IntentSubmit{RunID: runID, Method: method, Task: raw}, nil
}

func (s *Service) SubmitTaskJSON(raw string) (IntentSubmit, error) {
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()
	trimmed := strings.TrimSpace(raw)
	if err := task.ValidateDocument([]byte(trimmed)); err != nil {
		return IntentSubmit{}, mapErr(err)
	}
	tk, err := task.Parse([]byte(trimmed))
	if err != nil {
		return IntentSubmit{}, mapErr(err)
	}
	if strings.TrimSpace(tk.Source.EntryPoint) == "" {
		tk.Source.EntryPoint = entryPoint
	}
	pretty, _ := json.Marshal(tk)
	runID, err := s.core.SubmitTask(ctx, tk)
	if err != nil {
		return IntentSubmit{}, mapErr(err)
	}
	return IntentSubmit{RunID: runID, Method: "json", Task: pretty}, nil
}

func (s *Service) GetRun(runID string) (api.RunSnapshot, error) {
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()
	snap, err := s.core.GetRun(ctx, runID)
	if err != nil {
		return api.RunSnapshot{}, mapErr(err)
	}
	return snap, nil
}

func (s *Service) ListRuns(limit int) ([]api.RunSummary, error) {
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()
	if limit <= 0 {
		limit = 200
	}
	rows, err := s.core.ListRuns(ctx, limit)
	if err != nil {
		return nil, mapErr(err)
	}
	return rows, nil
}

func (s *Service) CancelRun(runID string) error {
	return mapErr(s.core.CancelRun(runID))
}

func (s *Service) ApproveRun(runID string) error {
	return mapErr(s.core.ApproveRun(runID, runner.DecisionGrant))
}

func (s *Service) DenyRun(runID string) error {
	return mapErr(s.core.ApproveRun(runID, runner.DecisionDeny))
}

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	var re result.Error
	if errors.As(err, &re) {
		return re
	}
	return result.Validation(result.CodeInvalidInput, err.Error(), nil)
}
