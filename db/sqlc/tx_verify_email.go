package db

import (
	"context"
	"database/sql"
)

// VerifyEmailTxParams contains the input parameters of the VerifyEmail transaction
type VerifyEmailTxParams struct {
	EmailId    int64
	SecretCode string
}

// VerifyEmailTxResult contains the result of the VerifyEmail transaction
// Note* User must be anonymously embedded to make json marshall/unmarshal work in user_test
type VerifyEmailTxResult struct {
	User        User
	VerifyEmail VerifyEmail
}

// VerifyEmailTx updates entrirs in verify_emails and users tables
// query to find verify_email record with give email_id & secret_code & update is_used to true
//
//	and to update the is_email_verified of the corresponding user to true
func (store *SQLStore) VerifyEmailTx(ctx context.Context, arg VerifyEmailTxParams) (VerifyEmailTxResult, error) {
	var result VerifyEmailTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		// call store:update_verify_email and update verify_email in db - set is_used to true for record with given id & secret_code
		result.VerifyEmail, err = q.UpdateVerifyEmail(ctx, UpdateVerifyEmailParams{
			ID:         arg.EmailId,
			SecretCode: arg.SecretCode,
		})
		if err != nil {
			return err
		}
		// call store:update_user and update user in db - set is_email_verified to true for record with given username
		result.User, err = q.UpdateUser(ctx, UpdateUserParams{
			Username: result.VerifyEmail.Username,
			IsEmailVerified: sql.NullBool{
				Bool:  true,
				Valid: true,
			},
		})

		return err
	})

	return result, err
}
