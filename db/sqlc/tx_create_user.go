package db

import "context"

// CreateUserTxParams contains the input parameters of the CreateUser transaction
type CreateUserTxParams struct {
	CreateUserParams                       // embedded CreateUserParams - to be used to call store:create_user
	AfterCreate      func(user User) error // special AfterCreate - callback fn to be   executed after user is inserted in same db tx
}

// CreateUserTxResult contains the result of the CreateUser transaction
type CreateUserTxResult struct {
	User User
}

// CreateUserTx creates a new user and sends a welcome email to user's email
func (store *SQLStore) CreateUserTx(ctx context.Context, arg CreateUserTxParams) (CreateUserTxResult, error) {
	var result CreateUserTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		// call store:create_user and create user in db
		result.User, err = q.CreateUser(ctx, arg.CreateUserParams)
		if err != nil {
			return err
		}

		// execute the callback fn AfterCreate now by passing the created user
		err = arg.AfterCreate(result.User)
		return err
	})

	return result, err
}
