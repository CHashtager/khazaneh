package expense

import (
	"context"
	"time"

	"github.com/CHashtager/khazaneh/internal/calendar"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, tx CreateTransaction, location *time.Location) (Transaction, error) {
	j := calendar.FromTime(tx.OccurredAt, location)
	return s.repo.Create(ctx, tx, j.Year, j.Month, j.Day, j.YearMonth())
}

func (s *Service) ListRecent(ctx context.Context, userID int64, limit int) ([]Transaction, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return s.repo.ListRecent(ctx, userID, limit)
}

func (s *Service) Get(ctx context.Context, userID, id int64) (Transaction, error) {
	return s.repo.Get(ctx, userID, id)
}

func (s *Service) Update(ctx context.Context, userID, id int64, categoryID *int64, kind, amount, merchant, note string) (Transaction, error) {
	return s.repo.Update(ctx, userID, id, categoryID, kind, amount, merchant, note)
}

func (s *Service) Summary(ctx context.Context, userID int64, jalaliYM string) (MonthlySummary, error) {
	return s.repo.Summary(ctx, userID, jalaliYM)
}
