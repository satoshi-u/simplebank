package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/web3dev6/simplebank/token"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

func authMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// check auth header
		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			abortWithErrorResponse(ctx, http.StatusUnauthorized, ErrMissingAuthHeader)
			return
		}

		// check auth header format
		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			abortWithErrorResponse(ctx, http.StatusUnauthorized, ErrInvalidAuthHeaderFormat)
			return
		}

		// check auth header type if bearer or not
		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			abortWithErrorResponse(ctx, http.StatusUnauthorized, ErrUnsupportedAuthType)
			return
		}

		// check bearer token
		accessToken := fields[1]
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			// return http.StatusUnauthorized code with encountered error
			abortWithErrorResponse(ctx, http.StatusUnauthorized, err)
		}

		ctx.Set(authorizationPayloadKey, payload) // set payload in ctx values bag
		ctx.Next()                                // pass in the chain
	}
}

// loggerMiddleware logs a gin HTTP request in JSON format
func loggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now() // Start timer
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Initialize params
		param := gin.LogFormatterParams{}

		// Fill the params
		param.TimeStamp = time.Now() // Stop timer
		param.Latency = param.TimeStamp.Sub(start)
		if param.Latency > time.Minute {
			param.Latency = param.Latency.Truncate(time.Second)
		}
		param.ClientIP = c.ClientIP()
		param.Method = c.Request.Method
		param.StatusCode = c.Writer.Status()
		param.ErrorMessage = c.Errors.ByType(gin.ErrorTypeAny).String()
		param.BodySize = c.Writer.Size()
		if raw != "" {
			path = path + "?" + raw
		}
		param.Path = path

		// default looger type is info
		logger := log.Info()
		// change logger type to err if received error
		if c.Writer.Status() != http.StatusOK {
			// Note* Adding error as a key with value as param.ErrorMessage
			logger = log.Error().Str("error", param.ErrorMessage)
		}

		// log here
		// Note* Msg is sent as empty if no param.ErrorMessage
		logger.Str("client_id", param.ClientIP).
			Str("method", param.Method).
			Int("status_code", param.StatusCode).
			Str("status_text ", http.StatusText(param.StatusCode)).
			Int("body_size", param.BodySize).
			Str("path", param.Path).
			Str("latency", param.Latency.String()).
			Msg("http request -> Gin http server")
	}
}
