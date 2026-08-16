package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gspaim/Runtgine/internal/core/event"
	_ "modernc.org/sqlite"
)

type Status string

const (
	StatusAccepted  Status = "accepted"
	StatusRejected  Status = "rejected"
	StatusPlanned   Status = "planned"
	StatusRunning   Status = "running"
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
)

type Run struct {
	RunID        string
	TaskID       string
	ParentRunID  string
	Status       Status
	ErrorJSON    string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Subtask struct {
	SubtaskID            string `json:"subtask_id"`
	ParentRunID          string `json:"parent_run_id"`
	TaskID               string `json:"task_id"`
	Summary              string `json:"summary"`
	SuggestedCapability  string `json:"suggested_capability"`
	Notes                string `json:"notes,omitempty"`
	ChildRunID           string `json:"child_run_id,omitempty"`
}

type StepOutput struct {
	StepID     string
	Capability string
	Output     json.RawMessage
}

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS runs (
  run_id TEXT PRIMARY KEY,
  task_id TEXT NOT NULL,
  parent_run_id TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL,
  error_json TEXT NOT NULL DEFAULT '',
  task_json TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS events (
  event_id TEXT PRIMARY KEY,
  run_id TEXT NOT NULL,
  task_id TEXT NOT NULL,
  type TEXT NOT NULL,
  ts TEXT NOT NULL,
  step_id TEXT,
  payload_json TEXT NOT NULL,
  FOREIGN KEY(run_id) REFERENCES runs(run_id)
);
CREATE TABLE IF NOT EXISTS step_outputs (
  run_id TEXT NOT NULL,
  step_id TEXT NOT NULL,
  capability TEXT NOT NULL,
  output_json TEXT NOT NULL,
  PRIMARY KEY(run_id, step_id)
);
CREATE TABLE IF NOT EXISTS subtasks (
  subtask_id TEXT PRIMARY KEY,
  parent_run_id TEXT NOT NULL,
  task_id TEXT NOT NULL,
  summary TEXT NOT NULL,
  suggested_capability TEXT NOT NULL DEFAULT '',
  notes TEXT NOT NULL DEFAULT '',
  child_run_id TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_events_run ON events(run_id);
CREATE INDEX IF NOT EXISTS idx_subtasks_parent ON subtasks(parent_run_id);
`)
	if err != nil {
		return err
	}
	// Existing DBs from slice 1 may lack parent_run_id.
	_, _ = s.db.Exec(`ALTER TABLE runs ADD COLUMN parent_run_id TEXT NOT NULL DEFAULT ''`)
	return nil
}

func (s *Store) InsertRun(ctx context.Context, run Run, taskJSON []byte) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := s.db.ExecContext(ctx, `
INSERT INTO runs(run_id, task_id, parent_run_id, status, error_json, task_json, created_at, updated_at)
VALUES(?,?,?,?,?,?,?,?)`,
		run.RunID, run.TaskID, run.ParentRunID, string(run.Status), "", string(taskJSON), now, now)
	return err
}

func (s *Store) UpdateRunStatus(ctx context.Context, runID string, status Status, errJSON string) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := s.db.ExecContext(ctx, `
UPDATE runs SET status=?, error_json=?, updated_at=? WHERE run_id=?`,
		string(status), errJSON, now, runID)
	return err
}

func (s *Store) GetRun(ctx context.Context, runID string) (Run, []byte, error) {
	var r Run
	var status string
	var created, updated string
	var taskJSON string
	err := s.db.QueryRowContext(ctx, `
SELECT run_id, task_id, parent_run_id, status, error_json, task_json, created_at, updated_at
FROM runs WHERE run_id=?`, runID).Scan(
		&r.RunID, &r.TaskID, &r.ParentRunID, &status, &r.ErrorJSON, &taskJSON, &created, &updated)
	if err != nil {
		return r, nil, err
	}
	r.Status = Status(status)
	r.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
	r.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updated)
	return r, []byte(taskJSON), nil
}

func (s *Store) AppendEvent(ctx context.Context, e event.Event) error {
	payload, err := json.Marshal(e.Payload)
	if err != nil {
		return err
	}
	var step any
	if e.StepID != nil {
		step = *e.StepID
	}
	_, err = s.db.ExecContext(ctx, `
INSERT INTO events(event_id, run_id, task_id, type, ts, step_id, payload_json)
VALUES(?,?,?,?,?,?,?)`,
		e.EventID, e.RunID, e.TaskID, e.Type, e.TS.Format(time.RFC3339Nano), step, string(payload))
	return err
}

func (s *Store) ListEvents(ctx context.Context, runID string) ([]event.Event, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT event_id, run_id, task_id, type, ts, step_id, payload_json
FROM events WHERE run_id=? ORDER BY ts ASC, event_id ASC`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []event.Event
	for rows.Next() {
		var e event.Event
		var ts string
		var step sql.NullString
		var payload string
		if err := rows.Scan(&e.EventID, &e.RunID, &e.TaskID, &e.Type, &ts, &step, &payload); err != nil {
			return nil, err
		}
		e.SchemaVersion = event.SchemaVersion
		e.TS, _ = time.Parse(time.RFC3339Nano, ts)
		if step.Valid {
			s := step.String
			e.StepID = &s
		}
		_ = json.Unmarshal([]byte(payload), &e.Payload)
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *Store) SaveStepOutput(ctx context.Context, runID, stepID, capability string, output json.RawMessage) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO step_outputs(run_id, step_id, capability, output_json)
VALUES(?,?,?,?)
ON CONFLICT(run_id, step_id) DO UPDATE SET capability=excluded.capability, output_json=excluded.output_json`,
		runID, stepID, capability, string(output))
	return err
}

func (s *Store) ListStepOutputs(ctx context.Context, runID string) ([]StepOutput, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT step_id, capability, output_json FROM step_outputs WHERE run_id=?`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []StepOutput
	for rows.Next() {
		var o StepOutput
		var raw string
		if err := rows.Scan(&o.StepID, &o.Capability, &raw); err != nil {
			return nil, err
		}
		o.Output = json.RawMessage(raw)
		out = append(out, o)
	}
	return out, rows.Err()
}

func (s *Store) InsertSubtask(ctx context.Context, st Subtask) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO subtasks(subtask_id, parent_run_id, task_id, summary, suggested_capability, notes, child_run_id)
VALUES(?,?,?,?,?,?,?)`,
		st.SubtaskID, st.ParentRunID, st.TaskID, st.Summary, st.SuggestedCapability, st.Notes, st.ChildRunID)
	return err
}

func (s *Store) SetSubtaskChildRun(ctx context.Context, subtaskID, childRunID string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE subtasks SET child_run_id=? WHERE subtask_id=?`, childRunID, subtaskID)
	return err
}

