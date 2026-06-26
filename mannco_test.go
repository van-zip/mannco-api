package mannco

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

type testCase[ResponsePayload any] struct {
	name           string
	mockStatus     int
	mockResponse   string
	expectedPath   string
	expectedMethod string
	runTest        func(ctx context.Context, client *Client) (ResponsePayload, error)
	assertResponse func(t *testing.T, res ResponsePayload)
	assertError    func(t *testing.T, err error)
}

func runAPITest[T any](t *testing.T, tc testCase[T]) {
	t.Run(tc.name, func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != tc.expectedMethod {
				t.Errorf("expected %s request, got %s", tc.expectedMethod, r.Method)
			}
			if r.URL.Path != tc.expectedPath {
				t.Errorf("expected path %s, got %s", tc.expectedPath, r.URL.Path)
			}
			if r.Header.Get("Authorization") != "Bearer fake_token" {
				t.Errorf("Expected Auth header 'Bearer fake_token', got '%s'", r.Header.Get("Authorization"))
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(tc.mockStatus)
			// this failing shouldn't matter for a test suite
			_, _ = w.Write([]byte(tc.mockResponse))
		}))
		defer server.Close()

		client := NewClient("fake_token", nil)
		client.SetBaseURL(server.URL + "/")

		resp, err := tc.runTest(context.Background(), client)
		if tc.assertError != nil {
			tc.assertError(t, err)
		} else if err != nil {
			t.Fatalf("unexpected operational error execution: %v", err)
		}

		if tc.assertResponse != nil {
			tc.assertResponse(t, resp)
		}
	})
}
func TestUserLogin(t *testing.T) {
	tests := []testCase[string]{
		{
			name:           "Successful Login Flow",
			mockStatus:     http.StatusOK,
			mockResponse:   `{"err":false,"success":true,"message":"","content":{"jwt":"fake_token"}}`,
			expectedPath:   "/user/login",
			expectedMethod: http.MethodPost,
			runTest: func(ctx context.Context, client *Client) (string, error) {
				return client.UserLogin(ctx, "valid_api_key_123")
			},
			assertResponse: func(t *testing.T, jwt string) {
				expectedToken := "fake_token"
				if jwt != expectedToken {
					t.Errorf("expected JWT to be %q, got %q", expectedToken, jwt)
				}
			},
		},
		{
			name:           "Invalid API key",
			mockStatus:     http.StatusForbidden,
			mockResponse:   `{"err":true,"success":false,"message":"Invalid API key","content":""}`,
			expectedPath:   "/user/login",
			expectedMethod: http.MethodPost,
			runTest: func(ctx context.Context, client *Client) (string, error) {
				return client.UserLogin(ctx, "invalid key")
			},
			assertResponse: nil,
			assertError: func(t *testing.T, err error) {
				if err == nil {
					t.Fatal("expected an authentication failure error, but got nil")
				}
				expectedErrMsg := "authentication error: server error with status code 403: Invalid API key"
				if err.Error() != expectedErrMsg {
					t.Errorf("expected error message %q, got %q", expectedErrMsg, err.Error())
				}
			},
		},
	}

	for _, tc := range tests {
		runAPITest(t, tc)
	}
}
func TestBalance(t *testing.T) {
	tests := []testCase[int]{
		{
			name:           "Successful Balance Retrieval",
			mockStatus:     http.StatusOK,
			mockResponse:   `{"err": false, "success": true, "content": {"balance": 1250}}`,
			expectedPath:   "/user/balance",
			expectedMethod: http.MethodGet,
			runTest: func(ctx context.Context, client *Client) (int, error) {
				return client.Balance(ctx)
			},
			assertResponse: func(t *testing.T, balance int) {
				if balance != 1250 {
					t.Errorf("expected 1250 pennies, got %d", balance)
				}
			},
		},
		{
			name:           "Server Internal Error Handling",
			mockStatus:     http.StatusInternalServerError,
			mockResponse:   `{"err": true, "success": false}`,
			expectedPath:   "/user/balance",
			expectedMethod: http.MethodGet,
			runTest: func(ctx context.Context, client *Client) (int, error) {
				return client.Balance(ctx)
			},
			assertError: func(t *testing.T, err error) {
				if err == nil {
					t.Fatal("expected server error registration but received nil")
				}
			},
		},
		{
			name:           "Malformed JSON Payload Handling",
			mockStatus:     http.StatusOK,
			mockResponse:   `{"err": false, "success": true, "content": { "balance": "this-should-be-an-int-not-a-string" }}`,
			expectedPath:   "/user/balance",
			expectedMethod: http.MethodGet,
			runTest: func(ctx context.Context, client *Client) (int, error) {
				return client.Balance(ctx)
			},
			assertError: func(t *testing.T, err error) {
				if err == nil {
					t.Fatal("expected a JSON unmarshaling failure error, but got nil")
				}
				if !strings.Contains(err.Error(), "failed decoding response JSON") {
					t.Errorf("expected JSON decoding error message context, got: %q", err.Error())
				}
			},
		},
	}

	for _, tc := range tests {
		runAPITest(t, tc)
	}
}

