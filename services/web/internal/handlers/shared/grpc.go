package shared

import (
	"net/http"

	sharedviews "refurbished-marketplace/services/web/internal/views/shared"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func WriteGRPCError(w http.ResponseWriter, r *http.Request, err error) {
	st, ok := status.FromError(err)
	if !ok {
		WriteHTML(w, r, http.StatusInternalServerError, sharedviews.MessagePage("Error", "internal server error"))
		return
	}

	switch st.Code() {
	case codes.InvalidArgument:
		WriteHTML(w, r, http.StatusBadRequest, sharedviews.MessagePage("Bad request", st.Message()))
	case codes.NotFound:
		WriteHTML(w, r, http.StatusNotFound, sharedviews.MessagePage("Not found", st.Message()))
	case codes.PermissionDenied:
		WriteHTML(w, r, http.StatusForbidden, sharedviews.MessagePage("Forbidden", st.Message()))
	case codes.AlreadyExists:
		WriteHTML(w, r, http.StatusConflict, sharedviews.MessagePage("Conflict", st.Message()))
	case codes.Unauthenticated:
		WriteHTML(w, r, http.StatusUnauthorized, sharedviews.MessagePage("Unauthorized", st.Message()))
	default:
		WriteHTML(w, r, http.StatusInternalServerError, sharedviews.MessagePage("Error", "internal server error"))
	}
}
