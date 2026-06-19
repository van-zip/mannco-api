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
	"time"
)

const BaseUrl = "https://api.mannco.store/"

var httpClient = &http.Client{
	Timeout: 60 * time.Second,
}

type ApiResponse[T any] struct {
	Err     bool   `json:"err"`
	Success bool   `json:"success"`
	Message string `json:"message"`
	Content T      `json:"content"`
}

type PriceHistoryPoint struct {
	Date  string `json:"date"`
	Price int    `json:"price"`
	Nb    int    `json:"nb"` // Number of sales (volume)
}

type PriceHistoryPayload struct {
	Values []PriceHistoryPoint `json:"values"`
}

type Period string

// all valid time frames
const (
	Period1Month  Period = "1M"
	Period3Months Period = "3M"
	Period6Months Period = "6M"
	Period1Year   Period = "1Y"
	Period5Years  Period = "5Y"
	PeriodAll     Period = "ALL"
)

type LoginPayload struct {
	JWT string `json:"jwt"`
}

type BalancePayload struct {
	Balance int `json:"balance"` // Returned in pennies
}

type Item struct {
	Count        int    `json:"count"`
	Date         int64  `json:"date"` // Unix timestamp
	Effect       string `json:"effect"`
	Festivized   int    `json:"festivized"`
	IDBackpack   int    `json:"idbackpack"`
	IDItem       int    `json:"iditem"`
	Image        string `json:"image"` // The image asset hash
	Inspect      string `json:"inspect"`
	Killstreaker string `json:"killstreaker"`
	Level        int    `json:"level"`
	Name         string `json:"name"`
	Paint        string `json:"paint"`
	Parts        string `json:"parts"`
	Price        int    `json:"price"` // Price in pennies
	Sheen        string `json:"sheen"`
	Spell        string `json:"spell"`
	URL          string `json:"url"`
}

type LastSale struct {
	Date  int64 `json:"date"`  // Unix timestamp
	Price int   `json:"price"` // Price in pennies
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
	BuyOrders map[string]BuyOrderInfo `json:"informations"`
}

type buyOrderRequest struct {
	ItemID int `json:"itemid"`
	Value  int `json:"value"`
	Amount int `json:"amount"`
}

func executeRequest[T any](ctx context.Context, method, endpoint, jwt string, body []byte, queryParams url.Values) (T, error) {
	var target T

	u, err := url.Parse(BaseUrl + endpoint)
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
	if jwt != "" {
		req.Header.Set("Authorization", "Bearer "+jwt)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return target, fmt.Errorf("network call execution failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return target, fmt.Errorf("server rejected request with status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return target, fmt.Errorf("failed reading raw response bytes: %w", err)
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

func UserLogin(ctx context.Context, apiKey string) (string, error) {
	data := map[string]string{"apiKey": apiKey}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("error encoding json for user login: %w", err)
	}

	content, err := executeRequest[LoginPayload](ctx, "POST", "user/login", "", jsonData, nil)
	if err != nil {
		return "", err
	}
	return content.JWT, nil
}

func Balance(ctx context.Context, jwt string) (int, error) {
	content, err := executeRequest[BalancePayload](ctx, "GET", "user/balance", jwt, nil, nil)
	if err != nil {
		return -1, err
	}
	return content.Balance, nil
}

func TransactionHistory(ctx context.Context, jwt string, page, limit int) (InventoryPayload, error) {
	params := url.Values{}
	if page > 0 {
		params.Add("page", strconv.Itoa(page))
	}
	if limit > 0 {
		params.Add("limit", strconv.Itoa(limit))
	}
	return executeRequest[InventoryPayload](ctx, "GET", "user/getTransactionHistory", jwt, nil, params)
}

func SalesHistory(ctx context.Context, jwt string, page, perPage int, period Period, search string) (InventoryPayload, error) {
	params := url.Values{}
	if page > 0 {
		params.Add("page", strconv.Itoa(page))
	}
	if perPage > 0 {
		params.Add("perPage", strconv.Itoa(perPage))
	}
	if period != "" {
		params.Add("timeRange", string(period))
	}
	if search != "" {
		params.Add("search", search)
	}
	return executeRequest[InventoryPayload](ctx, "GET", "user/getSalesHistory", jwt, nil, params)
}

func PurchaseHistory(ctx context.Context, jwt string, page, count int) (InventoryPayload, error) {
	params := url.Values{}
	if page > 0 {
		params.Add("page", strconv.Itoa(page))
	}
	if count > 0 {
		params.Add("count", strconv.Itoa(count))
	}
	return executeRequest[InventoryPayload](ctx, "GET", "user/getPurchaseHistory", jwt, nil, params)
}

func ItemPricing(ctx context.Context, jwt string, id int) (PriceItem, error) {
	return executeRequest[PriceItem](ctx, "GET", "item/pricing/"+strconv.Itoa(id), jwt, nil, nil)
}

func ItemPricingBulk(ctx context.Context, jwt string, ids []int) (BulkPricingPayload, error) {
	if len(ids) > 100 {
		return BulkPricingPayload{}, fmt.Errorf("the pricing bulk endpoint has a limit of 100 ids")
	}
	idStrings := make([]string, len(ids))
	for i, id := range ids {
		idStrings[i] = strconv.Itoa(id)
	}
	params := url.Values{}
	params.Add("items", strings.Join(idStrings, ","))
	return executeRequest[BulkPricingPayload](ctx, "GET", "item/pricing/bulk", jwt, nil, params)
}

func CreateBuyOrder(ctx context.Context, jwt string, itemid, value, amount int) error {
	payload := buyOrderRequest{
		ItemID: itemid,
		Value:  value,
		Amount: amount,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error encoding json for creating buy order: %w", err)
	}
	_, err = executeRequest[json.RawMessage](ctx, "POST", "item/buyorder", jwt, jsonData, nil)
	return err
}

func ItemSalesGraph(ctx context.Context, jwt string, itemid int, period Period) (PriceHistoryPayload, error) {
	if period == "" {
		period = Period1Month
	}
	params := url.Values{}
	params.Add("period", string(period))
	return executeRequest[PriceHistoryPayload](ctx, "GET", "item/salesGraph/"+strconv.Itoa(itemid), jwt, nil, params)
}

func ItemListings(ctx context.Context, jwt string, itemid int, userid string, count, page, game int) (ListingPayload, error) {
	endpoint := "item/listing/" + strconv.Itoa(itemid)
	if userid != "" {
		endpoint += "/" + userid
	}

	params := url.Values{}
	if count > 0 {
		params.Add("count", strconv.Itoa(count))
	}
	if page > 0 {
		params.Add("page", strconv.Itoa(page))
	}
	if game > 0 {
		params.Add("game", strconv.Itoa(game))
	}

	return executeRequest[ListingPayload](ctx, "GET", endpoint, jwt, nil, params)
}

func BuyOrderList(ctx context.Context, jwt string, itemid int) (BuyOrderPayload, error) {
	return executeRequest[BuyOrderPayload](ctx, "GET", "item/buyorderList/"+strconv.Itoa(itemid), jwt, nil, nil)
}
