package bot

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/CHashtager/khazaneh/internal/category"
	"github.com/CHashtager/khazaneh/internal/expense"
	"github.com/CHashtager/khazaneh/internal/user"
)

type Handler struct {
	users           *user.Service
	categories      *category.Service
	expenses        *expense.Service
	allowedUserIDs  map[int64]bool
	defaultCurrency string
}

func NewHandler(users *user.Service, categories *category.Service, expenses *expense.Service, allowed map[int64]bool, defaultCurrency string) *Handler {
	return &Handler{
		users:           users,
		categories:      categories,
		expenses:        expenses,
		allowedUserIDs:  allowed,
		defaultCurrency: defaultCurrency,
	}
}

func (h *Handler) HandleMessage(ctx context.Context, msg Message) string {
	if msg.From == nil || !h.allowedUserIDs[msg.From.ID] {
		return unauthorizedMessage
	}

	text := strings.TrimSpace(msg.Text)
	if text == "" {
		return "Please send a text expense."
	}
	if strings.HasPrefix(text, "/start") || strings.HasPrefix(text, "/help") {
		return welcomeMessage
	}

	u, err := h.users.EnsureTelegramUser(ctx, msg.From.ID, displayName(*msg.From), msg.From.Username)
	if err != nil {
		return "Could not create your user."
	}
	_ = h.categories.SeedDefaults(ctx, u.ID)

	parsed, err := ParseTransaction(text)
	if err != nil {
		return "I could not parse that. Try: 250000 groceries bread"
	}

	c, err := h.categories.FindOrCreate(ctx, u.ID, parsed.Kind, parsed.Category)
	if err != nil {
		return "Could not resolve the category."
	}

	location, err := time.LoadLocation(u.Timezone)
	if err != nil {
		location = time.Local
	}
	categoryID := c.ID
	tx, err := h.expenses.Create(ctx, expense.CreateTransaction{
		UserID:     u.ID,
		CategoryID: &categoryID,
		Kind:       parsed.Kind,
		Amount:     parsed.Amount,
		Currency:   h.defaultCurrency,
		Note:       parsed.Note,
		RawText:    text,
		OccurredAt: time.Now(),
	}, location)
	if err != nil {
		return "Could not save the transaction."
	}

	return fmt.Sprintf("Saved %s %s %s in %s for %s.", tx.Kind, tx.Amount, tx.Currency, c.Name, tx.JalaliYM)
}

func displayName(u TelegramUser) string {
	name := strings.TrimSpace(strings.Join([]string{u.FirstName, u.LastName}, " "))
	if name == "" {
		return u.Username
	}
	return name
}
