package handlers

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/thiago/my_finances_api/internal/middleware"
	"github.com/thiago/my_finances_api/internal/models"
)

func (h *Handler) listTransfers(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)
	page, limit := parsePagination(r)
	offset := (page - 1) * limit

	var total int
	h.db.QueryRowContext(r.Context(),
		`SELECT COUNT(*) FROM transfers WHERE user_id=$1`, userID,
	).Scan(&total)

	rows, err := h.db.QueryContext(r.Context(),
		`SELECT id, user_id, source_account_id, target_account_id, name, value, date,
		        is_repeatable, repeatable_day, last_executed_at, created_at, updated_at
		 FROM transfers WHERE user_id=$1 ORDER BY date DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	defer rows.Close()

	transfers := []models.Transfer{}
	for rows.Next() {
		var t models.Transfer
		if err := rows.Scan(&t.ID, &t.UserID, &t.SourceAccountID, &t.TargetAccountID,
			&t.Name, &t.Value, &t.Date, &t.IsRepeatable, &t.RepeatableDay, &t.LastExecutedAt,
			&t.CreatedAt, &t.UpdatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		t.Tags = h.fetchTransferTags(r, t.ID)
		transfers = append(transfers, t)
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse[models.Transfer]{
		Data: transfers, Total: total, Page: page, Limit: limit, Pages: pages(total, limit),
	})
}

func (h *Handler) createTransfer(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)

	var req models.CreateTransferRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || req.SourceAccountID == "" || req.TargetAccountID == "" {
		writeError(w, http.StatusBadRequest, "name, source_account_id and target_account_id are required")
		return
	}
	if req.Value <= 0 {
		writeError(w, http.StatusBadRequest, "value must be positive")
		return
	}
	if req.SourceAccountID == req.TargetAccountID {
		writeError(w, http.StatusBadRequest, "source and target accounts must differ")
		return
	}
	if req.IsRepeatable {
		if req.RepeatableDay == nil || *req.RepeatableDay < 1 || *req.RepeatableDay > 31 {
			writeError(w, http.StatusBadRequest, "repeatable_day (1-31) is required for repeatable transfers")
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

	var sourceBalance float64
	err = tx.QueryRowContext(r.Context(),
		`SELECT balance FROM bank_accounts WHERE id=$1 AND user_id=$2 FOR UPDATE`,
		req.SourceAccountID, userID,
	).Scan(&sourceBalance)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "source bank account not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if !req.IsRepeatable && sourceBalance < req.Value {
		writeError(w, http.StatusBadRequest, "insufficient balance in source account")
		return
	}

	var targetExists bool
	tx.QueryRowContext(r.Context(),
		`SELECT EXISTS(SELECT 1 FROM bank_accounts WHERE id=$1 AND user_id=$2)`,
		req.TargetAccountID, userID,
	).Scan(&targetExists)
	if !targetExists {
		writeError(w, http.StatusNotFound, "target bank account not found")
		return
	}

	var t models.Transfer
	err = tx.QueryRowContext(r.Context(),
		`INSERT INTO transfers (user_id, source_account_id, target_account_id, name, value, date, is_repeatable, repeatable_day)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		 RETURNING id, user_id, source_account_id, target_account_id, name, value, date,
		           is_repeatable, repeatable_day, last_executed_at, created_at, updated_at`,
		userID, req.SourceAccountID, req.TargetAccountID, req.Name, req.Value, req.Date,
		req.IsRepeatable, req.RepeatableDay,
	).Scan(&t.ID, &t.UserID, &t.SourceAccountID, &t.TargetAccountID,
		&t.Name, &t.Value, &t.Date, &t.IsRepeatable, &t.RepeatableDay, &t.LastExecutedAt,
		&t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if !req.IsRepeatable {
		tx.ExecContext(r.Context(),
			`UPDATE bank_accounts SET balance=balance-$1, updated_at=NOW() WHERE id=$2`,
			req.Value, req.SourceAccountID)
		tx.ExecContext(r.Context(),
			`UPDATE bank_accounts SET balance=balance+$1, updated_at=NOW() WHERE id=$2`,
			req.Value, req.TargetAccountID)
	}

	for _, tagID := range req.TagIDs {
		tx.ExecContext(r.Context(),
			`INSERT INTO transfer_tags (transfer_id, tag_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
			t.ID, tagID)
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	t.Tags = h.fetchTransferTags(r, t.ID)

	if req.IsRepeatable {
		warning := h.projectedTransferWarning(r, userID, req.SourceAccountID, req.Value, t.ID)
		writeJSON(w, http.StatusCreated, models.CreateTransferResponse{
			Transfer:                t,
			ProjectedBalanceWarning: warning,
		})
		return
	}

	writeJSON(w, http.StatusCreated, t)
}

