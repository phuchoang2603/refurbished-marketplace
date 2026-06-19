package grpcerr

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Mapping struct {
	Err     error
	Code    codes.Code
	Message string
}

func Map(err error, mappings ...Mapping) error {
	if err == nil {
		return nil
	}
	for _, mapping := range mappings {
		if errors.Is(err, mapping.Err) {
			msg := mapping.Message
			if msg == "" {
				msg = mapping.Err.Error()
			}
			return status.Error(mapping.Code, msg)
		}
	}
	return status.Error(codes.Internal, "internal error")
}

func InvalidArgument(message string) error {
	return status.Error(codes.InvalidArgument, message)
}

func NotFound(message string) error {
	return status.Error(codes.NotFound, message)
}

func AlreadyExists(message string) error {
	return status.Error(codes.AlreadyExists, message)
}

func Unauthenticated(message string) error {
	return status.Error(codes.Unauthenticated, message)
}

func Internal() error {
	return status.Error(codes.Internal, "internal error")
}
