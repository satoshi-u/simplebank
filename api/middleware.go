package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
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
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(ErrMissingAuthHeader))
			return
		}

		// check auth header format
		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(ErrInvalidAuthHeaderFormat))
			return
		}

		// check auth header type if bearer or not
		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(ErrUnsupportedAuthType))
			return
		}

		// check bearer token
		accessToken := fields[1]
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			// return http.StatusUnauthorized code with encountered error
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
		}

		ctx.Set(authorizationPayloadKey, payload) // set payload in ctx values bag
		ctx.Next()                                // pass in the chain
	}
}
