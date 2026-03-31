package handlers

import (
	"encoding/json"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
	case codes.PermissionDenied:
		writeJSON(w, http.StatusForbidden, map[string]string{"error": st.Message()})
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
