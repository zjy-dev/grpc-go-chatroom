package utils

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// WrapGRPCError wraps a given error with a gRPC status error if the error is not
// already a gRPC status error. It returns the original error if it is already
// a gRPC status error.
//
// If the provided error is nil, wrapGRPCError returns nil. If the error is not
// a gRPC status error, it creates a new status error with the provided gRPC
// status code and a message that includes the original error message.
func WrapGRPCError(err error, code codes.Code, msg string) error {
	if err == nil {
		return nil
	}

	_, ok := status.FromError(err)
	if ok {
		return err
	}
	return status.Errorf(code, msg+": %v", err)
}
