package shared

import (
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func IsUnavailableError(err error) bool {
	st, ok := status.FromError(err)
	if !ok {
		return false
	}
	return st.Code() == codes.Unavailable || st.Code() == codes.DeadlineExceeded
}

func IsNotFoundError(err error) bool {
	st, ok := status.FromError(err)
	if !ok {
		return false
	}
	return st.Code() == codes.NotFound
}

func WriteGRPCError(w http.ResponseWriter, r *http.Request, err error) {
	st, ok := status.FromError(err)
	if !ok {
		WritePopup(w, r, http.StatusInternalServerError, "Error", "internal server error")
		return
	}

	switch st.Code() {
	case codes.InvalidArgument:
		WritePopup(w, r, http.StatusBadRequest, "Bad request", st.Message())
	case codes.NotFound:
		WritePopup(w, r, http.StatusNotFound, "Not found", st.Message())
	case codes.PermissionDenied:
		WritePopup(w, r, http.StatusForbidden, "Forbidden", st.Message())
	case codes.AlreadyExists:
		WritePopup(w, r, http.StatusConflict, "Conflict", st.Message())
	case codes.Unauthenticated:
		WritePopup(w, r, http.StatusUnauthorized, "Unauthorized", st.Message())
	case codes.Unavailable, codes.DeadlineExceeded:
		WritePopup(w, r, http.StatusServiceUnavailable, "Unavailable", st.Message())
	default:
		WritePopup(w, r, http.StatusInternalServerError, "Error", "internal server error")
	}
}