func (h *Handler) projectedTransferWarning(r *http.Request, userID, sourceAccountID string, newValue float64, excludeID string) string {
	var addSum, subSum float64
	h.db.QueryRowContext(r.Context(),
		`SELECT COALESCE(SUM(CASE WHEN operation='add' THEN value ELSE 0 END), 0),
		        COALESCE(SUM(CASE WHEN operation='subtract' THEN value ELSE 0 END), 0)
		 FROM transactions
		 WHERE bank_account_id=$1 AND user_id=$2 AND is_repeatable=true`,
		sourceAccountID, userID,
	).Scan(&addSum, &subSum)

	var transferInSum, transferOutSum float64
	h.db.QueryRowContext(r.Context(),
		`SELECT
		   COALESCE(SUM(CASE WHEN target_account_id=$1 THEN value ELSE 0 END), 0),
		   COALESCE(SUM(CASE WHEN source_account_id=$1 THEN value ELSE 0 END), 0)
		 FROM transfers
		 WHERE (source_account_id=$1 OR target_account_id=$1) AND user_id=$2 AND is_repeatable=true AND id!=$3`,
		sourceAccountID, userID, excludeID,
	).Scan(&transferInSum, &transferOutSum)

	net := addSum - subSum + transferInSum - transferOutSum - newValue
	if net < 0 {
		return "Projected monthly balance is negative. Review your scheduled transactions and transfers."
	}
	return ""
}

func (h *Handler) updateTransfer(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)
	id := r.PathValue("id")

	var req models.UpdateTransferRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	defer tx.Rollback()

	var t models.Transfer
	err = tx.QueryRowContext(r.Context(),
		`UPDATE transfers SET name=$1, date=$2, updated_at=NOW()
		 WHERE id=$3 AND user_id=$4
		 RETURNING id, user_id, source_account_id, target_account_id, name, value, date,
		           is_repeatable, repeatable_day, last_executed_at, created_at, updated_at`,
		req.Name, req.Date, id, userID,
	).Scan(&t.ID, &t.UserID, &t.SourceAccountID, &t.TargetAccountID,
		&t.Name, &t.Value, &t.Date, &t.IsRepeatable, &t.RepeatableDay, &t.LastExecutedAt,
		&t.CreatedAt, &t.UpdatedAt)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "transfer not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	tx.ExecContext(r.Context(), `DELETE FROM transfer_tags WHERE transfer_id=$1`, id)
	for _, tagID := range req.TagIDs {
		tx.ExecContext(r.Context(),
			`INSERT INTO transfer_tags (transfer_id, tag_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
			t.ID, tagID)
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	t.Tags = h.fetchTransferTags(r, t.ID)
	writeJSON(w, http.StatusOK, t)
}

func (h *Handler) deleteTransfer(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)
	id := r.PathValue("id")

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	defer tx.Rollback()

	var value float64
	var srcID, tgtID string
	var isRepeatable bool
	err = tx.QueryRowContext(r.Context(),
		`SELECT value, source_account_id, target_account_id, is_repeatable FROM transfers WHERE id=$1 AND user_id=$2`,
		id, userID,
	).Scan(&value, &srcID, &tgtID, &isRepeatable)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "transfer not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	tx.ExecContext(r.Context(), `DELETE FROM transfers WHERE id=$1`, id)

	if !isRepeatable {
		tx.ExecContext(r.Context(),
			`UPDATE bank_accounts SET balance=balance+$1, updated_at=NOW() WHERE id=$2`, value, srcID)
		tx.ExecContext(r.Context(),
			`UPDATE bank_accounts SET balance=balance-$1, updated_at=NOW() WHERE id=$2`, value, tgtID)
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
