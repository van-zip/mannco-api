package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const BaseURL = "https://api.mannco.store/"

/*
	Datatypes
*/

// period represents the valid time frames
type Period string

const (
	Period1Month  Period = "1M"
	Period3Months Period = "3M"
	Period6Months Period = "6M"
	Period1Year   Period = "1Y"
	Period5Years  Period = "5Y"
	PeriodAll     Period = "ALL"
)

type ApiResponse[T any] struct {
	Err     bool   `json:"err"`
	Success bool   `json:"success"`
	Message string `json:"message"`
	Content T      `json:"content"`
}

type PriceHistoryPoint struct {
	Date  string `json:"date"`
	Price int    `json:"price"`
	Nb    int    `json:"nb"`
}

type PriceHistoryPayload struct {
	Values []PriceHistoryPoint `json:"values"`
}

type LoginPayload struct {
	JWT string `json:"jwt"`
}

type BalancePayload struct {
	Balance int `json:"balance"`
}

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

type LastSale struct {
	Date  int64 `json:"date"`
	Price int   `json:"price"`
}

type PricingData struct {
	LastSale        LastSale `json:"last_sale"`
	LowestBuyOrder  int      `json:"lowest_buy_order"`
	LowestSalePrice int      `json:"lowest_sale_price"`
	SteamPrice      int      `json:"steam_price"`
	SuggestedPrice  int      `json:"suggested_price"`
}

type PriceItem struct {
	FromCache   bool        `json:"from_cache"`
	ItemID      int         `json:"item_id"`
	LastUpdated string      `json:"last_updated"`
	Pricing     PricingData `json:"pricing"`
}

type BulkPricingPayload struct {
	CachedItems    int         `json:"cached_items"`
	Items          []PriceItem `json:"items"`
	RefreshedItems int         `json:"refreshed_items"`
	TotalItems     int         `json:"total_items"`
}

type InventoryPayload struct {
	Count  int    `json:"count"`
	Values []Item `json:"values"`
}

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

type ListingPayload struct {
	Count  int       `json:"count"`
	Values []Listing `json:"values"`
}

type BuyOrderInfo struct {
	Count int `json:"count"`
	Price int `json:"price"`
}

type BuyOrderPayload struct {
	Informations map[string]BuyOrderInfo `json:"informations"`
}

type buyOrderRequest struct {
	ItemID int `json:"itemid"`
	Value  int `json:"value"`
	Amount int `json:"amount"`
}

type HistoryOptions struct {
	Page   int
	Limit  int // either count or per page b/c mannco is inconsistent
	Period Period
	Search string
}

type ListingOptions struct {
	Count int
	Page  int
	Game  int
}

type Client struct {
	httpClient *http.Client
	mu         sync.RWMutex
	baseURL    string
	jwt        string
}

/*
	Library specific functionality
*/

func NewClient(jwt string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 60 * time.Second,
		}
	}
	return &Client{
		baseURL:    BaseURL,
		jwt:        jwt,
		httpClient: httpClient,
	}
}

func (c *Client) SetJWT(token string) {
	c.mu.Lock()
	c.jwt = token
	c.mu.Unlock()
}
func (c *Client) GetJWT() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.jwt
}

// executeRequest performs generic parsing, safety handling, and raw IO operations
func executeRequest[T any](ctx context.Context, c *Client, method, endpoint string, body []byte, queryParams url.Values) (T, error) {
	var target T

	u, err := url.Parse(c.GetBaseURL() + endpoint)
	if err != nil {
		return target, fmt.Errorf("invalid endpoint url: %w", err)
	}
	if queryParams != nil {
		u.RawQuery = queryParams.Encode()
	}

	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewBuffer(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), reqBody)
	if err != nil {
		return target, fmt.Errorf("failed to construct request object: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if jwt := c.GetJWT(); jwt != "" {
		req.Header.Set("Authorization", "Bearer "+jwt)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return target, fmt.Errorf("network call execution failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return target, fmt.Errorf("failed reading raw response bytes: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr ApiResponse[json.RawMessage]
		if json.Unmarshal(bodyBytes, &apiErr) == nil && apiErr.Message != "" {
			return target, fmt.Errorf("server rejected request with status code %d: %s", resp.StatusCode, apiErr.Message)
		}
		return target, fmt.Errorf("server rejected request with status code: %d", resp.StatusCode)
	}

	var apiResponse ApiResponse[T]
	if err = json.Unmarshal(bodyBytes, &apiResponse); err != nil {
		return target, fmt.Errorf("failed decoding response JSON: %w", err)
	}

	if apiResponse.Err || !apiResponse.Success {
		if apiResponse.Message != "" {
			return target, fmt.Errorf("api error: %s", apiResponse.Message)
		}
		return target, fmt.Errorf("api returned a failure state")
	}

	return apiResponse.Content, nil
}

/*
Endpoints
*/

func (c *Client) UserLogin(ctx context.Context, apiKey string) (string, error) {
	data := map[string]string{"apiKey": apiKey}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("error encoding json for user login: %w", err)
	}

	content, err := executeRequest[LoginPayload](ctx, c, "POST", "user/login", jsonData, nil)
	if err != nil {
		return "", err
	}

	c.SetJWT(content.JWT)

	return content.JWT, nil
}

