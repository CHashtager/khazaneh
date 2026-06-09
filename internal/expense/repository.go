package expense

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

func (r *Repository) Create(ctx context.Context, tx CreateTransaction, jy, jm, jd int, jym string) (Transaction, error) {
	var out Transaction
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO transactions (
			user_id, category_id, kind, amount, currency, merchant, note, raw_text,
			occurred_at, jalali_year, jalali_month, jalali_day, jalali_ym
		)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), NULLIF($7, ''), NULLIF($8, ''), $9, $10, $11, $12, $13)
		RETURNING id, user_id, account_id, category_id, kind, amount::text, currency,
			COALESCE(merchant, ''), COALESCE(note, ''), COALESCE(raw_text, ''),
			occurred_at, jalali_year, jalali_month, jalali_day, jalali_ym, created_at, updated_at
	`, tx.UserID, tx.CategoryID, tx.Kind, tx.Amount, tx.Currency, tx.Merchant, tx.Note, tx.RawText, tx.OccurredAt, jy, jm, jd, jym).Scan(
		&out.ID,
		&out.UserID,
		&out.AccountID,
		&out.CategoryID,
		&out.Kind,
		&out.Amount,
		&out.Currency,
		&out.Merchant,
		&out.Note,
		&out.RawText,
		&out.OccurredAt,
		&out.JalaliYear,
		&out.JalaliMonth,
		&out.JalaliDay,
		&out.JalaliYM,
		&out.CreatedAt,
		&out.UpdatedAt,
	)
	return out, err
}

func (r *Repository) ListRecent(ctx context.Context, userID int64, limit int) ([]Transaction, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT t.id, t.user_id, t.account_id, t.category_id, COALESCE(c.name, ''),
			t.kind, t.amount::text, t.currency, COALESCE(t.merchant, ''), COALESCE(t.note, ''),
			COALESCE(t.raw_text, ''), t.occurred_at, t.jalali_year, t.jalali_month, t.jalali_day,
			t.jalali_ym, t.created_at, t.updated_at
		FROM transactions t
		LEFT JOIN categories c ON c.id = t.category_id
		WHERE t.user_id = $1
		ORDER BY t.occurred_at DESC, t.id DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var tx Transaction
		if err := rows.Scan(
			&tx.ID,
			&tx.UserID,
			&tx.AccountID,
			&tx.CategoryID,
			&tx.Category,
			&tx.Kind,
			&tx.Amount,
			&tx.Currency,
			&tx.Merchant,
			&tx.Note,
			&tx.RawText,
			&tx.OccurredAt,
			&tx.JalaliYear,
			&tx.JalaliMonth,
			&tx.JalaliDay,
			&tx.JalaliYM,
			&tx.CreatedAt,
			&tx.UpdatedAt,
		); err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}
	return transactions, rows.Err()
}

func (r *Repository) Get(ctx context.Context, userID, id int64) (Transaction, error) {
	var tx Transaction
	err := r.db.QueryRowContext(ctx, `
		SELECT t.id, t.user_id, t.account_id, t.category_id, COALESCE(c.name, ''),
			t.kind, t.amount::text, t.currency, COALESCE(t.merchant, ''), COALESCE(t.note, ''),
			COALESCE(t.raw_text, ''), t.occurred_at, t.jalali_year, t.jalali_month, t.jalali_day,
			t.jalali_ym, t.created_at, t.updated_at
		FROM transactions t
		LEFT JOIN categories c ON c.id = t.category_id
		WHERE t.user_id = $1 AND t.id = $2
	`, userID, id).Scan(
		&tx.ID,
		&tx.UserID,
		&tx.AccountID,
		&tx.CategoryID,
		&tx.Category,
		&tx.Kind,
		&tx.Amount,
		&tx.Currency,
		&tx.Merchant,
		&tx.Note,
		&tx.RawText,
		&tx.OccurredAt,
		&tx.JalaliYear,
		&tx.JalaliMonth,
		&tx.JalaliDay,
		&tx.JalaliYM,
		&tx.CreatedAt,
		&tx.UpdatedAt,
	)
	return tx, err
}

func (r *Repository) Update(ctx context.Context, userID, id int64, categoryID *int64, kind, amount, merchant, note string) (Transaction, error) {
	var tx Transaction
	err := r.db.QueryRowContext(ctx, `
		UPDATE transactions
		SET category_id = $3,
		    kind = $4,
		    amount = $5,
		    merchant = NULLIF($6, ''),
		    note = NULLIF($7, ''),
		    updated_at = now()
		WHERE user_id = $1 AND id = $2
		RETURNING id, user_id, account_id, category_id, kind, amount::text, currency,
			COALESCE(merchant, ''), COALESCE(note, ''), COALESCE(raw_text, ''),
			occurred_at, jalali_year, jalali_month, jalali_day, jalali_ym, created_at, updated_at
	`, userID, id, categoryID, kind, amount, merchant, note).Scan(
		&tx.ID,
		&tx.UserID,
		&tx.AccountID,
		&tx.CategoryID,
		&tx.Kind,
		&tx.Amount,
		&tx.Currency,
		&tx.Merchant,
		&tx.Note,
		&tx.RawText,
		&tx.OccurredAt,
		&tx.JalaliYear,
		&tx.JalaliMonth,
		&tx.JalaliDay,
		&tx.JalaliYM,
		&tx.CreatedAt,
		&tx.UpdatedAt,
	)
	return tx, err
}

func (r *Repository) Summary(ctx context.Context, userID int64, jalaliYM string) (MonthlySummary, error) {
	var summary MonthlySummary
	err := r.db.QueryRowContext(ctx, `
		SELECT $2,
			COALESCE(SUM(amount) FILTER (WHERE kind = 'expense'), 0)::text,
			COALESCE(SUM(amount) FILTER (WHERE kind = 'income'), 0)::text
		FROM transactions
		WHERE user_id = $1 AND jalali_ym = $2
	`, userID, jalaliYM).Scan(&summary.JalaliYM, &summary.ExpenseTotal, &summary.IncomeTotal)
	return summary, err
}
