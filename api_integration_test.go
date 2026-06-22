//go:build integration

package api_test

import (
	"context"
	"github.com/van-zip/mannco-api"
	"os"
	"testing"
)

func getTestClient(t *testing.T) (*api.Client, context.Context) {
	apiKey := os.Getenv("MANNCO_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping live test, MANNCO_API_KEY environment variable not set")
	}

	client := api.NewClient("", nil)
	ctx := context.Background()

	_, err := client.UserLogin(ctx, apiKey)
	if err != nil {
		t.Fatalf("Integration setup failed: Auth error: %v", err)
	}

	return client, ctx
}

func TestLive_Balance(t *testing.T) {
	client, ctx := getTestClient(t)

	_, err := client.Balance(ctx)
	if err != nil {
		t.Errorf("Live API Balance call failed: %v", err)
	}
}

func TestLive_BuyOrderList(t *testing.T) {
	client, ctx := getTestClient(t)

	_, err := client.BuyOrderList(ctx, 958)
	if err != nil {
		t.Errorf("Live API BuyOrderList call failed: %v", err)
	}
}

func TestLive_SalesHistory_WithFilters(t *testing.T) {
	client, ctx := getTestClient(t)

	opts := &api.HistoryOptions{
		Page:   1,
		Limit:  5,
		Period: api.Period3Months,
	}

	_, err := client.SalesHistory(ctx, opts)
	if err != nil {
		t.Errorf("Live API SalesHistory call failed when using filtering options: %v", err)
	}
}
