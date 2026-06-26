//go:build integration

package mannco_test

import (
	"context"
	"github.com/joho/godotenv"
	"github.com/van-zip/mannco-go"
	"os"
	"testing"
)

func getTestClient(t *testing.T) (*mannco.Client, context.Context) {
	err := godotenv.Load()
	if err != nil {
		t.Fatal("Error loading .env file")
	}
	apiKey := os.Getenv("MANNCO_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping live test, MANNCO_API_KEY environment variable not set")
	}

	client := mannco.NewClient("", nil)
	ctx := context.Background()

	_, err = client.UserLogin(ctx, apiKey)
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

	opts := &mannco.HistoryOptions{
		Page:   1,
		Limit:  5,
		Period: mannco.Period3Months,
	}

	_, err := client.SalesHistory(ctx, opts)
	if err != nil {
		t.Errorf("Live API SalesHistory call failed when using filtering options: %v", err)
	}
}
