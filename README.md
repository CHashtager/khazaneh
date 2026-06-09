# Khazaneh

Self-hosted personal expense tracker with Telegram-first input, PostgreSQL storage, and a Jalali-first web dashboard.

## Quick Start

1. Copy `.env.example` to `.env`.
2. Fill `TELEGRAM_BOT_TOKEN` and `ALLOWED_TELEGRAM_USER_IDS`.
3. Run:

```sh
docker compose up --build
```

The dashboard is available at `http://localhost:8080`.

## Telegram Input

The v1 parser is deterministic and amount-first. Examples:

```text
250000 groceries bread and milk
income 12000000 salary
```

Only Telegram users listed in `ALLOWED_TELEGRAM_USER_IDS` can use the bot.

## Dashboard

The dashboard uses session-cookie auth with `DASHBOARD_USERNAME` and `DASHBOARD_PASSWORD`.
Use it to review transactions and manage categories.

## Local Development

```sh
cp .env.example .env
docker compose up postgres
DATABASE_URL=postgres://expense:expense@localhost:5432/expense?sslmode=disable go run ./cmd/server
```

Migrations are applied automatically on startup.
