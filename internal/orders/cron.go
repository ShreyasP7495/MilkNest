package orders

import (
	"context"
	"database/sql"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/milknest/backend/internal/inventory"
	"github.com/milknest/backend/internal/subscriptions"
	"github.com/milknest/backend/pkg/config"
	"github.com/milknest/backend/pkg/logger"
	"github.com/milknest/backend/pkg/utils"
)

// productPriceLookup is a tiny helper so we don't have to import the products
// package (which would create a cycle later). It runs a single SELECT per
// generation.
type productPriceLookup struct{ db *sql.DB }

func (p *productPriceLookup) Price(ctx context.Context, tx *sql.Tx, productID string) (int64, error) {
	const q = `SELECT price_paise FROM products WHERE id = $1`
	var price int64
	err := tx.QueryRowContext(ctx, q, productID).Scan(&price)
	return price, err
}

// OrderGenerator owns the nightly job described in LLD §7. Concretely:
//
//	for each active subscription:
//	  in a single transaction:
//	    INSERT INTO orders (...) ON CONFLICT DO NOTHING       -- idempotent
//	    UPDATE inventory SET quantity = quantity - n,
//	                         reserved_quantity = reserved_quantity - n
//	                         WHERE product_id = ? AND hub_id = ? AND quantity >= n
//
// Failures on a single subscription do not abort the whole run.
type OrderGenerator struct {
	db    *sql.DB
	subs  *subscriptions.Repository
	inv   *inventory.Repository
	repo  *Repository
	price *productPriceLookup
}

func NewOrderGenerator(db *sql.DB, subs *subscriptions.Repository, inv *inventory.Repository, repo *Repository) *OrderGenerator {
	return &OrderGenerator{
		db:    db,
		subs:  subs,
		inv:   inv,
		repo:  repo,
		price: &productPriceLookup{db: db},
	}
}

// Start registers the cron schedule against the IST timezone. Returns the
// underlying cron so callers can Stop() it on shutdown.
func (g *OrderGenerator) Start(ctx context.Context, cfg config.CronConfig) (*cron.Cron, error) {
	loc, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		loc = utils.IST()
	}
	c := cron.New(cron.WithLocation(loc))
	if _, err := c.AddFunc(cfg.OrderGenSpec, func() {
		runCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		g.Run(runCtx)
	}); err != nil {
		return nil, err
	}
	c.Start()
	logger.L().Info().Str("spec", cfg.OrderGenSpec).Str("tz", cfg.Timezone).Msg("order generation cron scheduled")
	return c, nil
}

// Run executes one materialization pass. It is exported so the admin can
// trigger it manually (and so tests can drive it).
func (g *OrderGenerator) Run(ctx context.Context) {
	deliveryDate := utils.TomorrowIST()
	log := logger.L().With().Time("delivery_date", deliveryDate).Logger()

	active, err := g.subs.ListActiveAll(ctx)
	if err != nil {
		log.Error().Err(err).Msg("list active subscriptions failed")
		return
	}
	log.Info().Int("active_count", len(active)).Msg("starting order generation pass")

	var generated, skipped, failed int
	for _, sub := range active {
		if err := g.generateOne(ctx, sub, deliveryDate); err != nil {
			log.Error().Err(err).Str("subscription_id", sub.ID).Msg("failed to materialize subscription")
			failed++
		} else {
			generated++
		}
	}

	skipped = len(active) - generated - failed
	log.Info().
		Int("generated", generated).
		Int("skipped_or_dupe", skipped).
		Int("failed", failed).
		Msg("order generation pass complete")
}

func (g *OrderGenerator) generateOne(ctx context.Context, sub subscriptions.Subscription, deliveryDate time.Time) error {
	tx, err := g.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	price, err := g.price.Price(ctx, tx, sub.ProductID)
	if err != nil {
		return err
	}
	amount := price * int64(sub.Quantity)

	_, inserted, err := g.repo.InsertGeneratedOrder(ctx, tx, generatedOrderInput{
		SubscriptionID: sub.ID,
		UserID:         sub.UserID,
		ProductID:      sub.ProductID,
		HubID:          sub.HubID,
		AddressID:      sub.AddressID,
		Quantity:       sub.Quantity,
		AmountPaise:    amount,
	}, deliveryDate)
	if err != nil {
		return err
	}
	if !inserted {
		return tx.Commit() // already generated; idempotent skip
	}

	if err := g.inv.Consume(ctx, tx, sub.ProductID, sub.HubID, sub.Quantity); err != nil {
		return err
	}
	return tx.Commit()
}
