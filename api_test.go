package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBalance(t *testing.T) {
	fakeJSONResponse := `{
		"err": false,
		"success": true,
		"message": "",
		"content": {
			"balance": 5000
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer fake_token" {
			t.Errorf("Expected Auth header 'Bearer fake_token', got '%s'", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fakeJSONResponse))
	}))
	defer server.Close()

	client := NewClient("fake_token", nil)
	client.SetBaseURL(server.URL + "/")

	balance, err := client.Balance(context.Background())

	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	expectedBalance := 5000
	if balance != expectedBalance {
		t.Errorf("Expected balance to be %d pennies, but got %d", expectedBalance, balance)
	}
}
