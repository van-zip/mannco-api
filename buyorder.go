package mannco

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
)

// BuyOrderInfo is a specific buy order
type BuyOrderInfo struct {
	Count int `json:"count"`
	Price int `json:"price"`
}

// BuyOrderPayload is the returned payload from the API
type BuyOrderPayload struct {
	BuyOrders []BuyOrderInfo
}

// UnmarshalJSON handles the inconsistently shaped responses from the upstream API
func (b *BuyOrderPayload) UnmarshalJSON(data []byte) error {
	// Unwraps to just the Informations entry of the json
	var envelope struct {
		Informations json.RawMessage `json:"informations"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return err
	}

	raw := bytes.TrimSpace(envelope.Informations)
	if len(raw) == 0 {
		// No buy orders
		return nil
	}

	// The upstream API either returns a JSON map using 'tiers' or just an array
	// Handle the array
	if raw[0] == '[' {
		var arr []BuyOrderInfo
		if err := json.Unmarshal(raw, &arr); err != nil {
			return err
		}
		b.BuyOrders = arr
		return nil
	}

	// Handle the map
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

// buyOrderRequest represents to build a buy order request
type buyOrderRequest struct {
	ItemID int `json:"itemid"`
	Value  int `json:"value"`
	Amount int `json:"amount"`
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
