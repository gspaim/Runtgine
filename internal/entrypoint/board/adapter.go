package board

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/gspaim/Runtgine/internal/core/api"
	"github.com/gspaim/Runtgine/internal/core/event"
	corepipe "github.com/gspaim/Runtgine/internal/core/pipeline"
	"github.com/gspaim/Runtgine/internal/core/store"
)

type Adapter struct {
	Core   *api.Core
	GitHub *GitHub
	Log    *slog.Logger
	Seen   map[string]string // card id -> run_id
}

func NewAdapter(core *api.Core, gh *GitHub, log *slog.Logger) *Adapter {
	if log == nil {
		log = slog.Default()
	}
	return &Adapter{Core: core, GitHub: gh, Log: log, Seen: map[string]string{}}
}

type PollOptions struct {
	Repo          string
	Label         string
	ProjectOwner  string
	ProjectNumber int
	OwnerIsOrg    bool
	Wait          bool
}

func (a *Adapter) PollOnce(ctx context.Context, opt PollOptions) ([]string, error) {
	cards, err := a.fetch(ctx, opt)
	if err != nil {
		return nil, err
	}
	var runIDs []string
	for _, card := range cards {
		if _, ok := a.Seen[card.ID]; ok {
			continue
		}
		t, err := corepipe.NewTaskIR(card.Title, card.Body, "board", card.ID)
		if err != nil {
			return runIDs, err
		}
		runID, err := a.Core.SubmitTask(ctx, t)
		if err != nil {
			a.Log.Error("submit from board", "card", card.ID, "err", err)
			continue
		}
		a.Seen[card.ID] = runID
		runIDs = append(runIDs, runID)
		_ = a.writeBack(ctx, card, runID, store.StatusRunning, "")
		if opt.Wait {
			status, errJSON := a.wait(ctx, runID)
			_ = a.writeBack(ctx, card, runID, status, errJSON)
		} else {
			go a.watch(card, runID)
		}
	}
	return runIDs, nil
}

func (a *Adapter) fetch(ctx context.Context, opt PollOptions) ([]Card, error) {
	if opt.ProjectNumber > 0 && opt.ProjectOwner != "" {
		return a.GitHub.ListProjectItems(ctx, opt.ProjectOwner, opt.ProjectNumber, opt.OwnerIsOrg)
	}
	return a.GitHub.ListIssueCards(ctx, opt.Repo, opt.Label)
}

func (a *Adapter) watch(card Card, runID string) {
	status, errJSON := a.wait(context.Background(), runID)
	_ = a.writeBack(context.Background(), card, runID, status, errJSON)
}

func (a *Adapter) wait(ctx context.Context, runID string) (store.Status, string) {
	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return store.StatusCancelled, ctx.Err().Error()
		case <-ticker.C:
			snap, err := a.Core.GetRun(ctx, runID)
			if err != nil {
				continue
			}
			st := store.Status(snap.Status)
			switch st {
			case store.StatusSucceeded, store.StatusFailed, store.StatusCancelled, store.StatusRejected:
				return st, snap.Error
			}
		}
	}
}

func (a *Adapter) writeBack(ctx context.Context, card Card, runID string, status store.Status, errJSON string) error {
	if a.GitHub == nil {
		return nil
	}
	label := "runtgine:" + string(status)
	_ = a.GitHub.AddLabel(ctx, card.Repo, card.Number, label)
	var b strings.Builder
	fmt.Fprintf(&b, "Runtgine run `%s` → **%s**\n", runID, status)
	if errJSON != "" {
		fmt.Fprintf(&b, "\n```\n%s\n```\n", errJSON)
	}
	if status == store.StatusSucceeded || status == store.StatusFailed {
		snap, err := a.Core.GetRun(ctx, runID)
		if err == nil && len(snap.Subtasks) > 0 {
			fmt.Fprintf(&b, "\nSubtasks: %d (source of truth in Core; no board cards created).\n", len(snap.Subtasks))
		}
	}
	return a.GitHub.Comment(ctx, card.Repo, card.Number, b.String())
}

// MapEvent is reserved for future live write-back from the bus.
func (a *Adapter) MapEvent(_ event.Event) {}
