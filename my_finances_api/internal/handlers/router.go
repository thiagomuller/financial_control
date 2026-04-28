package handlers

import (
	"database/sql"
	"net/http"

	"github.com/thiago/my_finances_api/internal/config"
	"github.com/thiago/my_finances_api/internal/middleware"
)

type Handler struct {
	db  *sql.DB
	cfg *config.Config
}

func NewRouter(db *sql.DB, cfg *config.Config) http.Handler {
	h := &Handler{db: db, cfg: cfg}
	protect := middleware.Auth(cfg.JWTSecret)

	mux := http.NewServeMux()

	// Public auth routes
	mux.HandleFunc("POST /api/auth/register", h.register)
	mux.HandleFunc("POST /api/auth/login", h.login)

	// User profile
	mux.HandleFunc("GET /api/users/me", protect(h.getMe))
	mux.HandleFunc("PUT /api/users/me", protect(h.updateMe))

	// Bank accounts
	mux.HandleFunc("GET /api/bank-accounts/{id}/statement", protect(h.getBankAccountStatement))
	mux.HandleFunc("GET /api/bank-accounts", protect(h.listBankAccounts))
	mux.HandleFunc("POST /api/bank-accounts", protect(h.createBankAccount))
	mux.HandleFunc("GET /api/bank-accounts/{id}", protect(h.getBankAccount))
	mux.HandleFunc("PUT /api/bank-accounts/{id}", protect(h.updateBankAccount))
	mux.HandleFunc("DELETE /api/bank-accounts/{id}", protect(h.deleteBankAccount))

	// Tags
	mux.HandleFunc("GET /api/tags", protect(h.listTags))
	mux.HandleFunc("POST /api/tags", protect(h.createTag))
	mux.HandleFunc("PUT /api/tags/{id}", protect(h.updateTag))
	mux.HandleFunc("DELETE /api/tags/{id}", protect(h.deleteTag))

	// Transactions
	mux.HandleFunc("GET /api/transactions", protect(h.listTransactions))
	mux.HandleFunc("POST /api/transactions", protect(h.createTransaction))
	mux.HandleFunc("PUT /api/transactions/{id}", protect(h.updateTransaction))
	mux.HandleFunc("DELETE /api/transactions/{id}", protect(h.deleteTransaction))

	// Transfers
	mux.HandleFunc("GET /api/transfers", protect(h.listTransfers))
	mux.HandleFunc("POST /api/transfers", protect(h.createTransfer))
	mux.HandleFunc("PUT /api/transfers/{id}", protect(h.updateTransfer))
	mux.HandleFunc("DELETE /api/transfers/{id}", protect(h.deleteTransfer))

	// Goals
	mux.HandleFunc("GET /api/goals", protect(h.listGoals))
	mux.HandleFunc("POST /api/goals", protect(h.createGoal))
	mux.HandleFunc("PUT /api/goals/{id}", protect(h.updateGoal))
	mux.HandleFunc("DELETE /api/goals/{id}", protect(h.deleteGoal))

	// Dashboard
	mux.HandleFunc("GET /api/dashboard", protect(h.getDashboard))

	return middleware.CORS(mux)
}
