package gapi

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GrpcLogger(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	startTime := time.Now()
	// call the hadnler func - forward the request to the handler to be processed
	result, err := handler(ctx, req)
	duration := time.Since(startTime)

	statusCode := codes.Unknown
	if status, ok := status.FromError(err); ok {
		statusCode = status.Code()
	}

	// default looger type is info
	logger := log.Info()
	// change logger type to err if error from respective gRPC handler
	if err != nil {
		logger = log.Error().Stack().Err(err)
	}

	// log here
	logger.
		Str("protocol", "grpc").
		Str("method", info.FullMethod).
		Dur("duration", duration).
		Int("status_code", int(statusCode)).
		Str("status_text ", statusCode.String()).
		Msg("received a gRPC request")

	return result, err
}
