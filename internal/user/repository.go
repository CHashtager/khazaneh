package user

import (
	"context"
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) UpsertTelegram(ctx context.Context, telegramUserID int64, displayName, username, currency, timezone string) (User, error) {
	var u User
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO users (telegram_user_id, display_name, username, currency, timezone)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (telegram_user_id) DO UPDATE
		SET display_name = EXCLUDED.display_name,
		    username = EXCLUDED.username
		RETURNING id, telegram_user_id, COALESCE(display_name, ''), COALESCE(username, ''), currency, timezone, created_at
	`, telegramUserID, displayName, username, currency, timezone).Scan(
		&u.ID,
		&u.TelegramUserID,
		&u.DisplayName,
		&u.Username,
		&u.Currency,
		&u.Timezone,
		&u.CreatedAt,
	)
	return u, err
}

func (r *Repository) First(ctx context.Context) (User, error) {
	var u User
	err := r.db.QueryRowContext(ctx, `
		SELECT id, telegram_user_id, COALESCE(display_name, ''), COALESCE(username, ''), currency, timezone, created_at
		FROM users
		ORDER BY created_at ASC
		LIMIT 1
	`).Scan(&u.ID, &u.TelegramUserID, &u.DisplayName, &u.Username, &u.Currency, &u.Timezone, &u.CreatedAt)
	return u, err
}
