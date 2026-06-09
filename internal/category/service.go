package category

import (
	"context"
	"database/sql"
	"strings"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, userID int64, name, kind, emoji string) (Category, error) {
	name = strings.TrimSpace(name)
	kind = strings.TrimSpace(kind)
	if kind == "" {
		kind = "expense"
	}
	return s.repo.Create(ctx, userID, name, kind, strings.TrimSpace(emoji))
}

func (s *Service) List(ctx context.Context, userID int64) ([]Category, error) {
	return s.repo.List(ctx, userID)
}

func (s *Service) FindOrCreate(ctx context.Context, userID int64, kind, name string) (Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "uncategorized"
	}
	if c, err := s.repo.FindByNameOrAlias(ctx, userID, kind, name); err == nil {
		return c, nil
	} else if err != sql.ErrNoRows {
		return Category{}, err
	}
	return s.repo.Create(ctx, userID, name, kind, "")
}

func (s *Service) SeedDefaults(ctx context.Context, userID int64) error {
	defaults := []struct {
		name string
		kind string
	}{
		{"groceries", "expense"},
		{"transport", "expense"},
		{"food", "expense"},
		{"bills", "expense"},
		{"health", "expense"},
		{"salary", "income"},
	}
	for _, item := range defaults {
		if _, err := s.repo.Create(ctx, userID, item.name, item.kind, ""); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) AddAlias(ctx context.Context, userID, categoryID int64, alias string) error {
	return s.repo.AddAlias(ctx, userID, categoryID, strings.TrimSpace(alias))
}
