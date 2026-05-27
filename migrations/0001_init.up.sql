-- MilkNest MVP initial schema. Aligns with HLD §7 and LLD §5.
-- Money stored as BIGINT paise. Lat/long as NUMERIC(10,7).

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE user_role AS ENUM ('customer', 'admin', 'delivery');

CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            TEXT NOT NULL,
    phone_number    TEXT NOT NULL UNIQUE,
    email           TEXT,
    role            user_role NOT NULL DEFAULT 'customer',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_users_role ON users(role);

CREATE TABLE addresses (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    house_number    TEXT NOT NULL,
    street          TEXT NOT NULL,
    city            TEXT NOT NULL,
    state           TEXT NOT NULL,
    pincode         TEXT NOT NULL,
    latitude        NUMERIC(10,7) NOT NULL,
    longitude       NUMERIC(10,7) NOT NULL,
    is_default      BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_addresses_user ON addresses(user_id);

CREATE TABLE hubs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            TEXT NOT NULL,
    city            TEXT NOT NULL,
    pincode         TEXT NOT NULL,
    latitude        NUMERIC(10,7) NOT NULL,
    longitude       NUMERIC(10,7) NOT NULL,
    active          BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE products (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            TEXT NOT NULL,
    brand           TEXT NOT NULL,
    unit_size_ml    INTEGER NOT NULL CHECK (unit_size_ml > 0),
    price_paise     BIGINT NOT NULL CHECK (price_paise >= 0),
    image_url       TEXT,
    active          BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_products_active ON products(active);

CREATE TABLE inventory (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id          UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    hub_id              UUID NOT NULL REFERENCES hubs(id) ON DELETE RESTRICT,
    quantity            INTEGER NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    reserved_quantity   INTEGER NOT NULL DEFAULT 0 CHECK (reserved_quantity >= 0),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(product_id, hub_id),
    CHECK (reserved_quantity <= quantity)
);

CREATE TYPE subscription_frequency AS ENUM ('daily', 'weekly', 'monthly');
CREATE TYPE subscription_status AS ENUM ('active', 'paused', 'cancelled');

CREATE TABLE subscriptions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id      UUID NOT NULL REFERENCES products(id),
    hub_id          UUID NOT NULL REFERENCES hubs(id),
    address_id      UUID NOT NULL REFERENCES addresses(id),
    quantity        INTEGER NOT NULL CHECK (quantity > 0),
    free_quantity   INTEGER NOT NULL DEFAULT 0,
    frequency       subscription_frequency NOT NULL,
    delivery_time   TEXT NOT NULL,
    start_date      DATE NOT NULL,
    status          subscription_status NOT NULL DEFAULT 'active',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_subs_user_status ON subscriptions(user_id, status);
CREATE INDEX idx_subs_status ON subscriptions(status);

CREATE TYPE order_status AS ENUM (
    'pending', 'assigned', 'picked', 'in_transit', 'delivered', 'failed', 'cancelled'
);

CREATE TABLE orders (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_id UUID REFERENCES subscriptions(id),
    user_id         UUID NOT NULL REFERENCES users(id),
    product_id      UUID NOT NULL REFERENCES products(id),
    hub_id          UUID NOT NULL REFERENCES hubs(id),
    address_id      UUID NOT NULL REFERENCES addresses(id),
    delivery_date   DATE NOT NULL,
    quantity        INTEGER NOT NULL CHECK (quantity > 0),
    amount_paise    BIGINT NOT NULL CHECK (amount_paise >= 0),
    status          order_status NOT NULL DEFAULT 'pending',
    payment_id      UUID,
    delivery_partner_id UUID REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(subscription_id, delivery_date)
);

CREATE INDEX idx_orders_user ON orders(user_id);
CREATE INDEX idx_orders_status_date ON orders(status, delivery_date);
CREATE INDEX idx_orders_partner ON orders(delivery_partner_id);

CREATE TYPE payment_method AS ENUM ('upi', 'card', 'netbanking', 'bnpl', 'wallet');
CREATE TYPE payment_status AS ENUM ('created', 'paid', 'failed', 'refunded');

CREATE TABLE payments (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id             UUID REFERENCES orders(id),
    user_id              UUID NOT NULL REFERENCES users(id),
    amount_paise         BIGINT NOT NULL CHECK (amount_paise >= 0),
    surcharge_paise      BIGINT NOT NULL DEFAULT 0 CHECK (surcharge_paise >= 0),
    method               payment_method NOT NULL,
    razorpay_order_id    TEXT,
    razorpay_payment_id  TEXT UNIQUE,
    status               payment_status NOT NULL DEFAULT 'created',
    idempotency_key      TEXT UNIQUE,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_payments_order ON payments(order_id);
CREATE INDEX idx_payments_user ON payments(user_id);
CREATE INDEX idx_payments_rzp_order ON payments(razorpay_order_id);

CREATE TYPE delivery_status AS ENUM (
    'picked', 'in_transit', 'delivered', 'failed'
);

CREATE TABLE delivery_tracking (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id            UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    delivery_partner_id UUID NOT NULL REFERENCES users(id),
    latitude            NUMERIC(10,7) NOT NULL,
    longitude           NUMERIC(10,7) NOT NULL,
    status              delivery_status NOT NULL,
    recorded_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_tracking_order_time ON delivery_tracking(order_id, recorded_at DESC);

CREATE TYPE complaint_type AS ENUM ('late_delivery', 'damaged_product', 'other');
CREATE TYPE complaint_status AS ENUM ('open', 'in_review', 'resolved');

CREATE TABLE complaints (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_id        UUID REFERENCES orders(id),
    type            complaint_type NOT NULL,
    description     TEXT NOT NULL,
    status          complaint_status NOT NULL DEFAULT 'open',
    resolution_note TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at     TIMESTAMPTZ
);

CREATE INDEX idx_complaints_user ON complaints(user_id);
CREATE INDEX idx_complaints_status ON complaints(status);

CREATE TABLE offers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            TEXT NOT NULL,
    min_quantity    INTEGER NOT NULL CHECK (min_quantity > 0),
    free_quantity   INTEGER NOT NULL CHECK (free_quantity >= 0),
    active          BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_offers_active_min ON offers(active, min_quantity);

CREATE TYPE notification_channel AS ENUM ('sms', 'push');
CREATE TYPE notification_status AS ENUM ('queued', 'sent', 'failed');

CREATE TABLE notifications (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    channel         notification_channel NOT NULL,
    title           TEXT NOT NULL,
    body            TEXT NOT NULL,
    status          notification_status NOT NULL DEFAULT 'queued',
    sent_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_notifications_user ON notifications(user_id, created_at DESC);

-- Wallet extension (user-requested, on top of API spec)
CREATE TABLE wallets (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    balance_paise   BIGINT NOT NULL DEFAULT 0 CHECK (balance_paise >= 0),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TYPE wallet_txn_type AS ENUM ('credit', 'debit');
CREATE TYPE wallet_ref_type AS ENUM ('topup', 'order', 'refund', 'subscription');

CREATE TABLE wallet_transactions (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id            UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    type                 wallet_txn_type NOT NULL,
    amount_paise         BIGINT NOT NULL CHECK (amount_paise > 0),
    ref_type             wallet_ref_type NOT NULL,
    ref_id               TEXT NOT NULL,
    balance_after_paise  BIGINT NOT NULL,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(ref_type, ref_id)
);

CREATE INDEX idx_wallet_txn_wallet ON wallet_transactions(wallet_id, created_at DESC);
