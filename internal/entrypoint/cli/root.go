package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gspaim/Runtgine/internal/config"
	"github.com/gspaim/Runtgine/internal/core/api"
	"github.com/gspaim/Runtgine/internal/core/event"
	"github.com/gspaim/Runtgine/internal/core/memory"
	corepipe "github.com/gspaim/Runtgine/internal/core/pipeline"
	"github.com/gspaim/Runtgine/internal/core/runner"
	"github.com/gspaim/Runtgine/internal/core/store"
	"github.com/gspaim/Runtgine/internal/core/task"
	"github.com/gspaim/Runtgine/internal/entrypoint/board"
	"github.com/gspaim/Runtgine/internal/entrypoint/desktop"
	"github.com/gspaim/Runtgine/internal/entrypoint/httpapi"
	tuientry "github.com/gspaim/Runtgine/internal/entrypoint/tui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func NewRoot() *cobra.Command {
	var (
		workspace string
		verbose   bool
	)
	root := &cobra.Command{
		Use:   "runtgine",
		Short: "Runtgine — deterministic-first execution runtime",
	}
	root.PersistentFlags().StringVar(&workspace, "workspace", "", "workspace root (default: cwd or RUNTGINE_WORKSPACE)")
	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "debug logging")

	root.AddCommand(newRunCmd(&workspace, &verbose))
	root.AddCommand(newIntentCmd(&workspace, &verbose))
	root.AddCommand(newStatusCmd(&workspace, &verbose))
	root.AddCommand(newApproveCmd(&workspace, &verbose, true))
	root.AddCommand(newApproveCmd(&workspace, &verbose, false))
	root.AddCommand(newCancelCmd(&workspace, &verbose))
	root.AddCommand(newGraphCmd(&workspace, &verbose))
	root.AddCommand(newMemoryCmd(&workspace, &verbose))
	root.AddCommand(newBlastCmd(&workspace, &verbose))
	root.AddCommand(newPipelineCmd(&workspace, &verbose))
	root.AddCommand(newBoardCmd(&workspace, &verbose))
	root.AddCommand(newTUICmd(&workspace, &verbose))
	root.AddCommand(newDesktopCmd(&workspace, &verbose))
	root.AddCommand(newLessonsCmd(&workspace, &verbose))
	root.AddCommand(newServeCmd(&workspace, &verbose))
	return root
}

func openCore(workspace string, verbose bool) (*api.Core, error) {
	cfg, err := config.Load(workspace)
	if err != nil {
		return nil, err
	}
	level := slog.LevelInfo
	if verbose || strings.EqualFold(cfg.LogLevel, "debug") {
		level = slog.LevelDebug
	}
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(log)
	return api.Open(cfg, log)
}

func newRunCmd(workspace *string, verbose *bool) *cobra.Command {
	var wait bool
	cmd := &cobra.Command{
		Use:   "run <task.json|task.yaml>",
		Short: "Submit a Task IR file and print run_id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			core, err := openCore(*workspace, *verbose)
			if err != nil {
				return err
			}
			defer core.Close()

			raw, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}
			jsonBytes, err := toJSON(args[0], raw)
			if err != nil {
				return err
			}
			if err := task.ValidateDocument(jsonBytes); err != nil {
				return err
			}
			t, err := task.Parse(jsonBytes)
			if err != nil {
				return err
			}
			if t.Source.EntryPoint == "" {
				t.Source.EntryPoint = "cli"
			}
			if t.Source.Ref == "" {
				t.Source.Ref = filepath.Base(args[0])
			}

			ctx := context.Background()
			var unsub func()
			var events <-chan event.Event
			if wait {
				events, unsub = core.Subscribe(256)
				defer unsub()
			}

			runID, err := core.SubmitTask(ctx, t)
			if err != nil {
				return err
			}
			fmt.Println(runID)

			if !wait {
				return nil
			}
			return waitForRun(ctx, core, runID, events)
		},
	}
	cmd.Flags().BoolVar(&wait, "wait", true, "wait for run completion")
	return cmd
}

