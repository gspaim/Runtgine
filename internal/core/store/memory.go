package store

import (
	"context"
	"database/sql"
	"time"
)

// MemoryEpisodeRow is the SQLite projection of a Project Memory episode (G-124).
type MemoryEpisodeRow struct {
	ID          string
	Kind        string
	Validity    string
	Title       string
	Body        string
	CreatedAt   time.Time
	RunID       string
	TaskID      string
	SuccessorID string
}

func (s *Store) InsertMemoryEpisode(ctx context.Context, row MemoryEpisodeRow) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO project_memory(id, kind, validity, title, body, created_at, run_id, task_id, successor_id)
VALUES(?,?,?,?,?,?,?,?,?)`,
		row.ID, row.Kind, row.Validity, row.Title, row.Body,
		row.CreatedAt.UTC().Format(time.RFC3339Nano),
		row.RunID, row.TaskID, row.SuccessorID)
	return err
}

func (s *Store) GetMemoryEpisode(ctx context.Context, id string) (MemoryEpisodeRow, error) {
	var row MemoryEpisodeRow
	var created string
	err := s.db.QueryRowContext(ctx, `
SELECT id, kind, validity, title, body, created_at, run_id, task_id, successor_id
FROM project_memory WHERE id=?`, id).Scan(
		&row.ID, &row.Kind, &row.Validity, &row.Title, &row.Body,
		&created, &row.RunID, &row.TaskID, &row.SuccessorID)
	if err != nil {
		return MemoryEpisodeRow{}, err
	}
	row.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
	return row, nil
}

func (s *Store) ListMemoryEpisodes(ctx context.Context, kind, validity string) ([]MemoryEpisodeRow, error) {
	q := `
SELECT id, kind, validity, title, body, created_at, run_id, task_id, successor_id
FROM project_memory`
	var args []any
	switch {
	case kind != "" && validity != "":
		q += ` WHERE kind=? AND validity=?`
		args = append(args, kind, validity)
	case kind != "":
		q += ` WHERE kind=?`
		args = append(args, kind)
	case validity != "":
		q += ` WHERE validity=?`
		args = append(args, validity)
	}
	q += ` ORDER BY created_at DESC, id ASC`
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMemoryEpisodes(rows)
}

func (s *Store) ListActiveMemoryEpisodes(ctx context.Context) ([]MemoryEpisodeRow, error) {
	return s.ListMemoryEpisodes(ctx, "", "active")
}

func (s *Store) ArchiveMemoryEpisode(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `
UPDATE project_memory SET validity='archived' WHERE id=?`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// SupersedeMemoryEpisode inserts the successor and marks the predecessor in one
// transaction (G-124).
func (s *Store) SupersedeMemoryEpisode(ctx context.Context, oldID string, neu MemoryEpisodeRow) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	var exists string
	if err := tx.QueryRowContext(ctx, `SELECT id FROM project_memory WHERE id=?`, oldID).Scan(&exists); err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
INSERT INTO project_memory(id, kind, validity, title, body, created_at, run_id, task_id, successor_id)
VALUES(?,?,?,?,?,?,?,?,?)`,
		neu.ID, neu.Kind, neu.Validity, neu.Title, neu.Body,
		neu.CreatedAt.UTC().Format(time.RFC3339Nano),
		neu.RunID, neu.TaskID, neu.SuccessorID)
	if err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx, `
UPDATE project_memory SET validity='superseded', successor_id=? WHERE id=?`, neu.ID, oldID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}

func scanMemoryEpisodes(rows *sql.Rows) ([]MemoryEpisodeRow, error) {
	out := []MemoryEpisodeRow{}
	for rows.Next() {
		var row MemoryEpisodeRow
		var created string
		if err := rows.Scan(
			&row.ID, &row.Kind, &row.Validity, &row.Title, &row.Body,
			&created, &row.RunID, &row.TaskID, &row.SuccessorID,
		); err != nil {
			return nil, err
		}
		row.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
		out = append(out, row)
	}
	return out, rows.Err()
}
