package mannco

import (
	"context"
	"testing"
)

func TestUserLogin(t *testing.T) {
	runAPITest(t, testCase[string]{
		name:           "UserLogin_success",
		mockStatus:     200,
		mockResponse:   `{"err":false,"success":true,"message":"","content":{"jwt":"fake_jwt_token"}}`,
		expectedPath:   "/user/login",
		expectedMethod: "POST",
		runTest: func(ctx context.Context, client *Client) (string, error) {
			return client.UserLogin(ctx, "valid_api_key_123")
		},
		assertResponse: func(t *testing.T, jwt string) {
			if jwt != "fake_jwt_token" {
				t.Errorf("expected JWT 'fake_jwt_token', got %q", jwt)
			}
		},
	})

	runAPITest(t, testCase[string]{
		name:           "UserLogin_invalid_key",
		mockStatus:     403,
		mockResponse:   `{"err":true,"success":false,"message":"Invalid API key","content":""}`,
		expectedPath:   "/user/login",
		expectedMethod: "POST",
		runTest: func(ctx context.Context, client *Client) (string, error) {
			return client.UserLogin(ctx, "invalid key")
		},
		assertError: func(t *testing.T, err error) {
			if err == nil {
				t.Fatal("expected authentication error, got nil")
			}
		},
	})
}

func TestBalance(t *testing.T) {
	runAPITest(t, testCase[int]{
		name:           "Balance_success",
		mockStatus:     200,
		mockResponse:   `{"err":false,"success":true,"message":"","content":{"balance":1250}}`,
		expectedPath:   "/user/balance",
		expectedMethod: "GET",
		runTest: func(ctx context.Context, client *Client) (int, error) {
			return client.Balance(ctx)
		},
		assertResponse: func(t *testing.T, balance int) {
			if balance != 1250 {
				t.Errorf("expected balance 1250, got %d", balance)
			}
		},
	})

	runAPITest(t, testCase[int]{
		name:           "Balance_server_error",
		mockStatus:     500,
		mockResponse:   `{"err":true,"success":false,"message":"Internal server error","content":null}`,
		expectedPath:   "/user/balance",
		expectedMethod: "GET",
		runTest: func(ctx context.Context, client *Client) (int, error) {
			return client.Balance(ctx)
		},
		assertError: func(t *testing.T, err error) {
			if err == nil {
				t.Fatal("expected server error, got nil")
			}
		},
	})

	runAPITest(t, testCase[int]{
		name:           "Balance_malformed_json",
		mockStatus:     200,
		mockResponse:   `{"err":false,"success":true,"message":"","content":{"balance":"not_an_int"}}`,
		expectedPath:   "/user/balance",
		expectedMethod: "GET",
		runTest: func(ctx context.Context, client *Client) (int, error) {
			return client.Balance(ctx)
		},
		assertError: func(t *testing.T, err error) {
			if err == nil {
				t.Fatal("expected JSON decode error, got nil")
			}
		},
	})
}

func TestTransactionHistory(t *testing.T) {
	runAPITest(t, testCase[InventoryPayload]{
		name:           "TransactionHistory_success",
		mockStatus:     200,
		mockResponse:   `{"err":false,"success":true,"message":"","content":{"count":1,"values":[{"count":1,"date":1704067200,"effect":"","festivized":0,"idbackpack":1,"iditem":440,"image":"","inspect":"","killstreaker":"","level":1,"name":"Test Item","paint":"","parts":"","price":5000,"sheen":"","spell":"","url":""}]}}`,
		expectedPath:   "/user/getTransactionHistory",
		expectedMethod: "GET",
		runTest: func(ctx context.Context, client *Client) (InventoryPayload, error) {
			return client.TransactionHistory(ctx, &HistoryOptions{Page: 1, Limit: 10})
		},
		assertResponse: func(t *testing.T, res InventoryPayload) {
			if res.Count != 1 {
				t.Errorf("expected count 1, got %d", res.Count)
			}
		},
	})
}

func TestSalesHistory(t *testing.T) {
	runAPITest(t, testCase[InventoryPayload]{
		name:           "SalesHistory_success",
		mockStatus:     200,
		mockResponse:   `{"err":false,"success":true,"message":"","content":{"count":1,"values":[]}}`,
		expectedPath:   "/user/getSalesHistory",
		expectedMethod: "GET",
		runTest: func(ctx context.Context, client *Client) (InventoryPayload, error) {
			return client.SalesHistory(ctx, &HistoryOptions{Page: 1, Limit: 5, Period: Period3Months, Search: "test"})
		},
		assertResponse: func(_ *testing.T, _ InventoryPayload) {},
	})
}

func TestPurchaseHistory(t *testing.T) {
	runAPITest(t, testCase[InventoryPayload]{
		name:           "PurchaseHistory_success",
		mockStatus:     200,
		mockResponse:   `{"err":false,"success":true,"message":"","content":{"count":1,"values":[]}}`,
		expectedPath:   "/user/getPurchaseHistory",
		expectedMethod: "GET",
		runTest: func(ctx context.Context, client *Client) (InventoryPayload, error) {
			return client.PurchaseHistory(ctx, &HistoryOptions{Page: 1, Limit: 10})
		},
		assertResponse: func(_ *testing.T, _ InventoryPayload) {},
	})
}