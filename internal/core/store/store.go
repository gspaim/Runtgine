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
	StatusAccepted        Status = "accepted"
	StatusRejected        Status = "rejected"
	StatusPlanned         Status = "planned"
	StatusRunning         Status = "running"
	StatusSucceeded       Status = "succeeded"
	StatusFailed          Status = "failed"
	StatusCancelled       Status = "cancelled"
	StatusWaitingApproval Status = "waiting_approval"
)

type Run struct {
	RunID             string
	TaskID            string
	ParentRunID       string
	Status            Status
	ErrorJSON         string
	PendingStepID     string
	PendingCapability string
	PendingPlayer     string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type RunRecord struct {
	Run
	TaskJSON json.RawMessage
}

type Subtask struct {
	SubtaskID           string `json:"subtask_id"`
	ParentRunID         string `json:"parent_run_id"`
	TaskID              string `json:"task_id"`
	Summary             string `json:"summary"`
	SuggestedCapability string `json:"suggested_capability"`
	Notes               string `json:"notes,omitempty"`
	ChildRunID          string `json:"child_run_id,omitempty"`
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
CREATE TABLE IF NOT EXISTS graph_nodes (
  kind TEXT NOT NULL,
  id   TEXT NOT NULL,
  attrs_json TEXT NOT NULL DEFAULT '{}',
  updated_at TEXT NOT NULL,
  PRIMARY KEY (kind, id)
);
CREATE TABLE IF NOT EXISTS graph_edges (
  kind TEXT NOT NULL,
  from_kind TEXT NOT NULL,
  from_id TEXT NOT NULL,
  to_kind TEXT NOT NULL,
  to_id TEXT NOT NULL,
  attrs_json TEXT NOT NULL DEFAULT '{}',
  updated_at TEXT NOT NULL,
  PRIMARY KEY (kind, from_kind, from_id, to_kind, to_id)
);
CREATE INDEX IF NOT EXISTS idx_graph_edges_from ON graph_edges(from_kind, from_id);
CREATE INDEX IF NOT EXISTS idx_graph_edges_to ON graph_edges(to_kind, to_id);
CREATE TABLE IF NOT EXISTS resource_claims (
  run_id TEXT NOT NULL,
  kind TEXT NOT NULL,
  key TEXT NOT NULL,
  step_id TEXT NOT NULL DEFAULT '',
  acquired_at TEXT NOT NULL,
  released_at TEXT
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_resource_claims_active
  ON resource_claims(kind, key) WHERE released_at IS NULL;
`)
	if err != nil {
		return err
	}
	// Existing DBs from slice 1 may lack parent_run_id.
	_, _ = s.db.Exec(`ALTER TABLE runs ADD COLUMN parent_run_id TEXT NOT NULL DEFAULT ''`)
	_, _ = s.db.Exec(`ALTER TABLE runs ADD COLUMN pending_step_id TEXT NOT NULL DEFAULT ''`)
	_, _ = s.db.Exec(`ALTER TABLE runs ADD COLUMN pending_capability TEXT NOT NULL DEFAULT ''`)
	_, _ = s.db.Exec(`ALTER TABLE runs ADD COLUMN pending_player TEXT NOT NULL DEFAULT ''`)
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

func (s *Store) SetPendingApproval(ctx context.Context, runID, stepID, capability, player string) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := s.db.ExecContext(ctx, `
UPDATE runs SET status=?, pending_step_id=?, pending_capability=?, pending_player=?, error_json='', updated_at=?
WHERE run_id=?`,
		string(StatusWaitingApproval), stepID, capability, player, now, runID)
	return err
}

func (s *Store) ClearPendingApproval(ctx context.Context, runID string) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := s.db.ExecContext(ctx, `
UPDATE runs SET pending_step_id='', pending_capability='', pending_player='', updated_at=?
WHERE run_id=?`, now, runID)
	return err
}

func (s *Store) GetRun(ctx context.Context, runID string) (Run, []byte, error) {
	var r Run
	var status string
	var created, updated string
	var taskJSON string
	err := s.db.QueryRowContext(ctx, `
SELECT run_id, task_id, parent_run_id, status, error_json, task_json,
       pending_step_id, pending_capability, pending_player, created_at, updated_at
FROM runs WHERE run_id=?`, runID).Scan(
		&r.RunID, &r.TaskID, &r.ParentRunID, &status, &r.ErrorJSON, &taskJSON,
		&r.PendingStepID, &r.PendingCapability, &r.PendingPlayer, &created, &updated)
	if err != nil {
		return r, nil, err
	}
	r.Status = Status(status)
	r.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
	r.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updated)
	return r, []byte(taskJSON), nil
}

func (s *Store) ListRuns(ctx context.Context, limit int) ([]RunRecord, error) {
	if limit < 1 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT run_id, task_id, parent_run_id, status, error_json, task_json,
       pending_step_id, pending_capability, pending_player, created_at, updated_at
FROM runs ORDER BY created_at DESC, run_id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []RunRecord
	for rows.Next() {
		var rec RunRecord
		var status, taskJSON, created, updated string
		if err := rows.Scan(
			&rec.RunID, &rec.TaskID, &rec.ParentRunID, &status,
			&rec.ErrorJSON, &taskJSON,
			&rec.PendingStepID, &rec.PendingCapability, &rec.PendingPlayer,
			&created, &updated,
		); err != nil {
			return nil, err
		}
		rec.Status = Status(status)
		rec.TaskJSON = json.RawMessage(taskJSON)
		rec.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
		rec.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updated)
		out = append(out, rec)
	}
	return out, rows.Err()
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

func (s *Store) ListRecentEvents(ctx context.Context, limit int) ([]event.Event, error) {
	if limit < 1 {
		limit = 200
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT event_id, run_id, task_id, type, ts, step_id, payload_json
FROM events ORDER BY ts DESC, event_id DESC LIMIT ?`, limit)
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
		if err := rows.Scan(
			&e.EventID, &e.RunID, &e.TaskID, &e.Type, &ts, &step, &payload,
		); err != nil {
			return nil, err
		}
		e.SchemaVersion = event.SchemaVersion
		e.TS, _ = time.Parse(time.RFC3339Nano, ts)
		if step.Valid {
			value := step.String
			e.StepID = &value
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

type GraphNode struct {
	Kind      string
	ID        string
	AttrsJSON string
	UpdatedAt time.Time
}

type GraphEdge struct {
	Kind      string
	FromKind  string
	FromID    string
	ToKind    string
	ToID      string
	AttrsJSON string
	UpdatedAt time.Time
}

func (s *Store) UpsertGraphNode(ctx context.Context, kind, id, attrsJSON string) error {
	if attrsJSON == "" {
		attrsJSON = "{}"
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := s.db.ExecContext(ctx, `
INSERT INTO graph_nodes(kind, id, attrs_json, updated_at)
VALUES(?,?,?,?)
ON CONFLICT(kind, id) DO UPDATE SET
  attrs_json=excluded.attrs_json,
  updated_at=excluded.updated_at`,
		kind, id, attrsJSON, now)
	return err
}

func (s *Store) UpsertGraphEdge(ctx context.Context, kind, fromKind, fromID, toKind, toID, attrsJSON string) error {
	if attrsJSON == "" {
		attrsJSON = "{}"
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := s.db.ExecContext(ctx, `
INSERT INTO graph_edges(kind, from_kind, from_id, to_kind, to_id, attrs_json, updated_at)
VALUES(?,?,?,?,?,?,?)
ON CONFLICT(kind, from_kind, from_id, to_kind, to_id) DO UPDATE SET
  attrs_json=excluded.attrs_json,
  updated_at=excluded.updated_at`,
		kind, fromKind, fromID, toKind, toID, attrsJSON, now)
	return err
}

func (s *Store) GetGraphNode(ctx context.Context, kind, id string) (GraphNode, error) {
	var n GraphNode
	var updated string
	err := s.db.QueryRowContext(ctx, `
SELECT kind, id, attrs_json, updated_at FROM graph_nodes WHERE kind=? AND id=?`,
		kind, id).Scan(&n.Kind, &n.ID, &n.AttrsJSON, &updated)
	if err != nil {
		return n, err
	}
	n.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updated)
	return n, nil
}

func (s *Store) ListGraphNodes(ctx context.Context) ([]GraphNode, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT kind, id, attrs_json, updated_at FROM graph_nodes ORDER BY kind ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []GraphNode
	for rows.Next() {
		var n GraphNode
		var updated string
		if err := rows.Scan(&n.Kind, &n.ID, &n.AttrsJSON, &updated); err != nil {
			return nil, err
		}
		n.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updated)
		out = append(out, n)
	}
	return out, rows.Err()
}

func (s *Store) ListGraphEdges(ctx context.Context) ([]GraphEdge, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT kind, from_kind, from_id, to_kind, to_id, attrs_json, updated_at
FROM graph_edges
ORDER BY kind ASC, from_kind ASC, from_id ASC, to_kind ASC, to_id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []GraphEdge
	for rows.Next() {
		var e GraphEdge
		var updated string
		if err := rows.Scan(&e.Kind, &e.FromKind, &e.FromID, &e.ToKind, &e.ToID, &e.AttrsJSON, &updated); err != nil {
			return nil, err
		}
		e.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updated)
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *Store) QueryGraphNeighbors(ctx context.Context, kind, id, edgeKind, direction string) ([]GraphNode, error) {
	var q string
	var args []any
	switch direction {
	case "in":
		q = `
SELECT n.kind, n.id, n.attrs_json, n.updated_at
FROM graph_edges e
JOIN graph_nodes n ON n.kind=e.from_kind AND n.id=e.from_id
WHERE e.to_kind=? AND e.to_id=? AND (?='' OR e.kind=?)
ORDER BY n.kind ASC, n.id ASC`
		args = []any{kind, id, edgeKind, edgeKind}
	default: // out
		q = `
SELECT n.kind, n.id, n.attrs_json, n.updated_at
FROM graph_edges e
JOIN graph_nodes n ON n.kind=e.to_kind AND n.id=e.to_id
WHERE e.from_kind=? AND e.from_id=? AND (?='' OR e.kind=?)
ORDER BY n.kind ASC, n.id ASC`
		args = []any{kind, id, edgeKind, edgeKind}
	}
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []GraphNode
	for rows.Next() {
		var n GraphNode
		var updated string
		if err := rows.Scan(&n.Kind, &n.ID, &n.AttrsJSON, &updated); err != nil {
			return nil, err
		}
		n.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updated)
		out = append(out, n)
	}
	return out, rows.Err()
}

type ResourceClaim struct {
	RunID  string
	Kind   string
	Key    string
	StepID string
}

// TryAcquireClaim inserts an exclusive claim. A non-empty holderRunID means
// conflict with another run (err is nil). Same-run re-acquire of the same
// kind+key is idempotent.
func (s *Store) TryAcquireClaim(ctx context.Context, runID, stepID, kind, key string, overlaps func(aKind, aKey, bKind, bKey string) bool) (holderRunID string, err error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	rows, err := tx.QueryContext(ctx, `
SELECT run_id, kind, key FROM resource_claims WHERE released_at IS NULL`)
	if err != nil {
		return "", err
	}
	var active []ResourceClaim
	for rows.Next() {
		var row ResourceClaim
		if err := rows.Scan(&row.RunID, &row.Kind, &row.Key); err != nil {
			_ = rows.Close()
			return "", err
		}
		active = append(active, row)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return "", err
	}
	if err := rows.Close(); err != nil {
		return "", err
	}

	for _, row := range active {
		if row.RunID == runID && row.Kind == kind && row.Key == key {
			if err := tx.Commit(); err != nil {
				return "", err
			}
			committed = true
			return "", nil
		}
		if row.RunID != runID && overlaps != nil && overlaps(kind, key, row.Kind, row.Key) {
			if err := tx.Commit(); err != nil {
				return "", err
			}
			committed = true
			return row.RunID, nil
		}
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err = tx.ExecContext(ctx, `
INSERT INTO resource_claims(run_id, kind, key, step_id, acquired_at, released_at)
VALUES(?,?,?,?,?,NULL)`, runID, kind, key, stepID, now)
	if err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	committed = true
	return "", nil
}

func (s *Store) ListActiveClaims(ctx context.Context) ([]ResourceClaim, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT run_id, kind, key, step_id FROM resource_claims WHERE released_at IS NULL`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ResourceClaim
	for rows.Next() {
		var c ResourceClaim
		if err := rows.Scan(&c.RunID, &c.Kind, &c.Key, &c.StepID); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) ReleaseClaimsForRun(ctx context.Context, runID string) ([]ResourceClaim, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	rows, err := tx.QueryContext(ctx, `
SELECT run_id, kind, key, step_id FROM resource_claims
WHERE run_id=? AND released_at IS NULL`, runID)
	if err != nil {
		return nil, err
	}
	var held []ResourceClaim
	for rows.Next() {
		var c ResourceClaim
		if err := rows.Scan(&c.RunID, &c.Kind, &c.Key, &c.StepID); err != nil {
			_ = rows.Close()
			return nil, err
		}
		held = append(held, c)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := tx.ExecContext(ctx, `
UPDATE resource_claims SET released_at=? WHERE run_id=? AND released_at IS NULL`, now, runID); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	committed = true
	return held, nil
}

func (s *Store) SweepOrphanClaims(ctx context.Context) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	res, err := s.db.ExecContext(ctx, `
UPDATE resource_claims SET released_at=?
WHERE released_at IS NULL
  AND run_id NOT IN (
    SELECT run_id FROM runs WHERE status IN ('running', 'waiting_approval')
  )`, now)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
