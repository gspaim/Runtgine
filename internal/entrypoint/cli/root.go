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
	"github.com/gspaim/Runtgine/internal/core/store"
	"github.com/gspaim/Runtgine/internal/core/task"
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
	root.AddCommand(newStatusCmd(&workspace, &verbose))
	root.AddCommand(newCancelCmd(&workspace, &verbose))
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
