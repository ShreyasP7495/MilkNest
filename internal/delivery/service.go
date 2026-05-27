package delivery

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/milknest/backend/internal/orders"
	"github.com/milknest/backend/pkg/middleware"
	"github.com/milknest/backend/pkg/utils"
)

const (
	redisLocationTTL = 60 * time.Second
)

type Service struct {
	repo   *Repository
	orders *orders.Repository
	rdb    *redis.Client
}

func NewService(repo *Repository, ordersRepo *orders.Repository, rdb *redis.Client) *Service {
	return &Service{repo: repo, orders: ordersRepo, rdb: rdb}
}

// Assign is admin-only: link a delivery partner to an order and mark
// status='assigned'.
func (s *Service) Assign(ctx context.Context, req AssignRequest) (*orders.Order, error) {
	o, err := s.orders.Assign(ctx, req.OrderID, req.DeliveryPartnerID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.NewNotFound("order")
	}
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	return o, nil
}

// UpdateStatus is called by the delivery partner. It persists the tracking
// row, mirrors the location to Redis with TTL, and advances the order status.
// The order must already be assigned to this partner.
func (s *Service) UpdateStatus(ctx context.Context, partnerID string, req StatusUpdateRequest) (*Tracking, error) {
	o, err := s.orders.Get(ctx, req.OrderID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.NewNotFound("order")
	}
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	if o.DeliveryPartnerID == nil || *o.DeliveryPartnerID != partnerID {
		return nil, utils.NewForbidden("order is not assigned to this partner")
	}

	t, err := s.repo.Append(ctx, req.OrderID, partnerID, req.Latitude, req.Longitude, req.Status)
	if err != nil {
		return nil, utils.NewInternal(err)
	}

	if _, err := s.orders.SetStatus(ctx, req.OrderID, req.Status); err != nil {
		return nil, utils.NewInternal(err)
	}

	payload, _ := json.Marshal(LiveLocation{Lat: req.Latitude, Lng: req.Longitude})
	_ = s.rdb.Set(ctx, redisLocKey(req.OrderID), payload, redisLocationTTL).Err()

	return t, nil
}

// Track returns the order's current status + live location (Redis, falling
// back to last DB row).
func (s *Service) Track(ctx context.Context, callerID, callerRole, orderID string) (*TrackResponse, error) {
	o, err := s.orders.Get(ctx, orderID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.NewNotFound("order")
	}
	if err != nil {
		return nil, utils.NewInternal(err)
	}
	if callerRole == middleware.RoleCustomer && o.UserID != callerID {
		return nil, utils.NewForbidden("order belongs to another user")
	}

	resp := &TrackResponse{OrderID: o.ID, CurrentStatus: o.Status}

	val, err := s.rdb.Get(ctx, redisLocKey(orderID)).Result()
	if err == nil {
		var loc LiveLocation
		if jsonErr := json.Unmarshal([]byte(val), &loc); jsonErr == nil {
			resp.LiveLocation = &loc
			return resp, nil
		}
	}

	t, err := s.repo.LatestForOrder(ctx, orderID)
	if err == nil {
		resp.LiveLocation = &LiveLocation{Lat: t.Latitude, Lng: t.Longitude}
	}
	return resp, nil
}

func redisLocKey(orderID string) string { return fmt.Sprintf("delivery:%s:location", orderID) }
