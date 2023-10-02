package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
	mockdb "github.com/web3dev6/simplebank/db/mock"
	db "github.com/web3dev6/simplebank/db/sqlc"
	"github.com/web3dev6/simplebank/util"
	"github.com/web3dev6/simplebank/worker"
)

// type eqCreateUserParamsMatcher struct {
// 	arg      db.CreateUserParams
// 	password string
// }

// // passes iff -> password in request body of test-case gets hashed -> and CreateUser is called with corresponding db.CreateUserParams arg
// func (e eqCreateUserParamsMatcher) Matches(x interface{}) bool {
// 	// convert interface to expected type first
// 	arg, ok := x.(db.CreateUserParams)
// 	if !ok {
// 		return false
// 	}

// 	// e.password & arg.HashedPassword are coming from CreateUser exec context
// 	// (after password is hashed, and CreateUser is called with db.CreateUserParams, which includes hashedPassword)
// 	err := util.CheckPassword(e.password, arg.HashedPassword)
// 	if err != nil {
// 		return false
// 	}
// 	e.arg.HashedPassword = arg.HashedPassword

// 	// extra check to ensure- arg in our matcher struct is now same as the arg that was passed in invocation context
// 	return reflect.DeepEqual(e.arg, x)
// }

// func (e eqCreateUserParamsMatcher) String() string {
// 	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
// }

// func EqCreateUserParams(arg db.CreateUserParams, password string) gomock.Matcher {
// 	return eqCreateUserParamsMatcher{arg, password}
// }

type eqCreateUserTxParamsMatcher struct {
	arg      db.CreateUserTxParams
	password string
}

// passes iff -> password in request body of test-case gets hashed -> and CreateUserTx is called with corresponding db.CreateUserTxParams arg
func (e eqCreateUserTxParamsMatcher) Matches(x interface{}) bool {
	// convert interface to expected type first
	arg, ok := x.(db.CreateUserTxParams)
	if !ok {
		return false
	}

	// e.password & arg.HashedPassword are coming from CreateUserTx exec context
	// (after password is hashed, and CreateUserTx is called with db.CreateUserTxParams, which includes CreateUserParams:hashedPassword)
	err := util.CheckPassword(e.password, arg.CreateUserParams.HashedPassword)
	if err != nil {
		return false
	}
	e.arg.HashedPassword = arg.HashedPassword

	// extra check to ensure- arg in our matcher struct is now same as the arg that was passed in invocation context
	// Note* e.arg won't be the same as x, since CreateUserTxParams also contains AfterCreate callback fn instance which arg is missing, so add that
	e.arg.AfterCreate = x.(db.CreateUserTxParams).AfterCreate
	// return reflect.DeepEqual(e.arg, x)
	return true
}

func (e eqCreateUserTxParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}

func EqCreateUserTxParams(arg db.CreateUserTxParams, password string) gomock.Matcher {
	return eqCreateUserTxParamsMatcher{arg, password}
}

func TestCreateUserAPI(t *testing.T) {
	user, password := randomUser(t)

	// note* passing hashedPassword in CreateUser arg fails the test as in test exec, a new hashPassword is created and match fails
	// hashedPassword, err := util.HashPassword(password)
	// require.NoError(t, err)

	// checkResponse can have the same testing context {t *testing.T of TestCreateUserAPI} across all test-cases
	// note* In account_test, checkResponse signature had separate {t *testing.T}
	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateUserTxParams{
					CreateUserParams: db.CreateUserParams{
						Username: user.Username,
						FullName: user.FullName,
						Email:    user.Email,
					},
				}
				store.EXPECT().
					CreateUserTx(gomock.Any(), EqCreateUserTxParams(arg, password)).
					Times(1).
					Return(db.CreateUserTxResult{User: user}, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchCreateUserTxResult(t, recorder.Body, user)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.CreateUserTxResult{User: db.User{}}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "DuplicateUsername",
			body: gin.H{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.CreateUserTxResult{User: db.User{}}, db.ErrUniqueViolation)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "InvalidUsername",
			body: gin.H{
				"username":  "invalid-user#1",
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidEmail",
			body: gin.H{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     "invalid-email",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "TooShortPassword",
			body: gin.H{
				"username":  user.Username,
				"password":  "123",
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			store := mockdb.NewMockStore(ctrl)

			tc.buildStubs(store)

			// Redis task distributor
			taskDistributor := worker.NewRedisTaskDistributor(asynq.RedisClientOpt{
				Addr: "0.0.0.0:6379",
			})
			server := newTestServer(t, store, taskDistributor) // using newTestServer instead of NewServer
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON - POST request
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/users"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func randomUser(t *testing.T) (user db.User, password string) {
	password = util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	user = db.User{
		Username:       util.RandomOwner(),
		HashedPassword: hashedPassword,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}
	return
}

// func requireBodyMatchUser(t *testing.T, body *bytes.Buffer, user db.User) {
// 	data, err := io.ReadAll(body)
// 	require.NoError(t, err)

// 	var userFromBody db.User
// 	err = json.Unmarshal(data, &userFromBody)

// 	require.NoError(t, err)
// 	require.Equal(t, user.Username, userFromBody.Username)
// 	require.Equal(t, user.FullName, userFromBody.FullName)
// 	require.Equal(t, user.Email, userFromBody.Email)
// 	require.Empty(t, userFromBody.HashedPassword)
// }

func requireBodyMatchCreateUserTxResult(t *testing.T, body *bytes.Buffer, user db.User) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var txResultFromBody db.CreateUserTxResult
	err = json.Unmarshal(data, &txResultFromBody)

	require.NoError(t, err)
	require.NotNil(t, txResultFromBody)
	require.Equal(t, user.Username, txResultFromBody.User.Username)
	require.Equal(t, user.FullName, txResultFromBody.User.FullName)
	require.Equal(t, user.Email, txResultFromBody.User.Email)
	require.Empty(t, txResultFromBody.User.HashedPassword)
}
