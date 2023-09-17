package gapi

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GrpcLogger via interceptor
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

// ResponseRecorder - To capture the response details of our http requests
type ResponseRecorder struct {
	http.ResponseWriter
	StatusCode int
	Body       []byte
}

// override WriteHeader to save response status, and then call the original ResponseWriter.WriteHeader func
func (rec *ResponseRecorder) WriteHeader(statusCode int) {
	rec.StatusCode = statusCode
	rec.ResponseWriter.WriteHeader(statusCode)
}

// override Write to save response body details, and then call the original ResponseWriter.Write   func
func (rec *ResponseRecorder) Write(body []byte) (int, error) {
	rec.Body = body
	return rec.ResponseWriter.Write(body)
}

// HttpLogger middleware -
func HttpLogger(handler http.Handler) http.Handler {
	// doing a type conversion from a  nonymous func to http.HandlerFunc - which implemets the ServeHTTP func required by http.Handler interface
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		startTime := time.Now()
		// use ResponseRecorder instance when serving the request
		rec := &ResponseRecorder{
			ResponseWriter: res,
			StatusCode:     http.StatusOK, // default StatusCode is ok, will be set to correct value when over-ridden WriteHeader func is called
		}
		// call the hadnler func (with rec instead of res)  - forward the request to the handler to be processed
		handler.ServeHTTP(rec, req)
		duration := time.Since(startTime)

		// default looger type is info
		logger := log.Info()
		// change logger type to err if error from respective gRPC handler
		if rec.StatusCode != http.StatusOK {
			// var errObj map[string]interface{}
			// err := json.Unmarshal(rec.Body, &errObj)
			// log.Debug().Msgf("errObj: %+v", errObj)
			// if err != nil {
			// 	log.Fatal().Err(err).Msg("cannot unmarshal err response")
			// }
			logger = log.Error().RawJSON("error", rec.Body)
		}
		// log here
		logger.
			Str("protocol", "http").
			Str("method", req.Method).
			Str("path", req.RequestURI).
			Dur("duration", duration).
			Int("status_code", int(rec.StatusCode)).              // status code captured via ResponseRecorder
			Str("status_text ", http.StatusText(rec.StatusCode)). // status text via status code
			Msg("received an HTTP request")
	})
}
