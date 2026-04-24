package handlers

import (
	"database/sql"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/thiago/my_finances_api/internal/middleware"
	"github.com/thiago/my_finances_api/internal/models"
)

func (h *Handler) listBankAccounts(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)

	rows, err := h.db.QueryContext(r.Context(),
		`SELECT id, user_id, name, balance, icon_url, created_at, updated_at
		 FROM bank_accounts WHERE user_id = $1 ORDER BY name`,
		userID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	defer rows.Close()

	accounts := []models.BankAccount{}
	for rows.Next() {
		var a models.BankAccount
		if err := rows.Scan(&a.ID, &a.UserID, &a.Name, &a.Balance, &a.IconURL, &a.CreatedAt, &a.UpdatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		accounts = append(accounts, a)
	}

	writeJSON(w, http.StatusOK, accounts)
}

func (h *Handler) createBankAccount(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)

	var req models.CreateBankAccountRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	var a models.BankAccount
	err := h.db.QueryRowContext(r.Context(),
		`INSERT INTO bank_accounts (user_id, name, balance, icon_url)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, user_id, name, balance, icon_url, created_at, updated_at`,
		userID, req.Name, req.InitialBalance, req.IconURL,
	).Scan(&a.ID, &a.UserID, &a.Name, &a.Balance, &a.IconURL, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusCreated, a)
}

func (h *Handler) getBankAccount(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)
	id := r.PathValue("id")

	a, err := h.fetchBankAccount(r, id, userID)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "bank account not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, a)
}

func (h *Handler) updateBankAccount(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)
	id := r.PathValue("id")

	var req models.UpdateBankAccountRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	var a models.BankAccount
	err := h.db.QueryRowContext(r.Context(),
		`UPDATE bank_accounts SET name=$1, icon_url=$2, updated_at=$3
		 WHERE id=$4 AND user_id=$5
		 RETURNING id, user_id, name, balance, icon_url, created_at, updated_at`,
		req.Name, req.IconURL, time.Now(), id, userID,
	).Scan(&a.ID, &a.UserID, &a.Name, &a.Balance, &a.IconURL, &a.CreatedAt, &a.UpdatedAt)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "bank account not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, a)
}