func (c *Client) Balance(ctx context.Context) (int, error) {
	content, err := executeRequest[BalancePayload](ctx, c, "GET", "user/balance", nil, nil)
	if err != nil {
		return 0, err
	}
	return content.Balance, nil
}

func (c *Client) TransactionHistory(ctx context.Context, opts *HistoryOptions) (InventoryPayload, error) {
	params := url.Values{}
	if opts != nil {
		if opts.Page > 0 {
			params.Add("page", strconv.Itoa(opts.Page))
		}
		if opts.Limit > 0 {
			params.Add("limit", strconv.Itoa(opts.Limit))
		}
	}
	return executeRequest[InventoryPayload](ctx, c, "GET", "user/getTransactionHistory", nil, params)
}

func (c *Client) SalesHistory(ctx context.Context, opts *HistoryOptions) (InventoryPayload, error) {
	params := url.Values{}
	if opts != nil {
		if opts.Page > 0 {
			params.Add("page", strconv.Itoa(opts.Page))
		}
		if opts.Limit > 0 {
			params.Add("perPage", strconv.Itoa(opts.Limit))
		}
		if opts.Period != "" {
			params.Add("timeRange", string(opts.Period))
		}
		if opts.Search != "" {
			params.Add("search", opts.Search)
		}
	}
	return executeRequest[InventoryPayload](ctx, c, "GET", "user/getSalesHistory", nil, params)
}

func (c *Client) PurchaseHistory(ctx context.Context, opts *HistoryOptions) (InventoryPayload, error) {
	params := url.Values{}
	if opts != nil {
		if opts.Page > 0 {
			params.Add("page", strconv.Itoa(opts.Page))
		}
		if opts.Limit > 0 {
			params.Add("count", strconv.Itoa(opts.Limit))
		}
	}
	return executeRequest[InventoryPayload](ctx, c, "GET", "user/getPurchaseHistory", nil, params)
}

func (c *Client) ItemPricing(ctx context.Context, itemID int) (PriceItem, error) {
	return executeRequest[PriceItem](ctx, c, "GET", "item/pricing/"+strconv.Itoa(itemID), nil, nil)
}

func (c *Client) ItemPricingBulk(ctx context.Context, itemIDs []int) (BulkPricingPayload, error) {
	if len(itemIDs) > 100 {
		return BulkPricingPayload{}, fmt.Errorf("the pricing bulk endpoint has a limit of 100 ids")
	}
	idStrings := make([]string, len(itemIDs))
	for i, id := range itemIDs {
		idStrings[i] = strconv.Itoa(id)
	}
	params := url.Values{}
	params.Add("items", strings.Join(idStrings, ","))
	return executeRequest[BulkPricingPayload](ctx, c, "GET", "item/pricing/bulk", nil, params)
}

func (c *Client) CreateBuyOrder(ctx context.Context, itemID, value, amount int) error {
	payload := buyOrderRequest{
		ItemID: itemID,
		Value:  value,
		Amount: amount,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error encoding json for creating buy order: %w", err)
	}
	_, err = executeRequest[json.RawMessage](ctx, c, "POST", "item/buyorder", jsonData, nil)
	return err
}

func (c *Client) ItemSalesGraph(ctx context.Context, itemID int, period Period) (PriceHistoryPayload, error) {
	if period == "" {
		period = Period1Month
	}
	params := url.Values{}
	params.Add("period", string(period))
	return executeRequest[PriceHistoryPayload](ctx, c, "GET", "item/salesGraph/"+strconv.Itoa(itemID), nil, params)
}

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

	return executeRequest[ListingPayload](ctx, c, "GET", endpoint, nil, params)
}

func (c *Client) BuyOrderList(ctx context.Context, itemID int) (BuyOrderPayload, error) {
	return executeRequest[BuyOrderPayload](ctx, c, "GET", "item/buyorderList/"+strconv.Itoa(itemID), nil, nil)
}

func (c *Client) SetBaseURL(url string) {
	c.baseURL = url
}

func (c *Client) GetBaseURL() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.baseURL

}
