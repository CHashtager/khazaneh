package category

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

func (r *Repository) Create(ctx context.Context, userID int64, name, kind, emoji string) (Category, error) {
	var c Category
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO categories (user_id, name, kind, emoji)
		VALUES ($1, $2, $3, NULLIF($4, ''))
		ON CONFLICT (user_id, name, kind) DO UPDATE SET emoji = COALESCE(EXCLUDED.emoji, categories.emoji)
		RETURNING id, user_id, name, kind, COALESCE(emoji, ''), created_at
	`, userID, name, kind, emoji).Scan(&c.ID, &c.UserID, &c.Name, &c.Kind, &c.Emoji, &c.CreatedAt)
	return c, err
}

func (r *Repository) List(ctx context.Context, userID int64) ([]Category, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, name, kind, COALESCE(emoji, ''), created_at
		FROM categories
		WHERE user_id = $1
		ORDER BY kind, name
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.Kind, &c.Emoji, &c.CreatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

func (r *Repository) FindByNameOrAlias(ctx context.Context, userID int64, kind, name string) (Category, error) {
	var c Category
	err := r.db.QueryRowContext(ctx, `
		SELECT c.id, c.user_id, c.name, c.kind, COALESCE(c.emoji, ''), c.created_at
		FROM categories c
		WHERE c.user_id = $1 AND c.kind = $2 AND lower(c.name) = lower($3)
		UNION ALL
		SELECT c.id, c.user_id, c.name, c.kind, COALESCE(c.emoji, ''), c.created_at
		FROM category_aliases a
		JOIN categories c ON c.id = a.category_id
		WHERE a.user_id = $1 AND c.kind = $2 AND lower(a.alias) = lower($3)
		LIMIT 1
	`, userID, kind, name).Scan(&c.ID, &c.UserID, &c.Name, &c.Kind, &c.Emoji, &c.CreatedAt)
	return c, err
}

func (r *Repository) AddAlias(ctx context.Context, userID, categoryID int64, alias string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO category_aliases (user_id, category_id, alias)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, alias) DO UPDATE SET category_id = EXCLUDED.category_id
	`, userID, categoryID, alias)
	return err
}