func (s *Store) ListSubtasks(ctx context.Context, parentRunID string) ([]Subtask, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT subtask_id, parent_run_id, task_id, summary, suggested_capability, notes, child_run_id
FROM subtasks WHERE parent_run_id=?`, parentRunID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Subtask
	for rows.Next() {
		var st Subtask
		if err := rows.Scan(&st.SubtaskID, &st.ParentRunID, &st.TaskID, &st.Summary, &st.SuggestedCapability, &st.Notes, &st.ChildRunID); err != nil {
			return nil, err
		}
		out = append(out, st)
	}
	return out, rows.Err()
}

func (s *Store) ListChildRuns(ctx context.Context, parentRunID string) ([]Run, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT run_id, task_id, parent_run_id, status, error_json, created_at, updated_at
FROM runs WHERE parent_run_id=? ORDER BY created_at ASC`, parentRunID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Run
	for rows.Next() {
		var r Run
		var status, created, updated string
		if err := rows.Scan(&r.RunID, &r.TaskID, &r.ParentRunID, &status, &r.ErrorJSON, &created, &updated); err != nil {
			return nil, err
		}
		r.Status = Status(status)
		r.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
		r.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updated)
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func ErrNotFound(err error) bool {
	return err == sql.ErrNoRows
}

func FormatErr(err error) string {
	if err == nil {
		return ""
	}
	b, _ := json.Marshal(map[string]string{"message": err.Error()})
	return string(b)
}

func MustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%q", err.Error())
	}
	return string(b)
}
