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
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	mockdb "github.com/web3dev6/simplebank/db/mock"
	db "github.com/web3dev6/simplebank/db/sqlc"
	"github.com/web3dev6/simplebank/token"
	"github.com/web3dev6/simplebank/util"
)

func TestGetAccountApi(t *testing.T) {
	// test account to get
	user, _ := randomUser(t)
	account := randomAccount(user.Username)

	// initilalise ctrl and store
	// New in go1.14+, if you are passing a *testing.T into this function you no
	// longer need to call ctrl.Finish() in your test methods.
	ctrl := gomock.NewController(t)
	store := mockdb.NewMockStore(ctrl)
	// build stubs - select which methods in store will be called in this test
	store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(account, nil)

	// start test server
	// note* we dont have to start a real http server, use recorder from httptest package insted of server.listen
	server := newTestServer(t, store) // using newTestServer instead of NewServer
	recorder := httptest.NewRecorder()

	// create url and GET request
	// url path of the api to be called for ger account
	url := fmt.Sprintf("/accounts/%d", account.ID)
	// new http GET request
	request, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)

	// add correct auth-header to request with auth-token
	addAuthorization(t, request, server.tokenMaker, authorizationTypeBearer, user.Username, time.Minute)

	// send request
	server.router.ServeHTTP(recorder, request)

	// inspect status code - should be 200 with happy case
	require.Equal(t, http.StatusOK, recorder.Code)
	// inspect resp body - should have the expected account in body
	requireBodyMatchAccount(t, recorder.Body, account)
}

func TestGetAccountApiWithFullCoverage(t *testing.T) {
	// test account to get
	user, _ := randomUser(t)
	account := randomAccount(user.Username)

	// define test-cases here
	testcases := []struct {
		name          string
		accountID     int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name:      "NotFound",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(db.Account{}, db.ErrRecordNotFound) // db.ErrRecordNotFound is classified as 404 Error
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(db.Account{}, sql.ErrConnDone) // sql.ErrConnDone is classified as an InternalError
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "BadRequest",
			accountID: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:      "UnauthorizedUser",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// adding authorization token of unauthorized_user, account was created with user.Username
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "unauthorized_user", time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:      "NoAuthorization",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// no authorization token provided
			},
			buildStubs: func(store *mockdb.MockStore) {
				// GetAccount isn't called even as mw returns 404
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	// run all test-cases here
	for i := range testcases {
		// initialise test-case in a var
		tc := testcases[i]
		t.Run(tc.name, func(t *testing.T) {
			// initilalise ctrl and store
			ctrl := gomock.NewController(t)
			store := mockdb.NewMockStore(ctrl)

			// call buildStubs func of testcase here
			tc.buildStubs(store)

			// start test server
			// note* we dont have to start a real http server, use recorder from httptest package insted of server.listen
			server := newTestServer(t, store) // using newTestServer instead of NewServer
			recorder := httptest.NewRecorder()

			// create url and GET request
			// url path of the api to be called for ger account
			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			// new http GET request
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			// call setupAuth func of testcase here
			tc.setupAuth(t, request, server.tokenMaker)

			// send request
			server.router.ServeHTTP(recorder, request)

			// call checkResponse func of testcase here
			tc.checkResponse(t, recorder)
		})
	}
}

func randomAccount(owner string) db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    owner,
		Balance:  util.RandomBalance(),
		Currency: util.RandomCurrency(),
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, expectedAccount db.Account) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var accountFromBody db.Account
	err = json.Unmarshal(data, &accountFromBody)
	require.NoError(t, err)

	require.Equal(t, expectedAccount, accountFromBody)
}
