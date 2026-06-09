package user

import (
	"context"
	"database/sql"
)

type Service struct {
	repo            *Repository
	defaultCurrency string
	defaultTimezone string
}

func NewService(repo *Repository, defaultCurrency, defaultTimezone string) *Service {
	return &Service{repo: repo, defaultCurrency: defaultCurrency, defaultTimezone: defaultTimezone}
}

func (s *Service) EnsureTelegramUser(ctx context.Context, telegramUserID int64, displayName, username string) (User, error) {
	return s.repo.UpsertTelegram(ctx, telegramUserID, displayName, username, s.defaultCurrency, s.defaultTimezone)
}

func (s *Service) First(ctx context.Context) (User, error) {
	u, err := s.repo.First(ctx)
	if err == sql.ErrNoRows {
		return User{}, err
	}
	return u, err
}
