package mannco

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// Period represents the valid time frames
type Period string

// All valid period values
const (
	Period1Month  Period = "1M"
	Period3Months Period = "3M"
	Period6Months Period = "6M"
	Period1Year   Period = "1Y"
	Period5Years  Period = "5Y"
	PeriodAll     Period = "ALL"
)

// A PriceHistoryPoint is a specific data point from the price history of an item
type PriceHistoryPoint struct {
	Date  string `json:"date"`
	Price int    `json:"price"`
	Nb    int    `json:"nb"`
}

// PriceHistoryPayload is the payload returned by the price history call
type PriceHistoryPayload struct {
	Values []PriceHistoryPoint `json:"values"`
}

// LoginPayload is the payload returned by the login call
type LoginPayload struct {
	JWT string `json:"jwt"`
}

// BalancePayload is the payload returned by the balance call
type BalancePayload struct {
	Balance int `json:"balance"`
}

// maxIDs is the max number of IDs able to be provided to bulk endpoints
const maxIDs = 100

// An Item represents a specific item and all its attributes
type Item struct {
	Count        int    `json:"count"`
	Date         int64  `json:"date"`
	Effect       string `json:"effect"`
	Festivized   int    `json:"festivized"`
	IDBackpack   int    `json:"idbackpack"`
	IDItem       int    `json:"iditem"`
	Image        string `json:"image"`
	Inspect      string `json:"inspect"`
	Killstreaker string `json:"killstreaker"`
	Level        int    `json:"level"`
	Name         string `json:"name"`
	Paint        string `json:"paint"`
	Parts        string `json:"parts"`
	Price        int    `json:"price"`
	Sheen        string `json:"sheen"`
	Spell        string `json:"spell"`
	URL          string `json:"url"`
}

// LastSale contains sale price and date
type LastSale struct {
	Date  int64 `json:"date"`
	Price int   `json:"price"`
}

// A PricingData contains information regarding an item's pricing information
type PricingData struct {
	LastSale        LastSale `json:"last_sale"`
	LowestBuyOrder  int      `json:"lowest_buy_order"`
	LowestSalePrice int      `json:"lowest_sale_price"`
	SteamPrice      int      `json:"steam_price"`
	SuggestedPrice  int      `json:"suggested_price"`
}

// A PriceItem wraps an item's PricingData with additional API relevant information
type PriceItem struct {
	FromCache   bool        `json:"from_cache"`
	ItemID      int         `json:"item_id"`
	LastUpdated string      `json:"last_updated"`
	Pricing     PricingData `json:"pricing"`
}

// BulkPricingPayload is the payload returned by ItemPricingBulk
type BulkPricingPayload struct {
	CachedItems    int         `json:"cached_items"`
	Items          []PriceItem `json:"items"`
	RefreshedItems int         `json:"refreshed_items"`
	TotalItems     int         `json:"total_items"`
}

// InventoryPayload is the payload returned by API calls that return multiple items
type InventoryPayload struct {
	Count  int    `json:"count"`
	Values []Item `json:"values"`
}

// Listing is a datatype which represents a specific listing for an item
type Listing struct {
	AssetID      string  `json:"assetId"`
	Bot          string  `json:"bot"`
	Game         int     `json:"game"`
	GetImage     string  `json:"getImage"`
	HTML         string  `json:"html"`
	ID           int     `json:"id"`
	ItemID       int     `json:"item_id"`
	Killstreaker string  `json:"killstreaker"`
	Paint        string  `json:"paint"`
	Parts        string  `json:"parts"`
	Price        int     `json:"price"`
	Sheen        string  `json:"sheen"`
	Spell        string  `json:"spell"`
	State        int     `json:"state"`
	User         string  `json:"user"`
	Wear         float64 `json:"wear"`
}

// ListingPayload is the payload returned by ItemListings
type ListingPayload struct {
	Count  int       `json:"count"`
	Values []Listing `json:"values"`
}

// HistoryOptions is an optional struct for filtering on some time relevant endpoints. Note that upstream, limit is represented by either count or per page
type HistoryOptions struct {
	Page   int
	Limit  int
	Period Period
	Search string
}

// ListingOptions is an optional struct for filtering listings
type ListingOptions struct {
	Count int
	Page  int
	Game  int
}

// ItemPricing gets item pricing data for a specific item ID
func (c *Client) ItemPricing(ctx context.Context, itemID int) (PriceItem, error) {
	return ExecuteRequest[PriceItem](ctx, c, "GET", "item/pricing/"+strconv.Itoa(itemID), nil, nil)
}

// ItemPricingBulk performs ItemPricing() but on up to 100 itemIDs
func (c *Client) ItemPricingBulk(ctx context.Context, itemIDs []int) (BulkPricingPayload, error) {
	if len(itemIDs) > maxIDs {
		return BulkPricingPayload{}, fmt.Errorf("%w: the pricing bulk endpoint has a limit of 100 ids", ErrInternal)
	}
	idStrings := make([]string, len(itemIDs))
	for i, id := range itemIDs {
		idStrings[i] = strconv.Itoa(id)
	}
	params := url.Values{}
	params.Add("items", strings.Join(idStrings, ","))
	return ExecuteRequest[BulkPricingPayload](ctx, c, "GET", "item/pricing/bulk", nil, params)
}

// ItemSalesGraph returns sales history for a given item ID over some period
func (c *Client) ItemSalesGraph(ctx context.Context, itemID int, period Period) (PriceHistoryPayload, error) {
	if period == "" {
		period = Period1Month
	}
	params := url.Values{}
	params.Add("period", string(period))
	return ExecuteRequest[PriceHistoryPayload](ctx, c, "GET", "item/salesGraph/"+strconv.Itoa(itemID), nil, params)
}

// ItemListings returns the active listings (with optional filtering) for a given item ID
func (c *Client) ItemListings(ctx context.Context, itemID int, userID string, opts *ListingOptions) (ListingPayload, error) {
	endpoint := "item/listing/" + strconv.Itoa(itemID)
	if userID != "" {
		endpoint += "/" + userID
	}

	params := url.Values{}
	if opts != nil {
		if opts.Count > 0 {
			params.Add("count", strconv.Itoa(opts.Count))
		}
		if opts.Page > 0 {
			params.Add("page", strconv.Itoa(opts.Page))
		}
		if opts.Game > 0 {
			params.Add("game", strconv.Itoa(opts.Game))
		}
	}

	return ExecuteRequest[ListingPayload](ctx, c, "GET", endpoint, nil, params)
}
