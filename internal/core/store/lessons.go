package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// LessonProposal is a HITL-gated postmortem (docs/33-evolution-v0.md §5).
type LessonProposal struct {
	ID        string
	CreatedAt time.Time
	RunID     string
	TaskID    string
	Kind      string
	Title     string
	Body      string
	Status    string
}

func (s *Store) InsertLesson(ctx context.Context, rec LessonProposal) error {
	if rec.CreatedAt.IsZero() {
		rec.CreatedAt = time.Now().UTC()
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO lesson_proposals (id, created_at, run_id, task_id, kind, title, body, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		rec.ID, rec.CreatedAt.UTC().Format(time.RFC3339Nano), rec.RunID, rec.TaskID,
		rec.Kind, rec.Title, rec.Body, rec.Status)
	return err
}

func (s *Store) GetLesson(ctx context.Context, id string) (LessonProposal, error) {
	var rec LessonProposal
	var created string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, created_at, run_id, task_id, kind, title, body, status
		FROM lesson_proposals WHERE id = ?`, id).Scan(
		&rec.ID, &created, &rec.RunID, &rec.TaskID, &rec.Kind, &rec.Title, &rec.Body, &rec.Status)
	if err != nil {
		return LessonProposal{}, err
	}
	rec.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
	return rec, nil
}

func (s *Store) ListLessons(ctx context.Context, status string, limit int) ([]LessonProposal, error) {
	if limit <= 0 {
		limit = 50
	}
	var (
		rows *sql.Rows
		err  error
	)
	if status != "" {
		rows, err = s.db.QueryContext(ctx, `
			SELECT id, created_at, run_id, task_id, kind, title, body, status
			FROM lesson_proposals WHERE status = ?
			ORDER BY created_at DESC LIMIT ?`, status, limit)
	} else {
		rows, err = s.db.QueryContext(ctx, `
			SELECT id, created_at, run_id, task_id, kind, title, body, status
			FROM lesson_proposals
			ORDER BY created_at DESC LIMIT ?`, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]LessonProposal, 0)
	for rows.Next() {
		var rec LessonProposal
		var created string
		if err := rows.Scan(&rec.ID, &created, &rec.RunID, &rec.TaskID, &rec.Kind, &rec.Title, &rec.Body, &rec.Status); err != nil {
			return nil, err
		}
		rec.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
		out = append(out, rec)
	}
	return out, rows.Err()
}

func (s *Store) UpdateLessonStatus(ctx context.Context, id, status string) error {
	res, err := s.db.ExecContext(ctx, `UPDATE lesson_proposals SET status = ? WHERE id = ?`, status, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("lesson not found: %s", id)
	}
	return nil
}
