package memory

import (
	"context"

	"github.com/gspaim/Runtgine/internal/core/memory"
)

// NopReader is a safe fallback used when the Player is constructed
// without a real Reader (defensive only; production wires *Service).
// It always returns empty results and never errors.
type NopReader struct{}

func (NopReader) Recall(_ context.Context, _ memory.RecallQuery) ([]memory.Hit, error) {
	return nil, nil
}

func (NopReader) Check(_ context.Context, _ string) (memory.CheckResult, error) {
	return memory.CheckResult{}, nil
}