func newIntentCmd(workspace *string, verbose *bool) *cobra.Command {
	var wait, dryRun bool
	cmd := &cobra.Command{
		Use:   "intent <text>",
		Short: "Compile natural language into Task IR and optionally submit",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			core, err := openCore(*workspace, *verbose)
			if err != nil {
				return err
			}
			defer core.Close()

			text := strings.Join(args, " ")
			ctx := context.Background()
			if dryRun {
				tk, method, err := core.CompileIntent(ctx, text, "cli", "intent")
				if err != nil {
					return err
				}
				fmt.Fprintf(os.Stderr, "intent_method %s\n", method)
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(tk)
			}

			var unsub func()
			var events <-chan event.Event
			if wait {
				events, unsub = core.Subscribe(256)
				defer unsub()
			}

			runID, method, err := core.SubmitIntent(ctx, text, "cli", "intent")
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "intent_method %s\n", method)
			fmt.Println(runID)
			if !wait {
				return nil
			}
			return waitForRun(ctx, core, runID, events)
		},
	}
	cmd.Flags().BoolVar(&wait, "wait", true, "wait for run completion")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print Task IR JSON without submitting")
	return cmd
}

func waitForRun(ctx context.Context, core *api.Core, runID string, events <-chan event.Event) error {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case e := <-events:
			if e.RunID != runID {
				continue
			}
			fmt.Fprintf(os.Stderr, "event %s\n", e.Type)
			switch e.Type {
			case event.TypeRunSucceeded:
				return nil
			case event.TypeRunFailed, event.TypeRunCancelled:
				snap, err := core.GetRun(ctx, runID)
				if err != nil {
					return fmt.Errorf("%s", e.Type)
				}
				return fmt.Errorf("run %s: %s", snap.Status, snap.Error)
			}
		case <-ticker.C:
			snap, err := core.GetRun(ctx, runID)
			if err != nil {
				continue
			}
			switch store.Status(snap.Status) {
			case store.StatusSucceeded:
				return nil
			case store.StatusFailed, store.StatusCancelled, store.StatusRejected:
				return fmt.Errorf("run %s: %s", snap.Status, snap.Error)
			}
		}
	}
}

func newStatusCmd(workspace *string, verbose *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "status <run_id>",
		Short: "Show run snapshot and events",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			core, err := openCore(*workspace, *verbose)
			if err != nil {
				return err
			}
			defer core.Close()
			snap, err := core.GetRun(context.Background(), args[0])
			if err != nil {
				return err
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(snap)
		},
	}
}

func newApproveCmd(workspace *string, verbose *bool, grant bool) *cobra.Command {
	use, short, decision := "deny", "Deny a run waiting for approval", runner.DecisionDeny
	if grant {
		use, short, decision = "approve", "Approve a run waiting for HITL", runner.DecisionGrant
	}
	return &cobra.Command{
		Use:   use + " <run_id>",
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			core, err := openCore(*workspace, *verbose)
			if err != nil {
				return err
			}
			defer core.Close()
			if err := core.ApproveRun(args[0], decision); err != nil {
				return err
			}
			if grant {
				fmt.Println("granted")
			} else {
				fmt.Println("denied")
			}
			return nil
		},
	}
}

func newCancelCmd(workspace *string, verbose *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "cancel <run_id>",
		Short: "Cancel an active run",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			core, err := openCore(*workspace, *verbose)
			if err != nil {
				return err
			}
			defer core.Close()
			return core.CancelRun(args[0])
		},
	}
}

