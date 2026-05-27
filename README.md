# MilkNest Backend

Hyperlocal subscription-based daily milk delivery platform for Indian households (Bangalore / Mumbai launch).

This repository contains the MVP backend, a **modular monolith** written in Go (Gin), backed by PostgreSQL and Redis. The design follows the documents under [`docs/`](docs/).

---

## Quick start

```bash
# 1. Copy env template and fill in credentials
cp .env.example .env

# 2. Start Postgres + Redis (+ optionally the app)
docker compose up -d db redis

# 3. Run migrations (install https://github.com/golang-migrate/migrate first)
make migrate-up

# 4. Run the API
make run
# or: docker compose up -d app
```

The API listens on `http://localhost:8080` and exposes `GET /health`.

## Tech stack

| Concern         | Choice                                     |
| --------------- | ------------------------------------------ |
| Language        | Go 1.22                                    |
| HTTP framework  | Gin                                        |
| Database        | PostgreSQL 15 (via `database/sql` + `pgx`) |
| Cache / OTP / live GPS | Redis 7                              |
| Migrations      | `golang-migrate` (SQL files)               |
| Auth            | Phone OTP -> JWT (HS256, access + refresh) |
| Payments        | Razorpay (UPI / Card / Netbanking / BNPL)  |
| Cron            | `robfig/cron/v3` (midnight IST)            |
| Logging         | `zerolog`                                  |
| Config          | `viper` + `.env`                           |

## Project layout

```
.
├── cmd/server/main.go             # entrypoint, DI wiring, cron start
├── internal/
│   ├── auth/         users/        addresses/   products/
│   ├── inventory/    subscriptions/orders/      payments/
│   ├── delivery/     complaints/   notifications/
│   ├── offers/       wallet/       admin/
├── pkg/
│   ├── db/   redis/   config/   logger/   middleware/   utils/
├── migrations/                    # SQL migrations
├── docs/                          # HLD / LLD / API Spec
├── Dockerfile  docker-compose.yml  Makefile  .env.example
```

Each `internal/<module>/` has the four LLD layers:

- `handler.go` - Gin HTTP handlers
- `service.go` - business rules + transactions
- `repository.go` - SQL access
- `model.go` - DTOs + DB row structs

## API surface

Versioned under `/api/v1`. See [`docs/Milk-Nest-API-Specification.txt`](docs/Milk-Nest-API-Specification.txt) for the full contract. Highlights:

- `POST /auth/send-otp`, `POST /auth/verify-otp`, `POST /auth/refresh`
- `GET/PUT /users/me`
- `POST/GET/PUT/DELETE /addresses[/:id]`
- `GET /products`, `GET /products/:id`
- `POST/GET/PUT /subscriptions[/:id]`, `/subscriptions/:id/pause|cancel`
- `GET /orders`, `GET /orders/:id`, `GET /admin/orders`
- `POST /payments/create`, `POST /payments/webhook`, `GET /payments/:id`
- `POST /delivery/assign`, `POST /delivery/status`, `GET /delivery/track/:order_id`
- `POST/GET /complaints`, `POST /complaints/:id/resolve`
- `GET /offers`
- `GET /wallet`, `POST /wallet/topup`, `GET /wallet/transactions`
- `GET /admin/dashboard`, `POST /admin/inventory/update`

JWT required on every route except `/auth/*`, `/health`, and `/payments/webhook` (signature-verified instead).

## Business rules implemented

- Phone-OTP login (OTP stored in Redis with 5-min TTL; SMS provider is a stub by default)
- **Max 10 packets per household per day** (LLD §6)
- **Offers**: 50 packets -> +2 free, 100 packets -> +5 free (data-driven from `offers` table)
- **BNPL** payments add a **10% surcharge** (LLD §11)
- Inventory is per `(product_id, hub_id)` and is **reserved at subscription create**, **decremented at order generation** (LLD §6, §8)
- Daily order generation runs as a cron job at midnight IST (LLD §7)
- Delivery partner GPS is persisted to `delivery_tracking` and mirrored to Redis at `delivery:{order_id}:location` with 60s TTL (LLD §10)
- Wallet ledger is row-locked, double-entry, and idempotent on `(ref_type, ref_id)`

## Configuration

All configuration is read from environment variables (see [`.env.example`](.env.example)). The app reads `.env` automatically in development.

## Database

Schema lives in [`migrations/0001_init.up.sql`](migrations/0001_init.up.sql). The seed migration (`0002_seed.up.sql`) inserts:

- The default offers ladder (50/+2, 100/+5)
- A sample hub and two sample products (Amul, Nandini)

## Future work

See `docs/Milk-Nest-LLD.txt` §19 (Kafka, auto routing, microservices split, WhatsApp ordering, etc.).
