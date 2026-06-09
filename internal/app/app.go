package app

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/CHashtager/khazaneh/internal/bot"
	"github.com/CHashtager/khazaneh/internal/category"
	"github.com/CHashtager/khazaneh/internal/config"
	"github.com/CHashtager/khazaneh/internal/db"
	"github.com/CHashtager/khazaneh/internal/expense"
	"github.com/CHashtager/khazaneh/internal/user"
	"github.com/CHashtager/khazaneh/internal/web"
)

type App struct {
	cfg    config.Config
	db     *sql.DB
	router http.Handler
	bot    *bot.Bot
	log    *slog.Logger
}

func New(ctx context.Context, cfg config.Config, log *slog.Logger) (*App, error) {
	conn, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	if err := db.Migrate(ctx, conn); err != nil {
		conn.Close()
		return nil, err
	}

	userRepo := user.NewRepository(conn)
	categoryRepo := category.NewRepository(conn)
	expenseRepo := expense.NewRepository(conn)

	userService := user.NewService(userRepo, cfg.DefaultCurrency, cfg.DefaultTimezone)
	categoryService := category.NewService(categoryRepo)
	expenseService := expense.NewService(expenseRepo)

	webHandlers := web.NewHandlers(cfg, userService, categoryService, expenseService)
	botHandler := bot.NewHandler(userService, categoryService, expenseService, cfg.AllowedTelegramUserIDs, cfg.DefaultCurrency)

	return &App{
		cfg:    cfg,
		db:     conn,
		router: web.Routes(webHandlers),
		bot:    bot.New(cfg.TelegramBotToken, botHandler, log),
		log:    log,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	server := &http.Server{
		Addr:         a.cfg.AppAddr,
		Handler:      a.router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go a.bot.Run(ctx)

	errCh := make(chan error, 1)
	go func() {
		a.log.Info("web dashboard listening", "addr", a.cfg.AppAddr)
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return a.db.Close()
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return a.db.Close()
		}
		a.db.Close()
		return err
	}
}
