package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type createUserRequest struct {
	Email    string  `json:"email"`
	Password string  `json:"password"`
	XPos     float64 `json:"x_pos"`
	YPos     float64 `json:"y_pos"`
}

type userResponse struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	XPos      float64 `json:"x_pos"`
	YPos      float64 `json:"y_pos"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

func mapUser(id, email string, xPos, yPos float64, createdAt, updatedAt *timestamppb.Timestamp) userResponse {
	var created string
	var updated string
	if createdAt != nil {
		created = createdAt.AsTime().UTC().Format("2006-01-02T15:04:05Z07:00")
	}
	if updatedAt != nil {
		updated = updatedAt.AsTime().UTC().Format("2006-01-02T15:04:05Z07:00")
	}
	return userResponse{ID: id, Email: email, XPos: xPos, YPos: yPos, CreatedAt: created, UpdatedAt: updated}
}

func (h *Handler) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	u, err := h.users.CreateUser(r.Context(), req.Email, req.Password, req.XPos, req.YPos)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, mapUser(u.Id, u.Email, u.XPos, u.YPos, u.CreatedAt, u.UpdatedAt))
}

func (h *Handler) handleGetUserByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}

	u, err := h.users.GetUserByID(r.Context(), id)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, mapUser(u.Id, u.Email, u.XPos, u.YPos, u.CreatedAt, u.UpdatedAt))
}
