package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	webAuth "refurbished-marketplace/services/web/internal/auth"
)

const timestampFormat = "2006-01-02T15:04:05Z07:00"

func formatTimestamp(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	return ts.AsTime().UTC().Format(timestampFormat)
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

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return false
	}
	return true
}

func requireUserID(w http.ResponseWriter, r *http.Request) (string, bool) {
	userID, ok := webAuth.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return "", false
	}
	return userID, true
}

func requirePathValue(w http.ResponseWriter, r *http.Request, key, errorMessage string) (string, bool) {
	value := strings.TrimSpace(r.PathValue(key))
	if value == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": errorMessage})
		return "", false
	}
	return value, true
}

func queryInt32Param(w http.ResponseWriter, r *http.Request, key string, defaultValue int32, minValue int32, errorMessage string) (int32, bool) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return defaultValue, true
	}

	v, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || int32(v) < minValue {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": errorMessage})
		return 0, false
	}

	return int32(v), true
}
