package handlers

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/thiago/my_finances_api/internal/middleware"
	"github.com/thiago/my_finances_api/internal/models"
)

func (h *Handler) listExpenses(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)

	rows, err := h.db.QueryContext(r.Context(),
		`SELECT id, user_id, bank_account_id, name, value, repeatable_day, last_executed_at, created_at, updated_at
		 FROM expenses WHERE user_id=$1 ORDER BY name`, userID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	defer rows.Close()

	expenses := []models.Expense{}
	for rows.Next() {
		var e models.Expense
		if err := rows.Scan(&e.ID, &e.UserID, &e.BankAccountID, &e.Name, &e.Value,
			&e.RepeatableDay, &e.LastExecutedAt, &e.CreatedAt, &e.UpdatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		expenses = append(expenses, e)
	}

	writeJSON(w, http.StatusOK, expenses)
}

func (h *Handler) createExpense(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)

	var req models.CreateExpenseRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || req.BankAccountID == "" {
		writeError(w, http.StatusBadRequest, "name and bank_account_id are required")
		return
	}
	if req.RepeatableDay < 1 || req.RepeatableDay > 31 {
		writeError(w, http.StatusBadRequest, "repeatable_day must be between 1 and 31")
		return
	}
	if req.Value <= 0 {
		writeError(w, http.StatusBadRequest, "value must be positive")
		return
	}

	var e models.Expense
	err := h.db.QueryRowContext(r.Context(),
		`INSERT INTO expenses (user_id, bank_account_id, name, value, repeatable_day)
		 VALUES ($1,$2,$3,$4,$5)
		 RETURNING id, user_id, bank_account_id, name, value, repeatable_day, last_executed_at, created_at, updated_at`,
		userID, req.BankAccountID, req.Name, req.Value, req.RepeatableDay,
	).Scan(&e.ID, &e.UserID, &e.BankAccountID, &e.Name, &e.Value,
		&e.RepeatableDay, &e.LastExecutedAt, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusCreated, e)
}

func (h *Handler) updateExpense(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)
	id := r.PathValue("id")

	var req models.UpdateExpenseRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || req.BankAccountID == "" {
		writeError(w, http.StatusBadRequest, "name and bank_account_id are required")
		return
	}
	if req.RepeatableDay < 1 || req.RepeatableDay > 31 {
		writeError(w, http.StatusBadRequest, "repeatable_day must be between 1 and 31")
		return
	}

	var e models.Expense
	err := h.db.QueryRowContext(r.Context(),
		`UPDATE expenses SET name=$1, value=$2, bank_account_id=$3, repeatable_day=$4, updated_at=$5
		 WHERE id=$6 AND user_id=$7
		 RETURNING id, user_id, bank_account_id, name, value, repeatable_day, last_executed_at, created_at, updated_at`,
		req.Name, req.Value, req.BankAccountID, req.RepeatableDay, time.Now(), id, userID,
	).Scan(&e.ID, &e.UserID, &e.BankAccountID, &e.Name, &e.Value,
		&e.RepeatableDay, &e.LastExecutedAt, &e.CreatedAt, &e.UpdatedAt)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "expense not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, e)
}

func (h *Handler) deleteExpense(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)
	id := r.PathValue("id")

	res, err := h.db.ExecContext(r.Context(),
		`DELETE FROM expenses WHERE id=$1 AND user_id=$2`, id, userID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		writeError(w, http.StatusNotFound, "expense not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
