package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"refurbished-marketplace/services/web/internal/usersclient"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Handler struct {
	users *usersclient.Client
}

func New(users *usersclient.Client) *Handler {
	return &Handler{users: users}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", h.handleHealthz)
	mux.HandleFunc("POST /users", h.handleCreateUser)
	mux.HandleFunc("GET /users/{id}", h.handleGetUserByID)
	mux.HandleFunc("POST /auth/login", h.handleLogin)
	mux.HandleFunc("POST /auth/refresh", h.handleRefresh)
	mux.HandleFunc("POST /auth/logout", h.handleLogout)
}

type createUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type userResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

func (h *Handler) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
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

	writeJSON(w, http.StatusOK, mapUser(u.Id, u.Email, u.CreatedAt, u.UpdatedAt))
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	tokens, err := h.users.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, mapTokens(tokens.AccessToken, tokens.RefreshToken, tokens.TokenType, tokens.ExpiresIn))
}

func (h *Handler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	tokens, err := h.users.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, mapTokens(tokens.AccessToken, tokens.RefreshToken, tokens.TokenType, tokens.ExpiresIn))
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	var req logoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	_, err := h.users.Logout(r.Context(), req.RefreshToken)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func mapUser(id, email string, createdAt, updatedAt *timestamppb.Timestamp) userResponse {
	var created string
	var updated string
	if createdAt != nil {
		created = createdAt.AsTime().UTC().Format("2006-01-02T15:04:05Z07:00")
	}
	if updatedAt != nil {
		updated = updatedAt.AsTime().UTC().Format("2006-01-02T15:04:05Z07:00")
	}
	return userResponse{ID: id, Email: email, CreatedAt: created, UpdatedAt: updated}
}

func mapTokens(access, refresh, tokenType string, expiresIn int64) tokenResponse {
	return tokenResponse{AccessToken: access, RefreshToken: refresh, TokenType: tokenType, ExpiresIn: expiresIn}
}

func writeGRPCError(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	switch st.Code() {
	case codes.InvalidArgument:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": st.Message()})
	case codes.NotFound:
		writeJSON(w, http.StatusNotFound, map[string]string{"error": st.Message()})
	case codes.AlreadyExists:
		writeJSON(w, http.StatusConflict, map[string]string{"error": st.Message()})
	case codes.Unauthenticated:
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": st.Message()})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
