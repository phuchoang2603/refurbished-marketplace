package grpcerr

import (
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ParseUUID(raw, field string) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid %s", field))
	}
	return id, nil
}
