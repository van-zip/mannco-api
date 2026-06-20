package api

import (
	"context"
	"net/http"
	"net/http/httptest"
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
			w.Write([]byte(tc.mockResponse))
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
				expectedErrMsg := "server rejected request with status code 403: Invalid API key"
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
	}

	for _, tc := range tests {
		runAPITest(t, tc)
	}
}
