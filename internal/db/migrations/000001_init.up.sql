CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    telegram_user_id BIGINT UNIQUE NOT NULL,
    display_name TEXT,
    username TEXT,
    currency TEXT NOT NULL DEFAULT 'TOMAN',
    timezone TEXT NOT NULL DEFAULT 'Asia/Tehran',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE accounts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    type TEXT NOT NULL DEFAULT 'cash',
    currency TEXT NOT NULL DEFAULT 'TOMAN',
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_accounts_user_id ON accounts(user_id);

CREATE TABLE categories (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    kind TEXT NOT NULL CHECK (kind IN ('expense', 'income')),
    emoji TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(user_id, name, kind)
);

CREATE INDEX idx_categories_user_id ON categories(user_id);

CREATE TABLE category_aliases (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id BIGINT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    alias TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(user_id, alias)
);

CREATE INDEX idx_category_aliases_user_id ON category_aliases(user_id);
CREATE INDEX idx_category_aliases_alias ON category_aliases(alias);

CREATE TABLE transactions (
    id BIGSERIAL PRIMARY KEY,

    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id BIGINT REFERENCES accounts(id) ON DELETE SET NULL,
    category_id BIGINT REFERENCES categories(id) ON DELETE SET NULL,

    kind TEXT NOT NULL CHECK (kind IN ('expense', 'income', 'transfer')),
    amount NUMERIC(18, 2) NOT NULL CHECK (amount > 0),
    currency TEXT NOT NULL DEFAULT 'TOMAN',

    merchant TEXT,
    note TEXT,
    raw_text TEXT,

    occurred_at TIMESTAMPTZ NOT NULL,

    jalali_year INT NOT NULL,
    jalali_month INT NOT NULL,
    jalali_day INT NOT NULL,
    jalali_ym TEXT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_user_jalali_ym ON transactions(user_id, jalali_ym);
CREATE INDEX idx_transactions_user_occurred_at ON transactions(user_id, occurred_at);
CREATE INDEX idx_transactions_category_id ON transactions(category_id);