func newGraphCmd(workspace *string, verbose *bool) *cobra.Command {
	cmd := &cobra.Command{Use: "graph", Short: "Runtime Graph (structural memory)"}
	cmd.AddCommand(&cobra.Command{
		Use:   "snapshot",
		Short: "Print the workspace graph as JSON (no secrets)",
		RunE: func(cmd *cobra.Command, args []string) error {
			core, err := openCore(*workspace, *verbose)
			if err != nil {
				return err
			}
			defer core.Close()
			snap, err := core.GetGraphSnapshot(context.Background())
			if err != nil {
				return err
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(snap)
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "refresh",
		Short: "Re-register players and capabilities from the current process",
		RunE: func(cmd *cobra.Command, args []string) error {
			core, err := openCore(*workspace, *verbose)
			if err != nil {
				return err
			}
			defer core.Close()
			if err := core.RefreshGraph(context.Background()); err != nil {
				return err
			}
			fmt.Println("ok")
			return nil
		},
	})
	return cmd
}

func newMemoryCmd(workspace *string, verbose *bool) *cobra.Command {
	cmd := &cobra.Command{Use: "memory", Short: "Project Memory (episodic, not a Player)"}

	list := &cobra.Command{
		Use:   "list",
		Short: "List episodes as JSON",
		RunE: func(cmd *cobra.Command, args []string) error {
			core, err := openCore(*workspace, *verbose)
			if err != nil {
				return err
			}
			defer core.Close()
			kind, _ := cmd.Flags().GetString("kind")
			validity, _ := cmd.Flags().GetString("validity")
			rows, err := core.MemoryList(context.Background(), memory.Filter{Kind: kind, Validity: validity})
			if err != nil {
				return err
			}
			return writeJSON(os.Stdout, rows)
		},
	}
	list.Flags().String("kind", "", "filter kind")
	list.Flags().String("validity", "", "filter validity")
	cmd.AddCommand(list)

	query := &cobra.Command{
		Use:   "query <text>",
		Short: "Lexical query of active episodes",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			core, err := openCore(*workspace, *verbose)
			if err != nil {
				return err
			}
			defer core.Close()
			limit, _ := cmd.Flags().GetInt("limit")
			hits, err := core.MemoryQuery(context.Background(), args[0], limit)
			if err != nil {
				return err
			}
			return writeJSON(os.Stdout, hits)
		},
	}
	query.Flags().Int("limit", 0, "max hits (default 8)")
	cmd.AddCommand(query)

	record := &cobra.Command{
		Use:   "record",
		Short: "Record an active episode",
		RunE: func(cmd *cobra.Command, args []string) error {
			core, err := openCore(*workspace, *verbose)
			if err != nil {
				return err
			}
			defer core.Close()
			ep, err := core.MemoryRecord(context.Background(), episodeFlags(cmd))
			if err != nil {
				return err
			}
			return writeJSON(os.Stdout, ep)
		},
	}
	bindEpisodeFlags(record)
	cmd.AddCommand(record)

	supersede := &cobra.Command{
		Use:   "supersede <id>",
		Short: "Mark an episode superseded and record a successor",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			core, err := openCore(*workspace, *verbose)
			if err != nil {
				return err
			}
			defer core.Close()
			ep, err := core.MemorySupersede(context.Background(), args[0], episodeFlags(cmd))
			if err != nil {
				return err
			}
			return writeJSON(os.Stdout, ep)
		},
	}
	bindEpisodeFlags(supersede)
	cmd.AddCommand(supersede)

	cmd.AddCommand(&cobra.Command{
		Use:   "archive <id>",
		Short: "Archive an episode (no physical delete)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			core, err := openCore(*workspace, *verbose)
			if err != nil {
				return err
			}
			defer core.Close()
			ep, err := core.MemoryArchive(context.Background(), args[0])
			if err != nil {
				return err
			}
			return writeJSON(os.Stdout, ep)
		},
	})
	return cmd
}

func bindEpisodeFlags(cmd *cobra.Command) {
	cmd.Flags().String("kind", "", "decision|failure|handoff|preference")
	cmd.Flags().String("title", "", "episode title")
	cmd.Flags().String("body", "", "episode body")
	cmd.Flags().String("run-id", "", "optional run id")
	cmd.Flags().String("task-id", "", "optional task id")
	_ = cmd.MarkFlagRequired("kind")
	_ = cmd.MarkFlagRequired("title")
}

