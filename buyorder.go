package mannco

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"
)

// API envelope key constants
const (
	envelopeKeyInformations = "informations"
	envelopeKeyMore         = "more"
	timestampKey            = "timestamp"
)

// UserItemBuyOrderPayload is the payload returned by UserItemBuyOrder
type UserItemBuyOrderPayload struct {
	ID        int       `json:"id"`
	Price     int       `json:"price"`
	Amount    int       `json:"amount"`
	Timestamp time.Time `json:"timestamp"`
}

// ErrNoInformations is returned when the API response is missing the 'informations' envelope
var ErrNoInformations = errors.New("missing 'informations' envelope in API response")

// unwrapEnvelope extracts the raw JSON for the given key from an API response envelope.
// Returns the raw JSON message.
// If the envelope key exists but is null/empty/empty-object, returns empty RawMessage with no error (valid empty response).
// If the envelope key is missing entirely, returns ErrNoInformations.
func unwrapEnvelope(data []byte, key string) (json.RawMessage, error) {
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, err
	}

	raw, ok := envelope[key]
	if !ok {
		// Key is missing entirely - this is an invalid/unexpected API response
		return nil, ErrNoInformations
	}

	// Key exists but is null, empty, or empty object - valid empty response
	if len(raw) == 0 || string(raw) == "null" || string(raw) == "{}" {
		return json.RawMessage(""), nil
	}

	return raw, nil
}

// UnmarshalJSON handles converting the API response into readable responses
func (u *UserItemBuyOrderPayload) UnmarshalJSON(data []byte) error {
	// Define the inner structure we expect inside 'informations'
	type infoDTO struct {
		ID        int    `json:"id"`
		Price     int    `json:"price"`
		Amount    int    `json:"amount"`
		Timestamp string `json:"timestamp"`
	}

	raw, err := unwrapEnvelope(data, envelopeKeyInformations)
	if err != nil {
		return err
	}

	var dto infoDTO
	if err := json.Unmarshal(raw, &dto); err != nil {
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
	raw, err := unwrapEnvelope(data, envelopeKeyInformations)
	if err != nil {
		// If there's no envelope, try direct unmarshal (for backward compatibility / tests)
		// but only if the error is specifically about missing envelope
		if errors.Is(err, ErrNoInformations) {
			// Allow empty array or null at top level for tests
			if len(bytes.TrimSpace(data)) == 0 || string(bytes.TrimSpace(data)) == "null" {
				return nil
			}
			// Try unmarshaling directly as array
			var arr []BuyOrderInfo
			if json.Unmarshal(data, &arr) == nil {
				b.BuyOrders = arr
				return nil
			}
		}
		return err
	}

	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || string(trimmed) == "null" {
		// No buy orders
		return nil
	}

	// The upstream API either returns a JSON array or a map with numeric keys + optional "more"
	// Handle the array case
	if trimmed[0] == '[' {
		var arr []BuyOrderInfo
		if err := json.Unmarshal(trimmed, &arr); err != nil {
			return err
		}
		b.BuyOrders = arr
		return nil
	}

	// Handle the map case
	var obj map[string]BuyOrderInfo
	if err := json.Unmarshal(trimmed, &obj); err != nil {
		return err
	}

	var more *BuyOrderInfo
	keys := make([]int, 0, len(obj))
	for k := range obj {
		if k == envelopeKeyMore {
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