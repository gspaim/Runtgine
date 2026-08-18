package claim

import (
	"context"

	"github.com/gspaim/Runtgine/internal/core/store"
)

// Service persists exclusive claims. Store owns SQL; this package owns overlap.
type Service struct {
	Store *store.Store
}

func New(st *store.Store) *Service {
	return &Service{Store: st}
}

func (s *Service) Acquire(ctx context.Context, runID, stepID string, res Resource) error {
	if s == nil || s.Store == nil {
		return nil
	}
	overlaps := func(aKind, aKey, bKind, bKey string) bool {
		return Overlaps(Resource{Kind: Kind(aKind), Key: aKey}, Resource{Kind: Kind(bKind), Key: bKey})
	}
	holder, err := s.Store.TryAcquireClaim(ctx, runID, stepID, string(res.Kind), res.Key, overlaps)
	if err != nil {
		return err
	}
	if holder != "" {
		return ConflictError(res, holder)
	}
	return nil
}

func (s *Service) ReleaseAll(ctx context.Context, runID string) ([]Resource, error) {
	if s == nil || s.Store == nil {
		return nil, nil
	}
	rows, err := s.Store.ReleaseClaimsForRun(ctx, runID)
	if err != nil {
		return nil, err
	}
	out := make([]Resource, 0, len(rows))
	for _, row := range rows {
		out = append(out, Resource{Kind: Kind(row.Kind), Key: row.Key})
	}
	return out, nil
}

func (s *Service) SweepOrphans(ctx context.Context) error {
	if s == nil || s.Store == nil {
		return nil
	}
	_, err := s.Store.SweepOrphanClaims(ctx)
	return err
}