func episodeFlags(cmd *cobra.Command) memory.EpisodeInput {
	kind, _ := cmd.Flags().GetString("kind")
	title, _ := cmd.Flags().GetString("title")
	body, _ := cmd.Flags().GetString("body")
	runID, _ := cmd.Flags().GetString("run-id")
	taskID, _ := cmd.Flags().GetString("task-id")
	return memory.EpisodeInput{
		Kind:   kind,
		Title:  title,
		Body:   body,
		RunID:  runID,
		TaskID: taskID,
	}
}

func writeJSON(w *os.File, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func newBlastCmd(workspace *string, verbose *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "blast <task.json|task.yaml>",
		Short: "Print a deterministic impact report for a Task IR (no execute)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			core, err := openCore(*workspace, *verbose)
			if err != nil {
				return err
			}
			defer core.Close()

			raw, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}
			jsonBytes, err := toJSON(args[0], raw)
			if err != nil {
				return err
			}
			if err := task.ValidateDocument(jsonBytes); err != nil {
				return err
			}
			t, err := task.Parse(jsonBytes)
			if err != nil {
				return err
			}
			if t.Source.EntryPoint == "" {
				t.Source.EntryPoint = "cli"
			}
			if t.Source.Ref == "" {
				t.Source.Ref = filepath.Base(args[0])
			}

			rep, err := core.BlastTask(context.Background(), t)
			if err != nil {
				return err
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(rep)
		},
	}
}

func newTUICmd(workspace *string, verbose *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Open Constellation Mission Control",
		RunE: func(cmd *cobra.Command, args []string) error {
			core, err := openCore(*workspace, *verbose)
			if err != nil {
				return err
			}
			defer core.Close()
			return tuientry.Run(core)
		},
	}
}

func newDesktopCmd(workspace *string, verbose *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "desktop",
		Short: "Open the Wails desktop Mission Control (G-159)",
		RunE: func(cmd *cobra.Command, args []string) error {
			core, err := openCore(*workspace, *verbose)
			if err != nil {
				return err
			}
			defer core.Close()
			return desktop.Run(core)
		},
	}
}

func newPipelineCmd(workspace *string, verbose *bool) *cobra.Command {
	cmd := &cobra.Command{Use: "pipeline", Short: "Board analysis pipeline (slice 2)"}
	var summary, notes string
	var wait bool
	run := &cobra.Command{
		Use:   "run",
		Short: "Submit the six-step pipeline Task IR for this workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			if summary == "" {
				summary = "Pipeline analysis"
			}
			core, err := openCore(*workspace, *verbose)
			if err != nil {
				return err
			}
			defer core.Close()
			t, err := corepipe.NewTaskIR(summary, notes, "cli", "pipeline run")
			if err != nil {
				return err
			}
			ctx := context.Background()
			var events <-chan event.Event
			var unsub func()
			if wait {
				events, unsub = core.Subscribe(256)
				defer unsub()
			}
			runID, err := core.SubmitTask(ctx, t)
			if err != nil {
				return err
			}
			fmt.Println(runID)
			if !wait {
				return nil
			}
			return waitForRun(ctx, core, runID, events)
		},
	}
	run.Flags().StringVar(&summary, "summary", "", "intent.summary")
	run.Flags().StringVar(&notes, "notes", "", "intent.notes")
	run.Flags().BoolVar(&wait, "wait", true, "wait for completion")
	cmd.AddCommand(run)
	return cmd
}

