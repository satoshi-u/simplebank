package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	db "github.com/web3dev6/simplebank/db/sqlc"
	"github.com/web3dev6/simplebank/token"
	"github.com/web3dev6/simplebank/util"
	"github.com/web3dev6/simplebank/worker"
)

type createUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"` // must only be alpha-numeric with validator's inbuilt alphanum tag
	Password string `json:"password" binding:"required,min=6"`    // must be atleast 6 chars
	FullName string `json:"full_name" binding:"required"`         // required
	Email    string `json:"email" binding:"required,email"`       // must be email with validator's inbuilt alphanum tag
}

type userResponse struct {
	Username          string    `json:"username"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

func newUserResponse(user db.User) userResponse {
	return userResponse{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}
}

func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		abortWithErrorResponse(ctx, http.StatusBadRequest, err)
		return
	}
	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		abortWithErrorResponse(ctx, http.StatusInternalServerError, err)
	}

	// using SQL DB Tx with CreateUserTx - as user shouldn't be created if taskDistributor fails - rollback
	// make create_user_tx params, instead of create_user params directly
	arg := db.CreateUserTxParams{
		CreateUserParams: db.CreateUserParams{
			Username:       req.Username,
			HashedPassword: hashedPassword,
			FullName:       req.FullName,
			Email:          req.Email,
		},
		AfterCreate: func(user db.User) error {
			// send verification email to user - put task in redis queue via distributor
			taskPayload := &worker.PayloadSendVerifyEmail{
				Username: user.Username,
			}
			// asynq options to configure task processing while putting it in queue
			opts := []asynq.Option{
				asynq.MaxRetry(10),                // retry fails 10 times
				asynq.ProcessIn(3 * time.Second),  // process in 3 secs
				asynq.Queue(worker.QueueCritical), // push in queue "critical"
			}
			err = server.taskDistributor.DistributeTaskSendVerifyEmail(ctx, taskPayload, opts...)
			if err != nil {
				return fmt.Errorf("failed to distribute task TaskSendVerifyEmail: %w", err)
			}
			return nil
		},
	}

	// now call CreateUserTx, instead of CreateUser directly
	txResult, err := server.store.CreateUserTx(ctx, arg)
	if err != nil {
		// username and email must be unique (UNIQUE)
		if db.ErrorCode(err) == db.UniqueViolation {
			abortWithErrorResponse(ctx, http.StatusForbidden, err)
			return
		}
		abortWithErrorResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	resp := newUserResponse(txResult.User)
	ctx.JSON(http.StatusOK, resp)
}

func (server *Server) getUserDetails(ctx *gin.Context) {
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	user, err := server.store.GetUser(ctx, authPayload.Username)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) || errors.Is(err, sql.ErrNoRows) {
			abortWithErrorResponse(ctx, http.StatusNotFound, err)
			return
		}
		abortWithErrorResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	resp := newUserResponse(user)
	ctx.JSON(http.StatusOK, resp)
}

type loginUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"` // must only be alpha-numeric with validator's inbuilt alphanum tag
	Password string `json:"password" binding:"required,min=6"`    // must be atleast 6 chars
}

