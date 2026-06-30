# mannco-go

A Go API client for [Mannco.store](https://mannco.store) a Team Fortress 2, CS2, and Rust item trading marketplace.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

---

## Features

| Category | Coverage |
|----------|----------|
| **Authentication** | API key -> JWT exchange |
| **Items & Pricing** | Sales graphs, listings, buy orders, pricing (single & bulk up to 100 items) |
| **User Buy Orders** | View your active buy orders (specific item or all) |
| **Market Orders** | Create buy orders |
| **User & History** | Balance, transaction / sales / purchase history |
| **Trades & Inventory** | Planned, offers, inventory, deposits, trades |
| **Cart & Checkout** | Planned, cart operations |

> **Status**: This library covers the endpoints relevant to my own project. Many trade, inventory, and cart endpoints are not yet implemented.

---

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

    // Custom HTTP client (optional, nil uses a 60s default)
    httpClient := &http.Client{Timeout: 10 * time.Second}

    // Create client with empty JWT; will be populated after login
    client := mannco.NewClient("", httpClient)

    // Exchange API key for JWT
    apiKey := "your-mannco-store-api-key"
    _, err := client.UserLogin(ctx, apiKey)
    if err != nil {
        log.Fatalf("login failed: %v", err)
    }

    // Check balance (returned in cents)
    balance, err := client.Balance(ctx)
    if err != nil {
        log.Fatalf("balance: %v", err)
    }
    fmt.Printf("Balance: $%.2f\n", float64(balance)/100)

    // Bulk pricing for up to 100 items (Max's Head, Earbuds, Bill's Hat)
    bulk, err := client.ItemPricingBulk(ctx, []int{371, 958, 803})
    if err != nil {
        log.Fatalf("bulk pricing: %v", err)
    }

    for _, item := range bulk.Items {
        fmt.Printf("Item %d | Lowest sale: $%.2f | Suggested: $%.2f\n",
            item.ItemID,
            float64(item.Pricing.LowestSalePrice)/100,
            float64(item.Pricing.SuggestedPrice)/100,
        )
    }
}
```

Run it:

```bash
go run examples/example.go
```

---

## API Reference

### Client

```go
client := mannco.NewClient(jwt string, httpClient *http.Client)
```

| Method | Description |
|--------|-------------|
| `SetJWT(token string)` | Update the bearer token |
| `GetJWT() string` | Retrieve current token |
| `SetBaseURL(url string)` | Override API base URL (useful for testing) |
| `GetBaseURL() string` | Get current base URL |

All methods accept `context.Context` as the first argument for cancellation/timeout control.

---

### Authentication

```go
// POST /user/login
jwt, err := client.UserLogin(ctx, apiKey)
```

Exchanges an API key for a session JWT. The returned token is also stored on the client automatically.

---

### Items & Pricing

```go
// GET /item/pricing/{item}
item, err := client.ItemPricing(ctx, itemID)

// GET /item/pricing/bulk?items=1,2,3 (max 100)
bulk, err := client.ItemPricingBulk(ctx, []int{371, 958, 803})

// GET /item/salesGraph/{item}?period=1M
graph, err := client.ItemSalesGraph(ctx, itemID, mannco.Period1Month)

// GET /item/listing/{item}[/{user}]?count=10&page=1&game=440
listings, err := client.ItemListings(ctx, itemID, "", &mannco.ListingOptions{
    Count: 20,
    Page:  1,
    Game:  440, // TF2
})

// GET /item/buyorderList/{item}
buyOrders, err := client.BuyOrderList(ctx, itemID)

// POST /item/buyorder
err := client.CreateBuyOrder(ctx, 371, 1500, 1) // itemID, price (cents), quantity
```

**Pricing periods** (`mannco.Period`): `Period1Month`, `Period3Months`, `Period6Months`, `Period1Year`, `Period5Years`, `PeriodAll`.

---

### User & History

```go
// GET /user/balance
balance, err := client.Balance(ctx) // returns cents

// GET /user/getTransactionHistory
// Note: HistoryOptions.Limit maps to "limit" query param
txns, err := client.TransactionHistory(ctx, &mannco.HistoryOptions{
    Page:  1,
    Limit: 50,
})

// GET /user/getSalesHistory
// Note: HistoryOptions.Limit maps to "perPage" query param
sales, err := client.SalesHistory(ctx, &mannco.HistoryOptions{
    Page:    1,
    Limit:   50,
    Period:  mannco.Period1Month,
    Search:  "unusual",
})

// GET /user/getPurchaseHistory
// Note: HistoryOptions.Limit maps to "count" query param
purchases, err := client.PurchaseHistory(ctx, &mannco.HistoryOptions{
    Page:  1,
    Limit: 50,
})
```

---

### User Buy Orders

```go
// GET /user/buyorder/{itemID}
buyOrder, err := client.UserItemBuyOrder(ctx, 371)

// GET /user/buyorders
buyOrders, err := client.GetUserBuyOrders(ctx)
```

---

## Testing

```bash
# Unit tests (mock HTTP)
go test -v ./...

# Integration tests (require MANNCO_API_KEY in .env or env)
go test -v -tags=integration ./...
```

Integration tests require a valid API key in the environment

```bash
MANNCO_API_KEY=your_key_here go test -v -tags=integration ./...
```

---

## License

MIT License — see [LICENSE](LICENSE).

---

PRs welcome for missing endpoints!
