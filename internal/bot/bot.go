package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Bot struct {
	token   string
	client  *http.Client
	handler *Handler
	log     *slog.Logger
}

type Update struct {
	UpdateID int64    `json:"update_id"`
	Message  *Message `json:"message"`
}

type Message struct {
	MessageID int64         `json:"message_id"`
	From      *TelegramUser `json:"from"`
	Chat      Chat          `json:"chat"`
	Text      string        `json:"text"`
}

type TelegramUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

type Chat struct {
	ID int64 `json:"id"`
}

type apiResponse[T any] struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
	Result      T      `json:"result"`
}

func New(token string, handler *Handler, log *slog.Logger) *Bot {
	return &Bot{
		token: token,
		client: &http.Client{
			Timeout: 40 * time.Second,
		},
		handler: handler,
		log:     log,
	}
}

func (b *Bot) Run(ctx context.Context) {
	if b.token == "" {
		b.log.Info("telegram bot disabled because TELEGRAM_BOT_TOKEN is empty")
		return
	}

	var offset int64
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		updates, err := b.getUpdates(ctx, offset)
		if err != nil {
			b.log.Warn("telegram getUpdates failed", "error", err)
			sleep(ctx, 3*time.Second)
			continue
		}

		for _, update := range updates {
			offset = update.UpdateID + 1
			if update.Message == nil {
				continue
			}
			reply := b.handler.HandleMessage(ctx, *update.Message)
			if err := b.sendMessage(ctx, update.Message.Chat.ID, reply); err != nil {
				b.log.Warn("telegram sendMessage failed", "error", err)
			}
		}
	}
}

func (b *Bot) getUpdates(ctx context.Context, offset int64) ([]Update, error) {
	values := url.Values{}
	values.Set("timeout", "30")
	if offset > 0 {
		values.Set("offset", strconv.FormatInt(offset, 10))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, b.apiURL("getUpdates")+"?"+values.Encode(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var decoded apiResponse[[]Update]
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, err
	}
	if !decoded.OK {
		return nil, fmt.Errorf(decoded.Description)
	}
	return decoded.Result, nil
}

func (b *Bot) sendMessage(ctx context.Context, chatID int64, text string) error {
	payload := map[string]any{
		"chat_id": chatID,
		"text":    text,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.apiURL("sendMessage"), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var decoded apiResponse[json.RawMessage]
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return err
	}
	if !decoded.OK {
		return fmt.Errorf(decoded.Description)
	}
	return nil
}

func (b *Bot) apiURL(method string) string {
	return "https://api.telegram.org/bot" + b.token + "/" + method
}

func sleep(ctx context.Context, duration time.Duration) {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}
