package handlers

import (
	"net/http"

	"google.golang.org/protobuf/types/known/timestamppb"
	"refurbished-marketplace/services/web/internal/views"
)

func mapUserView(id, email string, createdAt, updatedAt *timestamppb.Timestamp) views.UserView {
	return views.UserView{
		ID:        id,
		Email:     email,
		CreatedAt: formatTimestamp(createdAt),
		UpdatedAt: formatTimestamp(updatedAt),
	}
}

func createUserFromForm(r *http.Request) (string, string, error) {
	if !parseForm(r) {
		return "", "", errInvalidRequestBody
	}
	return r.FormValue("email"), r.FormValue("password"), nil
}

func (h *Handler) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	email, password, err := createUserFromForm(r)
	if err != nil || email == "" || password == "" {
		writeBadRequest(w, r, "invalid request body")
		return
	}

	u, err := h.users.CreateUser(r.Context(), email, password)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeHTML(w, r, http.StatusCreated, views.UserPage(mapUserView(u.Id, u.Email, u.CreatedAt, u.UpdatedAt)))
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

	writeHTML(w, r, http.StatusOK, views.UserPage(mapUserView(u.Id, u.Email, u.CreatedAt, u.UpdatedAt)))
}
