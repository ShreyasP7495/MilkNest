package subscriptions

import (
	"context"
	"database/sql"
	"errors"

	"github.com/milknest/backend/internal/inventory"
	"github.com/milknest/backend/internal/offers"
	"github.com/milknest/backend/pkg/config"
	"github.com/milknest/backend/pkg/utils"
)

// Service composes the subscription repository with inventory and offers. It
// owns the LLD §6 business rules (max 10 packets/day, offer calc, inventory
// reservation in a single transaction).
type Service struct {
	repo       *Repository
	inv        *inventory.Service
	offers     *offers.Service
	maxPackets int
}

func NewService(repo *Repository, inv *inventory.Service, of *offers.Service, cfg *config.Config) *Service {
	return &Service{repo: repo, inv: inv, offers: of, maxPackets: cfg.Business.MaxPacketsPerDay}
}

// Create runs the full LLD §6 flow inside a single Postgres transaction:
//  1. Verify quantity is <= MAX_PACKETS_PER_DAY (per-line check)
//  2. Verify SUM(active) + quantity <= MAX_PACKETS_PER_DAY (household check)
//  3. Compute free_quantity from offers
//  4. Insert the subscription row
//  5. Reserve inventory at the chosen hub
func (s *Service) Create(ctx context.Context, userID string, req CreateRequest) (*Subscription, error) {
	if req.Quantity > s.maxPackets {
		return nil, utils.NewBadRequest("max_packets_exceeded",
			"quantity exceeds the per-household daily limit")
	}

	tx, err := s.repo.DB().BeginTx(ctx, nil)
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	defer func() { _ = tx.Rollback() }()

	existing, err := s.repo.SumActiveQuantity(ctx, tx, userID)
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	if existing+req.Quantity > s.maxPackets {
		return nil, utils.NewBadRequest("max_packets_exceeded",
			"adding this subscription would exceed the household daily limit")
	}

	freeQty, err := s.offers.CalculateFree(ctx, req.Quantity)
	if err != nil {
		return nil, err
	}

	sub, err := s.repo.Create(ctx, tx, userID, req, freeQty)
	if err != nil {
		return nil, utils.NewInternal(err)
	}

	if err := s.inv.Repo.Reserve(ctx, tx, req.ProductID, req.HubID, req.Quantity); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, utils.NewConflict("insufficient_stock", "insufficient stock at hub")
		}
		return nil, utils.NewInternal(err)
	}

	if err := tx.Commit(); err != nil {
		return nil, utils.NewInternal(err)
	}
	return sub, nil
}

func (s *Service) List(ctx context.Context, userID string) ([]Subscription, error) {
	out, err := s.repo.ListForUser(ctx, userID)
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	return out, nil
}

func (s *Service) Get(ctx context.Context, userID, id string) (*Subscription, error) {
	sub, err := s.repo.GetForUser(ctx, userID, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.NewNotFound("subscription")
	}
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	return sub, nil
}

// Update adjusts quantity / delivery_time / address_id and re-reserves
// inventory as required.
func (s *Service) Update(ctx context.Context, userID, id string, req UpdateRequest) (*Subscription, error) {
	if req.Quantity != nil && *req.Quantity > s.maxPackets {
		return nil, utils.NewBadRequest("max_packets_exceeded", "quantity exceeds the daily limit")
	}

	tx, err := s.repo.DB().BeginTx(ctx, nil)
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	defer func() { _ = tx.Rollback() }()

	if req.Quantity != nil {
		existing, err := s.repo.SumActiveQuantity(ctx, tx, userID)
		if err != nil {
			return nil, utils.NewInternal(err)
		}
		current, err := s.repo.GetForUser(ctx, userID, id)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, utils.NewNotFound("subscription")
		}
		if err != nil {
			return nil, utils.NewInternal(err)
		}
		newHouseholdTotal := existing - current.Quantity + *req.Quantity
		if newHouseholdTotal > s.maxPackets {
			return nil, utils.NewBadRequest("max_packets_exceeded",
				"resulting quantity exceeds the household daily limit")
		}
	}

	updated, delta, err := s.repo.Update(ctx, tx, userID, id, req)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.NewNotFound("subscription")
	}
	if err != nil {
		return nil, utils.NewInternal(err)
	}

	if delta > 0 {
		if err := s.inv.Repo.Reserve(ctx, tx, updated.ProductID, updated.HubID, delta); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, utils.NewConflict("insufficient_stock", "insufficient stock at hub")
			}
			return nil, utils.NewInternal(err)
		}
	} else if delta < 0 {
		if err := s.inv.Repo.Release(ctx, tx, updated.ProductID, updated.HubID, -delta); err != nil {
			return nil, utils.NewInternal(err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, utils.NewInternal(err)
	}
	return updated, nil
}

// Pause moves the subscription to status=paused and releases its reserved
// inventory so other users can subscribe to the freed packets.
func (s *Service) Pause(ctx context.Context, userID, id string) (*Subscription, error) {
	return s.transitionStatus(ctx, userID, id, "paused", true)
}

// Cancel is identical to pause but terminal.
func (s *Service) Cancel(ctx context.Context, userID, id string) (*Subscription, error) {
	return s.transitionStatus(ctx, userID, id, "cancelled", true)
}

func (s *Service) transitionStatus(ctx context.Context, userID, id, status string, releaseInventory bool) (*Subscription, error) {
	tx, err := s.repo.DB().BeginTx(ctx, nil)
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	defer func() { _ = tx.Rollback() }()

	sub, err := s.repo.SetStatus(ctx, tx, userID, id, status)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.NewNotFound("subscription")
	}
	if err != nil {
		return nil, utils.NewInternal(err)
	}

	if releaseInventory {
		if err := s.inv.Repo.Release(ctx, tx, sub.ProductID, sub.HubID, sub.Quantity); err != nil {
			return nil, utils.NewInternal(err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, utils.NewInternal(err)
	}
	return sub, nil
}
