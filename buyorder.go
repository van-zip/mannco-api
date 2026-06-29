package mannco

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"
)

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
		Informations *struct {
			ID        int    `json:"id"`
			Price     int    `json:"price"`
			Amount    int    `json:"amount"`
			Timestamp string `json:"timestamp"`
		} `json:"informations"`
	}

	if err := json.Unmarshal(data, &dto); err != nil {
		return err
	}

	if dto.Informations == nil {
		return nil
	}

	u.ID = dto.Informations.ID
	u.Price = dto.Informations.Price
	u.Amount = dto.Informations.Amount

	if dto.Informations.Timestamp != "" {
		t, err := strconv.ParseInt(dto.Informations.Timestamp, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid timestamp format %q: %w", dto.Informations.Timestamp, err)
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
	// Unwraps to just the Informations entry of the json
	var envelope struct {
		Informations json.RawMessage `json:"informations"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return err
	}

	raw := bytes.TrimSpace(envelope.Informations)
	if len(raw) == 0 || string(raw) == "null" {
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

// UserItemBuyOrder gets the logged in user's buy orders for a specific item id
func (c *Client) UserItemBuyOrder(ctx context.Context, itemID int) (UserItemBuyOrderPayload, error) {
	return executeRequest[UserItemBuyOrderPayload](ctx, c, "GET", "user/buyorder/"+strconv.Itoa(itemID), nil, nil)
}
