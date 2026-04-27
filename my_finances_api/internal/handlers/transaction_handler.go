package handlers

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/thiago/my_finances_api/internal/middleware"
	"github.com/thiago/my_finances_api/internal/models"
)

func (h *Handler) listTransactions(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)
	page, limit := parsePagination(r)
	offset := (page - 1) * limit
	accountID := r.URL.Query().Get("bank_account_id")

	var (
		total int
		rows  *sql.Rows
		err   error
	)

	if accountID != "" {
		h.db.QueryRowContext(r.Context(),
			`SELECT COUNT(*) FROM transactions WHERE user_id=$1 AND bank_account_id=$2`,
			userID, accountID,
		).Scan(&total)
		rows, err = h.db.QueryContext(r.Context(),
			`SELECT id, user_id, bank_account_id, name, value, operation, date,
			        is_repeatable, repeatable_day, last_executed_at, created_at, updated_at
			 FROM transactions WHERE user_id=$1 AND bank_account_id=$2
			 ORDER BY date DESC LIMIT $3 OFFSET $4`,
			userID, accountID, limit, offset,
		)
	} else {
		h.db.QueryRowContext(r.Context(),
			`SELECT COUNT(*) FROM transactions WHERE user_id=$1`, userID,
		).Scan(&total)
		rows, err = h.db.QueryContext(r.Context(),
			`SELECT id, user_id, bank_account_id, name, value, operation, date,
			        is_repeatable, repeatable_day, last_executed_at, created_at, updated_at
			 FROM transactions WHERE user_id=$1
			 ORDER BY date DESC LIMIT $2 OFFSET $3`,
			userID, limit, offset,
		)
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	defer rows.Close()

	txns := []models.Transaction{}
	for rows.Next() {
		var t models.Transaction
		if err := rows.Scan(&t.ID, &t.UserID, &t.BankAccountID, &t.Name, &t.Value,
			&t.Operation, &t.Date, &t.IsRepeatable, &t.RepeatableDay, &t.LastExecutedAt,
			&t.CreatedAt, &t.UpdatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		t.Tags = h.fetchTransactionTags(r, t.ID)
		txns = append(txns, t)
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse[models.Transaction]{
		Data: txns, Total: total, Page: page, Limit: limit, Pages: pages(total, limit),
	})
}

func (h *Handler) createTransaction(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)

	var req models.CreateTransactionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || req.BankAccountID == "" {
		writeError(w, http.StatusBadRequest, "name and bank_account_id are required")
		return
	}
	if req.Operation != "add" && req.Operation != "subtract" {
		writeError(w, http.StatusBadRequest, "operation must be 'add' or 'subtract'")
		return
	}
	if req.Value <= 0 {
		writeError(w, http.StatusBadRequest, "value must be positive")
		return
	}
	if req.IsRepeatable {
		if req.RepeatableDay == nil || *req.RepeatableDay < 1 || *req.RepeatableDay > 31 {
			writeError(w, http.StatusBadRequest, "repeatable_day (1-31) is required for repeatable transactions")
			return
		}
	}
	if req.Date.IsZero() {
		req.Date = time.Now()
	}

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	defer tx.Rollback()

	var balance float64
	err = tx.QueryRowContext(r.Context(),
		`SELECT balance FROM bank_accounts WHERE id=$1 AND user_id=$2`,
		req.BankAccountID, userID,
	).Scan(&balance)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "bank account not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	var t models.Transaction
	err = tx.QueryRowContext(r.Context(),
		`INSERT INTO transactions (user_id, bank_account_id, name, value, operation, date, is_repeatable, repeatable_day)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		 RETURNING id, user_id, bank_account_id, name, value, operation, date,
		           is_repeatable, repeatable_day, last_executed_at, created_at, updated_at`,
		userID, req.BankAccountID, req.Name, req.Value, req.Operation, req.Date,
		req.IsRepeatable, req.RepeatableDay,
	).Scan(&t.ID, &t.UserID, &t.BankAccountID, &t.Name, &t.Value,
		&t.Operation, &t.Date, &t.IsRepeatable, &t.RepeatableDay, &t.LastExecutedAt,
		&t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if !req.IsRepeatable {
		if req.Operation == "add" {
			_, err = tx.ExecContext(r.Context(),
				`UPDATE bank_accounts SET balance=balance+$1, updated_at=NOW() WHERE id=$2`,
				req.Value, req.BankAccountID)
		} else {
			_, err = tx.ExecContext(r.Context(),
				`UPDATE bank_accounts SET balance=balance-$1, updated_at=NOW() WHERE id=$2`,
				req.Value, req.BankAccountID)
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
	}

	tagIDs := req.TagIDs
	if req.IsRepeatable {
		systemTagName := "Income"
		if req.Operation == "subtract" {
			systemTagName = "Expense"
		}
		var systemTagID string
		tx.QueryRowContext(r.Context(),
			`SELECT id FROM tags WHERE is_system=true AND name=$1`, systemTagName,
		).Scan(&systemTagID)
		if systemTagID != "" {
			tagIDs = appendUnique(tagIDs, systemTagID)
		}
	}

	for _, tagID := range tagIDs {
		tx.ExecContext(r.Context(),
			`INSERT INTO transaction_tags (transaction_id, tag_id) VALUES ($1,$2)
			 ON CONFLICT DO NOTHING`, t.ID, tagID)
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	t.Tags = h.fetchTransactionTags(r, t.ID)

	if req.IsRepeatable {
		warning := h.projectedTransactionWarning(r, userID, req.BankAccountID, req.Value, req.Operation, t.ID)
		writeJSON(w, http.StatusCreated, models.CreateTransactionResponse{
			Transaction:             t,
			ProjectedBalanceWarning: warning,
		})
		return
	}

	writeJSON(w, http.StatusCreated, t)
}

func (h *Handler) projectedTransactionWarning(r *http.Request, userID, accountID string, newValue float64, newOp string, excludeID string) string {
	var addSum, subSum float64
	h.db.QueryRowContext(r.Context(),
		`SELECT COALESCE(SUM(CASE WHEN operation='add' THEN value ELSE 0 END), 0),
		        COALESCE(SUM(CASE WHEN operation='subtract' THEN value ELSE 0 END), 0)
		 FROM transactions
		 WHERE bank_account_id=$1 AND user_id=$2 AND is_repeatable=true AND id!=$3`,
		accountID, userID, excludeID,
	).Scan(&addSum, &subSum)

	var transferInSum, transferOutSum float64
	h.db.QueryRowContext(r.Context(),
		`SELECT
		   COALESCE(SUM(CASE WHEN target_account_id=$1 THEN value ELSE 0 END), 0),
		   COALESCE(SUM(CASE WHEN source_account_id=$1 THEN value ELSE 0 END), 0)
		 FROM transfers
		 WHERE (source_account_id=$1 OR target_account_id=$1) AND user_id=$2 AND is_repeatable=true`,
		accountID, userID,
	).Scan(&transferInSum, &transferOutSum)

	net := addSum - subSum + transferInSum - transferOutSum
	if newOp == "subtract" {
		net -= newValue
	} else {
		net += newValue
	}

	if net < 0 {
		return "Projected monthly balance is negative. Review your scheduled transactions and transfers."
	}
	return ""
}

func appendUnique(ids []string, id string) []string {
	for _, existing := range ids {
		if existing == id {
			return ids
		}
	}
	return append(ids, id)
}

func (h *Handler) updateTransaction(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)
	id := r.PathValue("id")

	var req models.UpdateTransactionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Operation != "add" && req.Operation != "subtract" {
		writeError(w, http.StatusBadRequest, "operation must be 'add' or 'subtract'")
		return
	}
	if req.Value <= 0 {
		writeError(w, http.StatusBadRequest, "value must be positive")
		return
	}

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	defer tx.Rollback()

	var oldValue float64
	var oldOp, oldAccountID string
	var isRepeatable bool
	err = tx.QueryRowContext(r.Context(),
		`SELECT value, operation, bank_account_id, is_repeatable FROM transactions WHERE id=$1 AND user_id=$2`,
		id, userID,
	).Scan(&oldValue, &oldOp, &oldAccountID, &isRepeatable)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "transaction not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if !isRepeatable {
		if oldOp == "add" {
			tx.ExecContext(r.Context(),
				`UPDATE bank_accounts SET balance=balance-$1, updated_at=NOW() WHERE id=$2`,
				oldValue, oldAccountID)
		} else {
			tx.ExecContext(r.Context(),
				`UPDATE bank_accounts SET balance=balance+$1, updated_at=NOW() WHERE id=$2`,
				oldValue, oldAccountID)
		}

		if req.Operation == "add" {
			tx.ExecContext(r.Context(),
				`UPDATE bank_accounts SET balance=balance+$1, updated_at=NOW() WHERE id=$2`,
				req.Value, oldAccountID)
		} else {
			tx.ExecContext(r.Context(),
				`UPDATE bank_accounts SET balance=balance-$1, updated_at=NOW() WHERE id=$2`,
				req.Value, oldAccountID)
		}
	}

	var t models.Transaction
	err = tx.QueryRowContext(r.Context(),
		`UPDATE transactions SET name=$1, value=$2, operation=$3, date=$4, updated_at=NOW()
		 WHERE id=$5 AND user_id=$6
		 RETURNING id, user_id, bank_account_id, name, value, operation, date,
		           is_repeatable, repeatable_day, last_executed_at, created_at, updated_at`,
		req.Name, req.Value, req.Operation, req.Date, id, userID,
	).Scan(&t.ID, &t.UserID, &t.BankAccountID, &t.Name, &t.Value,
		&t.Operation, &t.Date, &t.IsRepeatable, &t.RepeatableDay, &t.LastExecutedAt,
		&t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	tx.ExecContext(r.Context(), `DELETE FROM transaction_tags WHERE transaction_id=$1`, id)
	for _, tagID := range req.TagIDs {
		tx.ExecContext(r.Context(),
			`INSERT INTO transaction_tags (transaction_id, tag_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
			t.ID, tagID)
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	t.Tags = h.fetchTransactionTags(r, t.ID)
	writeJSON(w, http.StatusOK, t)
}

func (h *Handler) deleteTransaction(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)
	id := r.PathValue("id")

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	defer tx.Rollback()

	var value float64
	var op, accountID string
	var isRepeatable bool
	err = tx.QueryRowContext(r.Context(),
		`SELECT value, operation, bank_account_id, is_repeatable FROM transactions WHERE id=$1 AND user_id=$2`,
		id, userID,
	).Scan(&value, &op, &accountID, &isRepeatable)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "transaction not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	tx.ExecContext(r.Context(), `DELETE FROM transactions WHERE id=$1`, id)

	if !isRepeatable {
		if op == "add" {
			tx.ExecContext(r.Context(),
				`UPDATE bank_accounts SET balance=balance-$1, updated_at=NOW() WHERE id=$2`,
				value, accountID)
		} else {
			tx.ExecContext(r.Context(),
				`UPDATE bank_accounts SET balance=balance+$1, updated_at=NOW() WHERE id=$2`,
				value, accountID)
		}
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
