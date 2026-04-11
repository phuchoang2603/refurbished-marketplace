package handlers

import (
	"net/http"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type createUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func mapUser(id, email string, createdAt, updatedAt *timestamppb.Timestamp) userResponse {
	return userResponse{
		ID:        id,
		Email:     email,
		CreatedAt: formatTimestamp(createdAt),
		UpdatedAt: formatTimestamp(updatedAt),
	}
}

func (h *Handler) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	u, err := h.users.CreateUser(r.Context(), req.Email, req.Password)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, mapUser(u.Id, u.Email, u.CreatedAt, u.UpdatedAt))
}

func (h *Handler) handleGetUserByID(w http.ResponseWriter, r *http.Request) {
	id, ok := requirePathValue(w, r, "id", "invalid user id")
	if !ok {
		return
	}

	u, err := h.users.GetUserByID(r.Context(), id)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, mapUser(u.Id, u.Email, u.CreatedAt, u.UpdatedAt))
}
