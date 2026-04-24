package handlers

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/thiago/my_finances_api/internal/middleware"
	"github.com/thiago/my_finances_api/internal/models"
)

func (h *Handler) listGoals(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)

	rows, err := h.db.QueryContext(r.Context(),
		`SELECT id, user_id, source_account_id, target_account_id, name,
		        start_date, end_date, interval_days, target_value, created_at, updated_at
		 FROM goals WHERE user_id=$1 ORDER BY start_date`, userID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	defer rows.Close()

	goals := []models.Goal{}
	for rows.Next() {
		var g models.Goal
		if err := rows.Scan(&g.ID, &g.UserID, &g.SourceAccountID, &g.TargetAccountID,
			&g.Name, &g.StartDate, &g.EndDate, &g.IntervalDays, &g.TargetValue,
			&g.CreatedAt, &g.UpdatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		goals = append(goals, g)
	}

	writeJSON(w, http.StatusOK, goals)
}

func (h *Handler) createGoal(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)

	var req models.CreateGoalRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || req.SourceAccountID == "" || req.TargetAccountID == "" {
		writeError(w, http.StatusBadRequest, "name, source_account_id and target_account_id are required")
		return
	}
	if req.IntervalDays <= 0 {
		writeError(w, http.StatusBadRequest, "interval_days must be positive")
		return
	}
	if req.TargetValue <= 0 {
		writeError(w, http.StatusBadRequest, "target_value must be positive")
		return
	}
	if req.StartDate.IsZero() {
		req.StartDate = time.Now()
	}
	if req.EndDate.Before(req.StartDate) {
		writeError(w, http.StatusBadRequest, "end_date must be after start_date")
		return
	}

	var g models.Goal
	err := h.db.QueryRowContext(r.Context(),
		`INSERT INTO goals (user_id, source_account_id, target_account_id, name,
		                    start_date, end_date, interval_days, target_value)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		 RETURNING id, user_id, source_account_id, target_account_id, name,
		           start_date, end_date, interval_days, target_value, created_at, updated_at`,
		userID, req.SourceAccountID, req.TargetAccountID, req.Name,
		req.StartDate, req.EndDate, req.IntervalDays, req.TargetValue,
	).Scan(&g.ID, &g.UserID, &g.SourceAccountID, &g.TargetAccountID,
		&g.Name, &g.StartDate, &g.EndDate, &g.IntervalDays, &g.TargetValue,
		&g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusCreated, g)
}

func (h *Handler) updateGoal(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)
	id := r.PathValue("id")

	var req models.UpdateGoalRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	var g models.Goal
	err := h.db.QueryRowContext(r.Context(),
		`UPDATE goals SET name=$1, end_date=$2, interval_days=$3, target_value=$4, updated_at=NOW()
		 WHERE id=$5 AND user_id=$6
		 RETURNING id, user_id, source_account_id, target_account_id, name,
		           start_date, end_date, interval_days, target_value, created_at, updated_at`,
		req.Name, req.EndDate, req.IntervalDays, req.TargetValue, id, userID,
	).Scan(&g.ID, &g.UserID, &g.SourceAccountID, &g.TargetAccountID,
		&g.Name, &g.StartDate, &g.EndDate, &g.IntervalDays, &g.TargetValue,
		&g.CreatedAt, &g.UpdatedAt)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "goal not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, g)
}

func (h *Handler) deleteGoal(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)
	id := r.PathValue("id")

	res, err := h.db.ExecContext(r.Context(),
		`DELETE FROM goals WHERE id=$1 AND user_id=$2`, id, userID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		writeError(w, http.StatusNotFound, "goal not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
