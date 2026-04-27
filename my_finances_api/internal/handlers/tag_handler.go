package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/thiago/my_finances_api/internal/middleware"
	"github.com/thiago/my_finances_api/internal/models"
)

func isReservedTagName(name string) bool {
	n := strings.ToLower(strings.TrimSpace(name))
	return n == "income" || n == "expense"
}

func (h *Handler) listTags(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)

	rows, err := h.db.QueryContext(r.Context(),
		`SELECT id, user_id, name, color, is_system, created_at, updated_at
		 FROM tags
		 WHERE (user_id=$1 AND is_system=false) OR is_system=true
		 ORDER BY is_system DESC, name`, userID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	defer rows.Close()

	tags := []models.Tag{}
	for rows.Next() {
		var t models.Tag
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.Color, &t.IsSystem, &t.CreatedAt, &t.UpdatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		tags = append(tags, t)
	}

	writeJSON(w, http.StatusOK, tags)
}

func (h *Handler) createTag(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)

	var req models.CreateTagRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if isReservedTagName(req.Name) {
		writeError(w, http.StatusBadRequest, "tag name is reserved")
		return
	}

	var t models.Tag
	err := h.db.QueryRowContext(r.Context(),
		`INSERT INTO tags (user_id, name, color) VALUES ($1,$2,$3)
		 RETURNING id, user_id, name, color, is_system, created_at, updated_at`,
		userID, req.Name, req.Color,
	).Scan(&t.ID, &t.UserID, &t.Name, &t.Color, &t.IsSystem, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusCreated, t)
}

func (h *Handler) updateTag(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)
	id := r.PathValue("id")

	var req models.UpdateTagRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if isReservedTagName(req.Name) {
		writeError(w, http.StatusBadRequest, "tag name is reserved")
		return
	}

	var t models.Tag
	err := h.db.QueryRowContext(r.Context(),
		`UPDATE tags SET name=$1, color=$2, updated_at=$3
		 WHERE id=$4 AND user_id=$5
		 RETURNING id, user_id, name, color, is_system, created_at, updated_at`,
		req.Name, req.Color, time.Now(), id, userID,
	).Scan(&t.ID, &t.UserID, &t.Name, &t.Color, &t.IsSystem, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		writeError(w, http.StatusNotFound, "tag not found")
		return
	}

	writeJSON(w, http.StatusOK, t)
}

func (h *Handler) deleteTag(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)
	id := r.PathValue("id")

	var isSystem bool
	err := h.db.QueryRowContext(r.Context(),
		`SELECT is_system FROM tags WHERE id=$1`, id,
	).Scan(&isSystem)
	if err != nil {
		writeError(w, http.StatusNotFound, "tag not found")
		return
	}
	if isSystem {
		writeError(w, http.StatusForbidden, "cannot delete system tag")
		return
	}

	res, err := h.db.ExecContext(r.Context(),
		`DELETE FROM tags WHERE id=$1 AND user_id=$2`, id, userID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		writeError(w, http.StatusNotFound, "tag not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) fetchTransactionTags(r *http.Request, txnID string) []models.Tag {
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT t.id, t.user_id, t.name, t.color, t.is_system, t.created_at, t.updated_at
		 FROM tags t
		 JOIN transaction_tags tt ON tt.tag_id = t.id
		 WHERE tt.transaction_id = $1`, txnID,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var t models.Tag
		rows.Scan(&t.ID, &t.UserID, &t.Name, &t.Color, &t.IsSystem, &t.CreatedAt, &t.UpdatedAt)
		tags = append(tags, t)
	}
	return tags
}

func (h *Handler) fetchTransferTags(r *http.Request, transferID string) []models.Tag {
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT t.id, t.user_id, t.name, t.color, t.is_system, t.created_at, t.updated_at
		 FROM tags t
		 JOIN transfer_tags tt ON tt.tag_id = t.id
		 WHERE tt.transfer_id = $1`, transferID,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var t models.Tag
		rows.Scan(&t.ID, &t.UserID, &t.Name, &t.Color, &t.IsSystem, &t.CreatedAt, &t.UpdatedAt)
		tags = append(tags, t)
	}
	return tags
}
