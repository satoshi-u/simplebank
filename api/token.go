package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type renewAccessTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type renewAccessTokenResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

func (server *Server) renewAccessToken(ctx *gin.Context) {
	var req renewAccessTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		abortWithErrorResponse(ctx, http.StatusBadRequest, err)
		return
	}

	// check if refreshToken is valid or not
	refreshPayload, err := server.tokenMaker.VerifyToken(req.RefreshToken)
	if err != nil {
		// token is invalid or expired
		abortWithErrorResponse(ctx, http.StatusUnauthorized, err)
		return
	}

	// if refreshPayload is good, get corresponding session
	session, err := server.store.GetSession(ctx, refreshPayload.ID)
	if err != nil {
		// session not found from sessionId (which is refreshToken's uuid)
		abortWithErrorResponse(ctx, http.StatusNotFound, ErrSessionNotFound)
		return
	}

	// if session found, check if this session is not blocked
	if session.IsBlocked {
		abortWithErrorResponse(ctx, http.StatusUnauthorized, ErrBlockedSession)
		return
	}
	// also check if RefreshToken's Username is same as corresponding session's Username (from db)
	if session.Username != refreshPayload.Username {
		abortWithErrorResponse(ctx, http.StatusUnauthorized, ErrIncorrectSessionUser)
		return
	}
	// also check if RefreshToken is same as corresponding session's RefreshToken (from db)
	if session.RefreshToken != req.RefreshToken {
		abortWithErrorResponse(ctx, http.StatusUnauthorized, ErrIncorrectSessionToken)
		return
	}
	// note* token expiration is already checked for in VerifyToken (*Payload.Valid()), still check for rare case
	if time.Now().After(refreshPayload.ExpiresAt) {
		abortWithErrorResponse(ctx, http.StatusUnauthorized, ErrExpiredSession)
		return
	}

	// issue a new accessToken
	newAccessToken, accessPayload, err := server.tokenMaker.CreateToken(refreshPayload.Username, server.config.AccessTokenDuration)
	if err != nil {
		abortWithErrorResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	// send ok response if all ok WITH renewAccessToken Response
	resp := renewAccessTokenResponse{
		AccessToken:          newAccessToken,
		AccessTokenExpiresAt: accessPayload.ExpiresAt,
	}
	ctx.JSON(http.StatusOK, resp)
}