func (h *Handler) deleteBankAccount(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)
	id := r.PathValue("id")

	res, err := h.db.ExecContext(r.Context(),
		`DELETE FROM bank_accounts WHERE id=$1 AND user_id=$2`, id, userID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		writeError(w, http.StatusNotFound, "bank account not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) getBankAccountStatement(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)
	id := r.PathValue("id")

	account, err := h.fetchBankAccount(r, id, userID)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "bank account not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	page, limit := parsePagination(r)
	offset := (page - 1) * limit

	var total int
	h.db.QueryRowContext(r.Context(),
		`SELECT COUNT(*) FROM transactions WHERE bank_account_id=$1 AND user_id=$2`,
		id, userID,
	).Scan(&total)

	rows, err := h.db.QueryContext(r.Context(),
		`SELECT id, user_id, bank_account_id, name, value, operation, date, created_at, updated_at
		 FROM transactions WHERE bank_account_id=$1 AND user_id=$2
		 ORDER BY date DESC LIMIT $3 OFFSET $4`,
		id, userID, limit, offset,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	defer rows.Close()

	txns := []models.Transaction{}
	for rows.Next() {
		var t models.Transaction
		if err := rows.Scan(&t.ID, &t.UserID, &t.BankAccountID, &t.Name, &t.Value,
			&t.Operation, &t.Date, &t.CreatedAt, &t.UpdatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		t.Tags = h.fetchTransactionTags(r, t.ID)
		txns = append(txns, t)
	}

	upcoming := h.buildUpcoming(r, id, userID)

	resp := models.StatementResponse{
		Account: account,
		Transactions: models.PaginatedResponse[models.Transaction]{
			Data:  txns,
			Total: total,
			Page:  page,
			Limit: limit,
			Pages: pages(total, limit),
		},
		Upcoming: upcoming,
	}

	writeJSON(w, http.StatusOK, resp)
}

// buildUpcoming returns upcoming income, expense, and goal transfer items for the next 90 days.
func (h *Handler) buildUpcoming(r *http.Request, accountID, userID string) []models.UpcomingItem {
	var items []models.UpcomingItem
	now := time.Now()
	horizon := now.AddDate(0, 3, 0) // 3 months out

	// Incomes
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT name, value, repeatable_day FROM incomes
		 WHERE bank_account_id=$1 AND user_id=$2`, accountID, userID,
	)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var name string
			var value float64
			var day int
			rows.Scan(&name, &value, &day)
			for _, d := range nextOccurrences(day, now, horizon) {
				items = append(items, models.UpcomingItem{
					Type: "income", Name: name, Value: value,
					Operation: "add", Date: d, Source: name,
				})
			}
		}
	}

	// Expenses
	rows2, err := h.db.QueryContext(r.Context(),
		`SELECT name, value, repeatable_day FROM expenses
		 WHERE bank_account_id=$1 AND user_id=$2`, accountID, userID,
	)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var name string
			var value float64
			var day int
			rows2.Scan(&name, &value, &day)
			for _, d := range nextOccurrences(day, now, horizon) {
				items = append(items, models.UpcomingItem{
					Type: "expense", Name: name, Value: value,
					Operation: "subtract", Date: d, Source: name,
				})
			}
		}
	}

	// Goal transfers where this account is source
	rows3, err := h.db.QueryContext(r.Context(),
		`SELECT name, target_value, start_date, end_date, interval_days
		 FROM goals WHERE source_account_id=$1 AND user_id=$2`,
		accountID, userID,
	)
	if err == nil {
		defer rows3.Close()
		for rows3.Next() {
			var name string
			var targetValue float64
			var start, end time.Time
			var intervalDays int
			rows3.Scan(&name, &targetValue, &start, &end, &intervalDays)
			if intervalDays <= 0 {
				continue
			}
			totalIntervals := math.Ceil(end.Sub(start).Hours() / 24 / float64(intervalDays))
			if totalIntervals <= 0 {
				continue
			}
			amountPerTransfer := targetValue / totalIntervals
			for d := start; !d.After(end) && !d.After(horizon); d = d.AddDate(0, 0, intervalDays) {
				if d.After(now) {
					items = append(items, models.UpcomingItem{
						Type: "goal_transfer", Name: name,
						Value: amountPerTransfer, Operation: "subtract",
						Date: d, Source: name,
					})
				}
			}
		}
	}

	return items
}

// nextOccurrences returns dates matching the given day-of-month from now up to horizon.
func nextOccurrences(day int, from, until time.Time) []time.Time {
	var out []time.Time
	// Start from current month, find first occurrence on or after `from`
	y, m, _ := from.Date()
	for t := time.Date(y, m, 1, 0, 0, 0, 0, time.UTC); !t.After(until); t = t.AddDate(0, 1, 0) {
		// Clamp day to last day of month
		lastDay := time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
		d := day
		if d > lastDay {
			d = lastDay
		}
		occurrence := time.Date(t.Year(), t.Month(), d, 0, 0, 0, 0, time.UTC)
		if occurrence.After(from) && !occurrence.After(until) {
			out = append(out, occurrence)
		}
	}
	return out
}

func (h *Handler) fetchBankAccount(r *http.Request, id, userID string) (models.BankAccount, error) {
	var a models.BankAccount
	err := h.db.QueryRowContext(r.Context(),
		`SELECT id, user_id, name, balance, icon_url, created_at, updated_at
		 FROM bank_accounts WHERE id=$1 AND user_id=$2`, id, userID,
	).Scan(&a.ID, &a.UserID, &a.Name, &a.Balance, &a.IconURL, &a.CreatedAt, &a.UpdatedAt)
	return a, err
}
