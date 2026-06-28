package mannco

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