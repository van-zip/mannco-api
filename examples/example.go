package example

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
	jwtToken := "your_jwt_token_here"
	client := api.NewClient(jwtToken, customHTTPClient)

	fmt.Println("--- Fetching User Balance ---")
	balance, err := client.Balance(ctx)
	if err != nil {
		log.Fatalf("Failed to retrieve balance: %v", err)
	}

	fmt.Printf("Current Account Balance: $%.2f\n\n", float64(balance)/100.0)

	fmt.Println("--- Fetching 6-Month Sales History (Filtered) ---")
	salesOpts := &api.HistoryOptions{
		Page:   1,
		Limit:  5,
		Period: api.Period6Months,
		Search: "Key",
	}

	salesHistory, err := client.SalesHistory(ctx, salesOpts)
	if err != nil {
		log.Fatalf("Failed to retrieve sales history: %v", err)
	}

	fmt.Printf("Found %d total historical sales matching criteria. Showing first page:\n", salesHistory.Count)
	for i, item := range salesHistory.Values {
		fmt.Printf("%d. [%s] Sold for: $%.2f (ID: %d)\n",
			i+1,
			item.Name,
			float64(item.Price)/100.0,
			item.IDItem,
		)
	}
	fmt.Println()

	fmt.Println("--- Fetching Bulk Pricing Data ---")

	targetIDs := []int{215, 30015, 1024}
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
		fmt.Printf(" - Item ID %d | Lowest Sale Listing: $%.2f | Suggested Value: $%.2f\n",
			item.ItemID,
			float64(item.Pricing.LowestSalePrice)/100.0,
			float64(item.Pricing.SuggestedPrice)/100.0,
		)
	}
}