func newBoardCmd(workspace *string, verbose *bool) *cobra.Command {
	cmd := &cobra.Command{Use: "board", Short: "GitHub Projects / issues Entry Point"}
	var (
		repo, label, owner string
		project            int
		org, wait          bool
		interval           time.Duration
	)
	poll := &cobra.Command{
		Use:   "poll",
		Short: "Poll cards, submit pipeline Task IR, write back status",
		RunE: func(cmd *cobra.Command, args []string) error {
			core, err := openCore(*workspace, *verbose)
			if err != nil {
				return err
			}
			defer core.Close()
			if core.Cfg.GitHubToken == "" {
				return fmt.Errorf("GITHUB_TOKEN or RUNTGINE_GITHUB_TOKEN is required")
			}
			ad := board.NewAdapter(core, board.NewGitHub(core.Cfg.GitHubToken), slog.Default())
			opt := board.PollOptions{
				Repo: repo, Label: label, ProjectOwner: owner,
				ProjectNumber: project, OwnerIsOrg: org, Wait: wait,
			}
			ctx := context.Background()
			if interval <= 0 {
				ids, err := ad.PollOnce(ctx, opt)
				if err != nil {
					return err
				}
				for _, id := range ids {
					fmt.Println(id)
				}
				return nil
			}
			for {
				ids, err := ad.PollOnce(ctx, opt)
				if err != nil {
					return err
				}
				for _, id := range ids {
					fmt.Println(id)
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(interval):
				}
			}
		},
	}
	poll.Flags().StringVar(&repo, "repo", "", "owner/name (issues mode)")
	poll.Flags().StringVar(&label, "label", "", "filter issues by label")
	poll.Flags().StringVar(&owner, "project-owner", "", "user/org login for Projects v2")
	poll.Flags().IntVar(&project, "project", 0, "GitHub Project number (Projects v2)")
	poll.Flags().BoolVar(&org, "org", false, "project-owner is an organization")
	poll.Flags().BoolVar(&wait, "wait", true, "wait each imported card before next")
	poll.Flags().DurationVar(&interval, "interval", 0, "repeat poll (e.g. 60s); 0 = once")
	cmd.AddCommand(poll)
	return cmd
}

func toJSON(path string, raw []byte) ([]byte, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		var v any
		if err := yaml.Unmarshal(raw, &v); err != nil {
			return nil, fmt.Errorf("yaml: %w", err)
		}
		return json.Marshal(v)
	default:
		return raw, nil
	}
}

func newLessonsCmd(workspace *string, verbose *bool) *cobra.Command {
	cmd := &cobra.Command{Use: "lessons", Short: "HITL postmortem proposals (G-150)"}

	list := &cobra.Command{
		Use:   "list",
		Short: "List lesson proposals as JSON",
		RunE: func(cmd *cobra.Command, args []string) error {
			core, err := openCore(*workspace, *verbose)
			if err != nil {
				return err
			}
			defer core.Close()
			status, _ := cmd.Flags().GetString("status")
			rows, err := core.ListLessons(context.Background(), status)
			if err != nil {
				return err
			}
			return writeJSON(os.Stdout, rows)
		},
	}
	list.Flags().String("status", "", "filter pending|approved|rejected")
	cmd.AddCommand(list)

	cmd.AddCommand(&cobra.Command{
		Use:   "approve <id>",
		Short: "Promote a pending proposal into Project Memory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			core, err := openCore(*workspace, *verbose)
			if err != nil {
				return err
			}
			defer core.Close()
			ep, err := core.ApproveLesson(context.Background(), args[0])
			if err != nil {
				return err
			}
			return writeJSON(os.Stdout, ep)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "reject <id>",
		Short: "Discard a pending proposal (no Memory write)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			core, err := openCore(*workspace, *verbose)
			if err != nil {
				return err
			}
			defer core.Close()
			return core.RejectLesson(context.Background(), args[0])
		},
	})
	return cmd
}

func newServeCmd(workspace *string, verbose *bool) *cobra.Command {
	var listen string
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "HTTP API Entry Point (G-153)",
		RunE: func(cmd *cobra.Command, args []string) error {
			core, err := openCore(*workspace, *verbose)
			if err != nil {
				return err
			}
			defer core.Close()
			if listen != "" {
				core.Cfg.API.Listen = listen
			}
			return httpapi.Serve(core)
		},
	}
	cmd.Flags().StringVar(&listen, "listen", "", "bind address (default: config api.listen / 127.0.0.1:7420)")
	return cmd
}