type loginUserResponse struct {
	SessionID             uuid.UUID    `json:"session_id"`
	AccessToken           string       `json:"access_token"`
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
	RefreshToken          string       `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
	User                  userResponse `json:"user"`
}

func (server *Server) loginUser(ctx *gin.Context) {
	var req loginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		abortWithErrorResponse(ctx, http.StatusBadRequest, err)
		return
	}

	// get user from db
	user, err := server.store.GetUser(ctx, req.Username)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) || errors.Is(err, sql.ErrNoRows) {
			abortWithErrorResponse(ctx, http.StatusNotFound, err)
			return
		}
		abortWithErrorResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	// check password and create tokens if all ok, or error out
	err = util.CheckPassword(req.Password, user.HashedPassword)
	if err != nil {
		abortWithErrorResponse(ctx, http.StatusUnauthorized, err)
		return
	}
	accessToken, accessPayload, err := server.tokenMaker.CreateToken(user.Username, server.config.AccessTokenDuration)
	if err != nil {
		abortWithErrorResponse(ctx, http.StatusInternalServerError, err)
		return
	}
	refreshToken, refreshPayload, err := server.tokenMaker.CreateToken(user.Username, server.config.RefreshTokenDuration)
	if err != nil {
		abortWithErrorResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	// create a session in sessions table for user
	session, err := server.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    ctx.Request.UserAgent(),
		ClientIp:     ctx.ClientIP(),
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiresAt,
	})
	if err != nil {
		abortWithErrorResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	// send ok response if all ok WITH loginUserResponse
	resp := loginUserResponse{
		SessionID:             session.ID,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiresAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiresAt,
		User:                  newUserResponse(user),
	}
	ctx.JSON(http.StatusOK, resp)
}

type updateUserRequest struct {
	Username string  `json:"username" binding:"required,alphanum"`         // required - update user based on this key
	Password *string `json:"password,omitempty" binding:"omitempty,min=6"` // optional - todo add regex
	FullName *string `json:"full_name,omitempty" binding:"omitempty"`      // optional - todo add regex
	Email    *string `json:"email,omitempty" binding:"omitempty,email"`    // optional
}

func (server *Server) updateUser(ctx *gin.Context) {
	// get updateUserRequest
	var req updateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		abortWithErrorResponse(ctx, http.StatusBadRequest, err)
		return
	}

	// check if authorized user from access_token
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if req.Username != authPayload.Username {
		abortWithErrorResponse(ctx, http.StatusUnauthorized, ErrUpdatingUserInfoFromUnauthorizedUser)
		return
	}

	// fmt.Printf("updateUserRequest: %+v", req)
	// make update_user params with username
	arg := db.UpdateUserParams{
		Username: req.Username,
	}
	if req.FullName != nil {
		// set hash_password
		arg.FullName = sql.NullString{
			String: *req.FullName,
			Valid:  true,
		}
	}
	if req.Email != nil {
		// set email
		arg.Email = sql.NullString{
			String: *req.Email,
			Valid:  true,
		}
	}
	if req.Password != nil {
		// hash password
		hashedPassword, err := util.HashPassword(*req.Password)
		if err != nil {
			abortWithErrorResponse(ctx, http.StatusInternalServerError, err)
		}
		// set hash_password
		arg.HashedPassword = sql.NullString{
			String: hashedPassword,
			Valid:  true,
		}
		arg.PasswordChangedAt = sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		}
	}

	// update user in db
	user, err := server.store.UpdateUser(ctx, arg)
	if err != nil {
		if db.ErrorCode(err) == db.ErrRecordNotFound.Error() {
			abortWithErrorResponse(ctx, http.StatusNotFound, err)
			return
		}
		abortWithErrorResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	// return resp
	resp := newUserResponse(user)
	ctx.JSON(http.StatusOK, resp)
}

func (server *Server) verifyUserEmail(ctx *gin.Context) {
	// get input from query params
	emailId := ctx.Query("email_id") // shortcut for c.Request.URL.Query().Get("email_id")
	eId, err := strconv.Atoi(emailId)
	if err != nil {
		// log.Err(err).Msg("error in strconv.Atoi")
		abortWithErrorResponse(ctx, http.StatusInternalServerError, err)
		return
	}
	secretCode := ctx.Query("secret_code")
	// log.Debug().Msgf("eId %d", int64(eId))
	// log.Debug().Msgf("secretCode %s", secretCode)

	// call VerifyEmailTx
	txResult, err := server.store.VerifyEmailTx(ctx, db.VerifyEmailTxParams{
		EmailId:    int64(eId),
		SecretCode: secretCode,
	})
	if err != nil {
		// log.Err(err).Msg("error in server.store.VerifyEmailTx")
		abortWithErrorResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	// return resp
	resp := struct {
		IsVerified bool `json:"is_verified"`
	}{
		IsVerified: txResult.User.IsEmailVerified,
	}
	// log.Debug().Msgf("resp %+v", resp)
	ctx.JSON(http.StatusOK, resp)
}
