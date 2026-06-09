package web

import (
	"context"
	"database/sql"
	"embed"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/CHashtager/khazaneh/internal/calendar"
	"github.com/CHashtager/khazaneh/internal/category"
	"github.com/CHashtager/khazaneh/internal/config"
	"github.com/CHashtager/khazaneh/internal/expense"
	"github.com/CHashtager/khazaneh/internal/user"

	"github.com/go-chi/chi/v5"
)

//go:embed templates/*.html
var templateFS embed.FS

type Handlers struct {
	cfg        config.Config
	users      *user.Service
	categories *category.Service
	expenses   *expense.Service
}

type pageData struct {
	Title        string
	User         *user.User
	Transactions []expense.Transaction
	Transaction  *expense.Transaction
	Categories   []category.Category
	Summary      expense.MonthlySummary
	CurrentYM    string
	SelectedID   int64
	Error        string
}

func NewHandlers(cfg config.Config, users *user.Service, categories *category.Service, expenses *expense.Service) *Handlers {
	return &Handlers{cfg: cfg, users: users, categories: categories, expenses: expenses}
}

func (h *Handlers) LoginForm(w http.ResponseWriter, r *http.Request) {
	h.render(w, "login", pageData{Title: "Login"})
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.render(w, "login", pageData{Title: "Login", Error: "Invalid form."})
		return
	}
	if r.FormValue("username") != h.cfg.DashboardUsername || r.FormValue("password") != h.cfg.DashboardPassword {
		h.render(w, "login", pageData{Title: "Login", Error: "Invalid username or password."})
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    h.newSessionValue(h.cfg.DashboardUsername),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 30,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: sessionCookieName, Value: "", Path: "/", MaxAge: -1})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *Handlers) Dashboard(w http.ResponseWriter, r *http.Request) {
	u, err := h.currentUser(r.Context())
	if err != nil {
		h.render(w, "dashboard", pageData{Title: "Dashboard", Error: "No Telegram user has been created yet. Send /start to the bot first."})
		return
	}

	location, _ := time.LoadLocation(u.Timezone)
	j := calendar.FromTime(time.Now(), location)
	currentYM := j.YearMonth()
	transactions, _ := h.expenses.ListRecent(r.Context(), u.ID, 10)
	summary, _ := h.expenses.Summary(r.Context(), u.ID, currentYM)

	h.render(w, "dashboard", pageData{
		Title:        "Dashboard",
		User:         &u,
		Transactions: transactions,
		Summary:      summary,
		CurrentYM:    currentYM,
	})
}

func (h *Handlers) Transactions(w http.ResponseWriter, r *http.Request) {
	u, err := h.currentUser(r.Context())
	if err != nil {
		h.render(w, "transactions", pageData{Title: "Transactions", Error: "No Telegram user has been created yet."})
		return
	}
	transactions, err := h.expenses.ListRecent(r.Context(), u.ID, 100)
	if err != nil {
		h.render(w, "transactions", pageData{Title: "Transactions", User: &u, Error: "Could not load transactions."})
		return
	}
	h.render(w, "transactions", pageData{Title: "Transactions", User: &u, Transactions: transactions})
}

func (h *Handlers) EditTransactionForm(w http.ResponseWriter, r *http.Request) {
	u, err := h.currentUser(r.Context())
	if err != nil {
		h.render(w, "edit_transaction", pageData{Title: "Edit Transaction", Error: "No Telegram user has been created yet."})
		return
	}
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	tx, err := h.expenses.Get(r.Context(), u.ID, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	selectedID := int64(0)
	if tx.CategoryID != nil {
		selectedID = *tx.CategoryID
	}
	categories, _ := h.categories.List(r.Context(), u.ID)
	h.render(w, "edit_transaction", pageData{
		Title:       "Edit Transaction",
		User:        &u,
		Transaction: &tx,
		Categories:  categories,
		SelectedID:  selectedID,
	})
}

func (h *Handlers) UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	u, err := h.currentUser(r.Context())
	if err != nil {
		http.Error(w, "No Telegram user has been created yet.", http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form.", http.StatusBadRequest)
		return
	}
	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil || amount <= 0 {
		http.Error(w, "Amount must be greater than zero.", http.StatusBadRequest)
		return
	}
	var categoryID *int64
	if raw := r.FormValue("category_id"); raw != "" {
		parsedID, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			http.Error(w, "Invalid category.", http.StatusBadRequest)
			return
		}
		categoryID = &parsedID
	}
	if _, err := h.expenses.Update(r.Context(), u.ID, id, categoryID, r.FormValue("kind"), r.FormValue("amount"), r.FormValue("merchant"), r.FormValue("note")); err != nil {
		http.Error(w, "Could not update transaction.", http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/transactions", http.StatusSeeOther)
}

func (h *Handlers) Categories(w http.ResponseWriter, r *http.Request) {
	u, err := h.currentUser(r.Context())
	if err != nil {
		h.render(w, "categories", pageData{Title: "Categories", Error: "No Telegram user has been created yet."})
		return
	}
	categories, err := h.categories.List(r.Context(), u.ID)
	if err != nil {
		h.render(w, "categories", pageData{Title: "Categories", User: &u, Error: "Could not load categories."})
		return
	}
	h.render(w, "categories", pageData{Title: "Categories", User: &u, Categories: categories})
}

func (h *Handlers) CreateCategory(w http.ResponseWriter, r *http.Request) {
	u, err := h.currentUser(r.Context())
	if err != nil {
		http.Error(w, "No Telegram user has been created yet.", http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form.", http.StatusBadRequest)
		return
	}
	if _, err := h.categories.Create(r.Context(), u.ID, r.FormValue("name"), r.FormValue("kind"), r.FormValue("emoji")); err != nil {
		http.Error(w, "Could not create category.", http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/categories", http.StatusSeeOther)
}

func (h *Handlers) currentUser(ctx context.Context) (user.User, error) {
	u, err := h.users.First(ctx)
	if err == sql.ErrNoRows {
		return user.User{}, err
	}
	return u, err
}

func (h *Handlers) render(w http.ResponseWriter, name string, data pageData) {
	tmpl, err := template.ParseFS(templateFS, "templates/layout.html", "templates/transaction_table.html", "templates/"+name+".html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
