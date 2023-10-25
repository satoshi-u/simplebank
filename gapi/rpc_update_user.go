package gapi

import (
	"context"
	"database/sql"
	"time"

	db "github.com/web3dev6/simplebank/db/sqlc"
	"github.com/web3dev6/simplebank/pb"
	"github.com/web3dev6/simplebank/util"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	// check if valid token & authorized user
	// Note: Using authorizeUser here makes things work for both grpc and http-gateway server
	authPayload, err := server.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}
	if req.Username != authPayload.Username {
		return nil, status.Errorf(codes.PermissionDenied, "mismatch in username from authToken and update_request payload: %s", err)
	}

	// validate update_request & err handling
	violations := validateUpdateUserRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	// make update_user params
	arg := db.UpdateUserParams{
		Username: req.GetUsername(),
		FullName: sql.NullString{
			String: req.GetFullName(),
			Valid:  req.FullName != nil,
		},
		Email: sql.NullString{
			String: req.GetEmail(),
			Valid:  req.Email != nil,
		},
	}

	if req.Password != nil {
		// hash password
		hashedPassword, err := util.HashPassword(req.GetPassword())
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to hash password: %s", err)
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

	// call UpdateUser for db
	user, err := server.store.UpdateUser(ctx, arg)
	if err != nil {
		if db.ErrorCode(err) == db.ErrRecordNotFound.Error() || err.Error() == sql.ErrNoRows.Error() {
			return nil, status.Errorf(codes.NotFound, "user with username not found: %s", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to update user: %s", err)
	}

	// return resp
	resp := &pb.UpdateUserResponse{
		User: convertUser(user),
	}
	return resp, nil
}

func validateUpdateUserRequest(req *pb.UpdateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	// required
	if err := ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}
	// optional Password
	if req.Password != nil {
		if err := ValidatePassword(req.GetPassword()); err != nil {
			violations = append(violations, fieldViolation("password", err))
		}
	}
	// optional Email
	if req.Email != nil {
		if err := ValidateEmail(req.GetEmail()); err != nil {
			violations = append(violations, fieldViolation("email", err))
		}
	}
	// optional FullName
	if req.FullName != nil {
		if err := ValidateFullname(req.GetFullName()); err != nil {
			violations = append(violations, fieldViolation("full_name", err))
		}
	}
	return violations
}
