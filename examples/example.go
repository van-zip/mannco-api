package main

import (
	"context"
	"fmt"
	"github.com/van-zip/mannco-api"
	"log"
	"net/http"
	"time"
)

func main() {
	// initialize a context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// setup a HTTP client (or pass nil for a default one)
	customHTTPClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	// instantiate your  API client
	client := api.NewClient("", customHTTPClient)

	// fetch your auth token
	fmt.Println("--- Fetching JWT ---")
	apiKey := "xxxxxxxxxxxxxxxxxxxxxxxxx"
	_, err := client.UserLogin(ctx, apiKey)
	if err != nil {
		log.Fatalf("Login failed: %v", err)
	}

	fmt.Println("--- Fetching User Balance ---")
	balance, err := client.Balance(ctx)
	if err != nil {
		log.Fatalf("Failed to retrieve balance: %v", err)
	}

	fmt.Printf("Current Account Balance: $%.2f\n\n", float64(balance)/100.0)

	fmt.Println("--- Fetching Bulk Pricing Data ---")

	// Max's severed head, Earbuds, Bill's hat ids
	targetIDs := []int{371, 958, 803}
	bulkData, err := client.ItemPricingBulk(ctx, targetIDs)
	if err != nil {
		log.Printf("Warning: Failed bulk pricing call: %v", err)
		return
	}

	fmt.Printf("Successfully analyzed %d items (%d from fresh live updates):\n",
		bulkData.TotalItems,
		bulkData.RefreshedItems,
	)
	for _, item := range bulkData.Items {
		fmt.Printf(" - Item ID %d | Lowest Sale: $%.2f | Suggested Value: $%.2f\n",
			item.ItemID,
			float64(item.Pricing.LowestSalePrice)/100.0,
			float64(item.Pricing.SuggestedPrice)/100.0,
		)
	}
}
