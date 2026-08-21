package lessons

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gspaim/Runtgine/internal/config"
	"github.com/gspaim/Runtgine/internal/core/event"
	"github.com/gspaim/Runtgine/internal/core/memory"
	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/gspaim/Runtgine/internal/core/store"
)

const (
	StatusPending  = "pending"
	StatusApproved = "approved"
	StatusRejected = "rejected"
)

type Proposal struct {
	ID        string    `json:"id"`
	RunID     string    `json:"run_id"`
	TaskID    string    `json:"task_id"`
	Kind      string    `json:"kind"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type Service struct {
	Store  *store.Store
	Memory *memory.Service
	Log    *slog.Logger
}

func New(st *store.Store, mem *memory.Service, log *slog.Logger) *Service {
	if log == nil {
		log = slog.Default()
	}
	return &Service{Store: st, Memory: mem, Log: log}
}

func (s *Service) CaptureEnabled(cfg config.Config) bool {
	return cfg.Lessons.Capture == config.LessonsCaptureFailures
}

func (s *Service) OnRunFailed(ctx context.Context, runID, taskID, summary, errText string, events []event.Event) error {
	title := "Failure: " + summary
	if title == "Failure: " {
		title = "Failure: " + runID
	}
	if len(title) > 200 {
		title = title[:200]
	}
	var b strings.Builder
	fmt.Fprintf(&b, "run %s failed: %s\n", runID, errText)
	for i, e := range events {
		if i > 12 {
			break
		}
		fmt.Fprintf(&b, "- %s %s\n", e.Type, e.TS.UTC().Format(time.RFC3339))
	}
	id, err := uuid.NewV7()
	if err != nil {
		return err
	}
	p := store.LessonProposal{
		ID:        id.String(),
		CreatedAt: time.Now().UTC(),
		RunID:     runID,
		TaskID:    taskID,
		Kind:      memory.KindFailure,
		Title:     title,
		Body:      b.String(),
		Status:    StatusPending,
	}
	return s.Store.InsertLesson(ctx, p)
}

func (s *Service) List(ctx context.Context, status string) ([]Proposal, error) {
	rows, err := s.Store.ListLessons(ctx, status, 50)
	if err != nil {
		return nil, err
	}
	out := make([]Proposal, 0, len(rows))
	for _, r := range rows {
		out = append(out, Proposal{
			ID: r.ID, RunID: r.RunID, TaskID: r.TaskID, Kind: r.Kind,
			Title: r.Title, Body: r.Body, Status: r.Status, CreatedAt: r.CreatedAt,
		})
	}
	return out, nil
}

func (s *Service) Approve(ctx context.Context, id string) (memory.Episode, error) {
	row, err := s.Store.GetLesson(ctx, id)
	if err == sql.ErrNoRows {
		return memory.Episode{}, result.Runtime(result.CodeNotFound, "lesson not found", false, nil)
	}
	if err != nil {
		return memory.Episode{}, err
	}
	if row.Status != StatusPending {
		return memory.Episode{}, result.Runtime(result.CodeInternal, "lesson is not pending", false, nil)
	}
	if s.Memory == nil {
		return memory.Episode{}, result.Runtime(result.CodeInternal, "memory provider missing", false, nil)
	}
	ep, err := s.Memory.Record(ctx, memory.EpisodeInput{
		Kind:   row.Kind,
		Title:  row.Title,
		Body:   row.Body,
		RunID:  row.RunID,
		TaskID: row.TaskID,
	})
	if err != nil {
		return memory.Episode{}, err
	}
	if err := s.Store.UpdateLessonStatus(ctx, id, StatusApproved); err != nil {
		return memory.Episode{}, err
	}
	return ep, nil
}

func (s *Service) Reject(ctx context.Context, id string) error {
	row, err := s.Store.GetLesson(ctx, id)
	if err == sql.ErrNoRows {
		return result.Runtime(result.CodeNotFound, "lesson not found", false, nil)
	}
	if err != nil {
		return err
	}
	if row.Status != StatusPending {
		return result.Runtime(result.CodeInternal, "lesson is not pending", false, nil)
	}
	return s.Store.UpdateLessonStatus(ctx, id, StatusRejected)
}
