package web

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Routes(handlers *Handlers) http.Handler {
	r := chi.NewRouter()
	r.Get("/login", handlers.LoginForm)
	r.Post("/login", handlers.Login)

	r.Group(func(r chi.Router) {
		r.Use(handlers.RequireAuth)
		r.Get("/", handlers.Dashboard)
		r.Get("/transactions", handlers.Transactions)
		r.Get("/transactions/{id}/edit", handlers.EditTransactionForm)
		r.Post("/transactions/{id}", handlers.UpdateTransaction)
		r.Get("/categories", handlers.Categories)
		r.Post("/categories", handlers.CreateCategory)
		r.Post("/logout", handlers.Logout)
	})

	return r
}
