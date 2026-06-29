# Mannco.store API Wrapper for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/van-zip/mannco-go.svg)](https://pkg.go.dev/github.com/van-zip/mannco-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/van-zip/mannco-go)](https://goreportcard.com/report/github.com/van-zip/mannco-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A Go client library for the [Mannco.store API](https://docs.mannco.store).

> ⚠️ **Not a complete library** — covers the endpoints I needed. PRs welcome.

## Installation

```bash
go get github.com/van-zip/mannco-go
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/van-zip/mannco-go"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := mannco.NewClient("", &http.Client{Timeout: 10 * time.Second})

	// Login with API key
	_, err := client.UserLogin(ctx, "your-api-key")
	if err != nil {
		log.Fatalf("Login failed: %v", err)
	}

	// Get balance
	balance, err := client.Balance(ctx)
	if err != nil {
		log.Fatalf("Balance failed: %v", err)
	}
	fmt.Printf("Balance: $%.2f\n", float64(balance)/100)

	// Bulk pricing
	ids := []int{371, 958, 803} // Max's Severed Head, Earbuds, Bill's Hat
	data, err := client.ItemPricingBulk(ctx, ids)
	if err != nil {
		log.Printf("Bulk pricing failed: %v", err)
		return
	}

	for _, item := range data.Items {
		fmt.Printf("Item %d: Lowest Sale $%.2f | Suggested $%.2f\n",
			item.ItemID,
			float64(item.Pricing.LowestSalePrice)/100,
			float64(item.Pricing.SuggestedPrice)/100,
		)
	}
}
```

## Features

| Category | Endpoints |
|----------|-----------|
| **Auth** | `UserLogin` (API key → JWT) |
| **Pricing** | `ItemPricing`, `ItemPricingBulk` (max 100 IDs), `ItemSalesGraph` |
| **Listings** | `ItemListings` (with pagination/user/game filters) |
| **Buy Orders** | `BuyOrderList`, `CreateBuyOrder` |
| **User** | `Balance`, `TransactionHistory`, `SalesHistory`, `PurchaseHistory` |

## Error Handling

All errors support `errors.Is` / `errors.As`:

```go
_, err := client.BuyOrderList(ctx, 958)
if errors.Is(err, mannco.ErrUnauthorized) {
    // Token expired or invalid
} else if errors.Is(err, mannco.ErrNetwork) {
    // Network failure
} else if errors.Is(err, mannco.ErrInternal) {
    // API returned non-200 or JSON decode failed
}

var apiErr *mannco.APIError
if errors.As(err, &apiErr) {
    fmt.Printf("Status: %d, Message: %s\n", apiErr.StatusCode, apiErr.Message)
}
```

## Testing

```bash
# Unit tests (no API key needed)
go test ./...

# Integration tests (requires MANNCO_API_KEY in .env)
go test -tags=integration ./...
```

## Configuration

```go
client := mannco.NewClient("", nil)
// or with custom HTTP client
client := mannco.NewClient("", &http.Client{
    Timeout: 15 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns: 100,
    },
})

// Change base URL (for testing/staging)
client.SetBaseURL("https://api-staging.mannco.store/")

// Manually set/rotate JWT
client.SetJWT("new-token")
```

## License

MIT — see [LICENSE](LICENSE).