package mannco

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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

// TestExecuteRequestErrorPaths tests error paths in executeRequest
func TestExecuteRequestErrorPaths(t *testing.T) {
	// Test network error
	t.Run("network_error", func(t *testing.T) {
		client := NewClient("fake_token", &http.Client{
			Transport: &failingTransport{},
		})

		_, err := client.Balance(context.Background())
		if err == nil {
			t.Fatal("expected network error, got nil")
		}
		if !errors.Is(err, ErrNetwork) {
			t.Errorf("expected ErrNetwork, got %v", err)
		}
	})

	// Test non-200 status with API error message
	t.Run("api_error_404", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"err":true,"success":false,"message":"Not found","content":null}`))
		}))
		defer server.Close()

		client := NewClient("fake_token", nil)
		client.SetBaseURL(server.URL + "/")

		_, err := client.Balance(context.Background())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Errorf("expected *APIError, got %T: %v", err, err)
		}
		if apiErr.StatusCode != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", apiErr.StatusCode)
		}
	})

	// Test 401 unauthorized
	t.Run("unauthorized_401", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"err":true,"success":false,"message":"Unauthorized","content":null}`))
		}))
		defer server.Close()

		client := NewClient("fake_token", nil)
		client.SetBaseURL(server.URL + "/")

		_, err := client.Balance(context.Background())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrUnauthorized) {
			t.Errorf("expected ErrUnauthorized, got %v", err)
		}
	})

	// Test 403 forbidden
	t.Run("forbidden_403", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"err":true,"success":false,"message":"Forbidden","content":null}`))
		}))
		defer server.Close()

		client := NewClient("fake_token", nil)
		client.SetBaseURL(server.URL + "/")

		_, err := client.Balance(context.Background())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrUnauthorized) {
			t.Errorf("expected ErrUnauthorized, got %v", err)
		}
	})

	// Test malformed JSON response
	t.Run("malformed_json", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`not valid json`))
		}))
		defer server.Close()

		client := NewClient("fake_token", nil)
		client.SetBaseURL(server.URL + "/")

		_, err := client.Balance(context.Background())
		if err == nil {
			t.Fatal("expected JSON decode error, got nil")
		}
		if !strings.Contains(err.Error(), "failed decoding response JSON") {
			t.Errorf("expected JSON decode error, got %v", err)
		}
	})

	// Test API response with err=true
	t.Run("api_response_err_true", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"err":true,"success":false,"message":"API error","content":null}`))
		}))
		defer server.Close()

		client := NewClient("fake_token", nil)
		client.SetBaseURL(server.URL + "/")

		_, err := client.Balance(context.Background())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "API error") {
			t.Errorf("expected API error message, got %v", err)
		}
	})

	// Test API response with success=false
	t.Run("api_response_success_false", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"err":false,"success":false,"message":"Operation failed","content":null}`))
		}))
		defer server.Close()

		client := NewClient("fake_token", nil)
		client.SetBaseURL(server.URL + "/")

		_, err := client.Balance(context.Background())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "Operation failed") {
			t.Errorf("expected operation failed message, got %v", err)
		}
	})

	// Test 500 internal server error
	t.Run("server_error_500", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"err":true,"success":false,"message":"Internal server error","content":null}`))
		}))
		defer server.Close()

		client := NewClient("fake_token", nil)
		client.SetBaseURL(server.URL + "/")

		_, err := client.Balance(context.Background())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Errorf("expected *APIError, got %T: %v", err, err)
		}
		if apiErr.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", apiErr.StatusCode)
		}
	})

	// Test invalid URL endpoint
	t.Run("invalid_endpoint_url", func(t *testing.T) {
		client := NewClient("fake_token", nil)
		client.SetBaseURL("not-a-valid-url")

		_, err := client.Balance(context.Background())
		if err == nil {
			t.Fatal("expected error for invalid URL, got nil")
		}
		// URL parsing happens at request time, error message varies
		if !strings.Contains(err.Error(), "invalid endpoint url") && !strings.Contains(err.Error(), "unsupported protocol scheme") {
			t.Errorf("expected URL error, got %v", err)
		}
	})

	// Test context cancellation
	t.Run("context_cancelled", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate slow response
			time.Sleep(100 * time.Millisecond)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"err":false,"success":true,"message":"","content":{"balance":100}}`))
		}))
		defer server.Close()

		client := NewClient("fake_token", nil)
		client.SetBaseURL(server.URL + "/")

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := client.Balance(ctx)
		if err == nil {
			t.Fatal("expected context cancelled error, got nil")
		}
		if !errors.Is(err, context.Canceled) && !strings.Contains(err.Error(), "context canceled") {
			t.Errorf("expected context cancelled error, got %v", err)
		}
	})
}

// failingTransport is a transport that always fails
type failingTransport struct{}

func (f *failingTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("connection refused")
}