func TestBuyOrderPayload_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    []BuyOrderInfo
		expectError bool
	}{
		{
			name: "array shape",
			input: `{
				"informations": [
					{"count": 1, "price": 9000},
					{"count": 1, "price": 8810}
				]
			}`,
			expected: []BuyOrderInfo{
				{Count: 1, Price: 9000},
				{Count: 1, Price: 8810},
			},
		},
		{
			name: "object shape with more, keys in order",
			input: `{
				"informations": {
					"0": {"count": 401, "price": 170},
					"1": {"count": 994, "price": 169},
					"2": {"count": 1124, "price": 168},
					"3": {"count": 1874, "price": 167},
					"4": {"count": 1138, "price": 166},
					"more": {"count": 4601, "price": 165}
				}
			}`,
			expected: []BuyOrderInfo{
				{Count: 401, Price: 170},
				{Count: 994, Price: 169},
				{Count: 1124, Price: 168},
				{Count: 1874, Price: 167},
				{Count: 1138, Price: 166},
				{Count: 4601, Price: 165},
			},
		},
		{
			name:     "empty informations object",
			input:    `{"informations": {}}`,
			expected: []BuyOrderInfo{},
		},
		{
			name:     "empty informations array",
			input:    `{"informations": []}`,
			expected: []BuyOrderInfo{},
		},
		{
			name: "unexpected non-numeric key",
			input: `{
				"informations": {
					"0": {"count": 1, "price": 100},
					"unexpected": {"count": 1, "price": 90}
				}
			}`,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var payload BuyOrderPayload
			err := json.Unmarshal([]byte(tc.input), &payload)

			if tc.expectError {
				if err == nil {
					t.Fatal("expected an error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(payload.BuyOrders, tc.expected) {
				t.Errorf("got = %+v, want %+v", payload.BuyOrders, tc.expected)
			}
		})
	}
}
func TestBuyOrderList(t *testing.T) {
	tests := []testCase[BuyOrderPayload]{
		{
			name:       "Array-shaped response",
			mockStatus: http.StatusOK,
			mockResponse: `{
				"err": false,
				"success": true,
				"content": {
					"informations": [
						{"count": 1, "price": 9000},
						{"count": 1, "price": 8810}
					]
				}
			}`,
			expectedPath:   "/item/buyorderList/958",
			expectedMethod: http.MethodGet,
			runTest: func(ctx context.Context, client *Client) (BuyOrderPayload, error) {
				return client.BuyOrderList(ctx, 958)
			},
			assertResponse: func(t *testing.T, got BuyOrderPayload) {
				expected := BuyOrderPayload{
					BuyOrders: []BuyOrderInfo{
						{Count: 1, Price: 9000},
						{Count: 1, Price: 8810},
					},
				}
				if !reflect.DeepEqual(got, expected) {
					t.Errorf("BuyOrderList() got = %+v, want %+v", got, expected)
				}
			},
		},
		{
			name:       "Object-shaped response with overflow tier",
			mockStatus: http.StatusOK,
			mockResponse: `{
				"err": false,
				"success": true,
				"content": {
					"informations": {
						"0": {"count": 401, "price": 170},
						"1": {"count": 994, "price": 169},
						"more": {"count": 4601, "price": 165}
					}
				}
			}`,
			expectedPath:   "/item/buyorderList/371",
			expectedMethod: http.MethodGet,
			runTest: func(ctx context.Context, client *Client) (BuyOrderPayload, error) {
				return client.BuyOrderList(ctx, 371)
			},
			assertResponse: func(t *testing.T, got BuyOrderPayload) {
				expected := BuyOrderPayload{
					BuyOrders: []BuyOrderInfo{
						{Count: 401, Price: 170},
						{Count: 994, Price: 169},
						{Count: 4601, Price: 165},
					},
				}
				if !reflect.DeepEqual(got, expected) {
					t.Errorf("BuyOrderList() got = %+v, want %+v", got, expected)
				}
			},
		},
	}

	for _, tc := range tests {
		runAPITest(t, tc)
	}
}
