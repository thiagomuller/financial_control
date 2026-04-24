package handlers

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/thiago/my_finances_api/internal/middleware"
	"github.com/thiago/my_finances_api/internal/models"
)

func (h *Handler) listIncomes(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)

	rows, err := h.db.QueryContext(r.Context(),
		`SELECT id, user_id, bank_account_id, name, value, repeatable_day, last_executed_at, created_at, updated_at
		 FROM incomes WHERE user_id=$1 ORDER BY name`, userID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	defer rows.Close()

	incomes := []models.Income{}
	for rows.Next() {
		var i models.Income
		if err := rows.Scan(&i.ID, &i.UserID, &i.BankAccountID, &i.Name, &i.Value,
			&i.RepeatableDay, &i.LastExecutedAt, &i.CreatedAt, &i.UpdatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		incomes = append(incomes, i)
	}

	writeJSON(w, http.StatusOK, incomes)
}

func (h *Handler) createIncome(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)

	var req models.CreateIncomeRequest
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

	var i models.Income
	err := h.db.QueryRowContext(r.Context(),
		`INSERT INTO incomes (user_id, bank_account_id, name, value, repeatable_day)
		 VALUES ($1,$2,$3,$4,$5)
		 RETURNING id, user_id, bank_account_id, name, value, repeatable_day, last_executed_at, created_at, updated_at`,
		userID, req.BankAccountID, req.Name, req.Value, req.RepeatableDay,
	).Scan(&i.ID, &i.UserID, &i.BankAccountID, &i.Name, &i.Value,
		&i.RepeatableDay, &i.LastExecutedAt, &i.CreatedAt, &i.UpdatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusCreated, i)
}

func (h *Handler) updateIncome(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)
	id := r.PathValue("id")

	var req models.UpdateIncomeRequest
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

	var i models.Income
	err := h.db.QueryRowContext(r.Context(),
		`UPDATE incomes SET name=$1, value=$2, bank_account_id=$3, repeatable_day=$4, updated_at=$5
		 WHERE id=$6 AND user_id=$7
		 RETURNING id, user_id, bank_account_id, name, value, repeatable_day, last_executed_at, created_at, updated_at`,
		req.Name, req.Value, req.BankAccountID, req.RepeatableDay, time.Now(), id, userID,
	).Scan(&i.ID, &i.UserID, &i.BankAccountID, &i.Name, &i.Value,
		&i.RepeatableDay, &i.LastExecutedAt, &i.CreatedAt, &i.UpdatedAt)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "income not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, i)
}

func (h *Handler) deleteIncome(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)
	id := r.PathValue("id")

	res, err := h.db.ExecContext(r.Context(),
		`DELETE FROM incomes WHERE id=$1 AND user_id=$2`, id, userID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		writeError(w, http.StatusNotFound, "income not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
