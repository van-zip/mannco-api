package mannco

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"
)

// unwrapEnvelope extracts the raw JSON for the given key from an API response envelope. Returns the raw JSON message.
func unwrapEnvelope(data []byte, key string) (json.RawMessage, error) {
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, err
	}

	raw, ok := envelope[key]
	if !ok {
		// key is missing
		return nil, fmt.Errorf("%w: missing '%s' envelope in API response", ErrInternal, key)
	}

	// key exists but is null / empty
	if len(raw) == 0 || string(raw) == "null" || string(raw) == "{}" {
		return json.RawMessage(""), nil
	}

	return raw, nil
}

// BulkBuyOrderEntry represents a single buy order entry in a bulk request
type BulkBuyOrderEntry struct {
	ItemID int `json:"itemid"`
	Value  int `json:"value"`
	Amount int `json:"amount"`
}

// BulkBuyOrderResult represents the result of a single entry in a bulk buy order request
type BulkBuyOrderResult struct {
	ItemID  int    `json:"itemid"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// BulkBuyOrdersContent represents the content of the bulk buy orders response
type BulkBuyOrdersContent struct {
	Total     int                  `json:"total"`
	Processed int                  `json:"processed"`
	Errors    int                  `json:"errors"`
	Results   []BulkBuyOrderResult `json:"results"`
}

// UserBuyOrdersPayload is the payload returned by GetUserBuyOrders
type UserBuyOrdersPayload struct {
	Values []UserBuyOrderItem `json:"values"`
	Count  BuyOrderCount      `json:"count"`
}

// UserBuyOrderItem represents a buy order in the user's buy orders list
type UserBuyOrderItem struct {
	ID     int    `json:"id"`
	ItemID int    `json:"itemid"`
	Price  int    `json:"price"`
	Amount int    `json:"amount"`
	Name   string `json:"name"`
	Game   int    `json:"game"`
}

// BuyOrderCount contains the total count of buy orders
type BuyOrderCount struct {
	Count int `json:"nb"`
}

// UserItemBuyOrderPayload is the payload returned by UserItemBuyOrder
type UserItemBuyOrderPayload struct {
	ID        int       `json:"id"`
	Price     int       `json:"price"`
	Amount    int       `json:"amount"`
	Timestamp time.Time `json:"timestamp"`
}

// UnmarshalJSON handles converting the API response into readable responses
func (u *UserItemBuyOrderPayload) UnmarshalJSON(data []byte) error {
	var dto struct {
		ID        int    `json:"id"`
		Price     int    `json:"price"`
		Amount    int    `json:"amount"`
		Timestamp string `json:"timestamp"`
	}
	raw, err := unwrapEnvelope(data, "informations")

	if err != nil {
		return fmt.Errorf("unwrapEnvelope on json response: %w", err)
	}

	if err = json.Unmarshal(raw, &dto); err != nil {
		return err
	}

	u.ID = dto.ID
	u.Price = dto.Price
	u.Amount = dto.Amount

	if dto.Timestamp != "" {
		t, err := strconv.ParseInt(dto.Timestamp, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid timestamp format %q: %w", dto.Timestamp, err)
		}
		u.Timestamp = time.Unix(t, 0)
	}

	return nil
}

// BuyOrderInfo is a specific buy order
type BuyOrderInfo struct {
	Count int `json:"count"`
	Price int `json:"price"`
}

// buyOrderRequest represents to build a buy order request
type buyOrderRequest struct {
	ItemID int `json:"itemid"`
	Value  int `json:"value"`
	Amount int `json:"amount"`
}

// BuyOrderPayload is the returned payload from the API
type BuyOrderPayload struct {
	BuyOrders []BuyOrderInfo
}

// UnmarshalJSON handles the inconsistently shaped responses from the upstream API
func (b *BuyOrderPayload) UnmarshalJSON(data []byte) error {
	raw, err := unwrapEnvelope(data, "informations")
	if err != nil {
		return fmt.Errorf("unwrapEnvelope on json response: %w", err)
	}

	if len(raw) == 0 || string(raw) == "null" {
		// No buy orders
		return nil
	}

	// The upstream API either returns a JSON map using 'tiers' or just an array

	// Handle the array case
	if raw[0] == '[' {
		var arr []BuyOrderInfo
		if err := json.Unmarshal(raw, &arr); err != nil {
			return err
		}
		b.BuyOrders = arr
		return nil
	}

	// Handle the map case
	var obj map[string]BuyOrderInfo
	if err := json.Unmarshal(raw, &obj); err != nil {
		return err
	}

	var more *BuyOrderInfo
	keys := make([]int, 0, len(obj))
	for k := range obj {
		if k == "more" {
			v := obj[k]
			more = &v
			continue
		}
		n, err := strconv.Atoi(k)
		if err != nil {
			return fmt.Errorf("unexpected non-numeric, non-'more' key %q in informations object", k)
		}
		keys = append(keys, n)
	}
	sort.Ints(keys)

	result := make([]BuyOrderInfo, 0, len(keys)+1)
	for _, k := range keys {
		result = append(result, obj[strconv.Itoa(k)])
	}
	if more != nil {
		result = append(result, *more)
	}
	b.BuyOrders = result
	return nil
}

// BuyOrderList gets the active buy orders for an item ID
func (c *Client) BuyOrderList(ctx context.Context, itemID int) (BuyOrderPayload, error) {
	return executeRequest[BuyOrderPayload](ctx, c, "GET", "item/buyorderList/"+strconv.Itoa(itemID), nil, nil)
}

// CreateBuyOrder creates a buy order for a specific item ID at a given price
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

// UpdateBuyOrder updates an existing buy order for an item with a new price and/or quantity.
func (c *Client) UpdateBuyOrder(ctx context.Context, itemID, value, amount int) error {
	payload := buyOrderRequest{
		ItemID: itemID,
		Value:  value,
		Amount: amount,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error encoding json for updating buy order: %w", err)
	}
	_, err = executeRequest[json.RawMessage](ctx, c, "POST", "item/buyorder/update", jsonData, nil)
	return err
}

// RemoveBuyOrder cancels and removes a buy order for an item, releasing the reserved balance.
func (c *Client) RemoveBuyOrder(ctx context.Context, itemID int) error {
	jsonData, err := json.Marshal(map[string]int{"itemid": itemID})
	if err != nil {
		return fmt.Errorf("error encoding json for removing buy order: %w", err)
	}
	_, err = executeRequest[json.RawMessage](ctx, c, "POST", "item/buyorder/remove", jsonData, nil)
	return err
}

// BulkBuyOrders processes up to 100 buy orders in a single request.
// For each entry, a new buy order is inserted, an existing one is updated,
// or it is removed when both value and amount are 0.
func (c *Client) BulkBuyOrders(ctx context.Context, orders []BulkBuyOrderEntry) (BulkBuyOrdersContent, error) {
	if len(orders) > MaxBulkItems {
		return BulkBuyOrdersContent{}, fmt.Errorf("%w: maximum 100 orders allowed in bulk request", ErrInternal)
	}
	payload := struct {
		Orders []BulkBuyOrderEntry `json:"orders"`
	}{
		Orders: orders,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return BulkBuyOrdersContent{}, fmt.Errorf("error encoding json for bulk buy orders: %w", err)
	}
	return executeRequest[BulkBuyOrdersContent](ctx, c, "POST", "item/buyorder/bulk", jsonData, nil)
}

// UserItemBuyOrder gets the logged in user's buy orders for a specific item id
func (c *Client) UserItemBuyOrder(ctx context.Context, itemID int) (UserItemBuyOrderPayload, error) {
	return executeRequest[UserItemBuyOrderPayload](ctx, c, "GET", "user/buyorder/"+strconv.Itoa(itemID), nil, nil)
}

// GetUserBuyOrders gets all the logged in user's buy orders
func (c *Client) GetUserBuyOrders(ctx context.Context) (UserBuyOrdersPayload, error) {
	return executeRequest[UserBuyOrdersPayload](ctx, c, "GET", "user/getBuyorder", nil, nil)
}
