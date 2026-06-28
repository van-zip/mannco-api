package mannco

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// UserLogin uses the provided API key to login and return a JWT
func (c *Client) UserLogin(ctx context.Context, apiKey string) (string, error) {
	data := map[string]string{"apiKey": apiKey}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("error encoding json for user login: %w", err)
	}

	content, err := ExecuteRequest[LoginPayload](ctx, c, "POST", "user/login", jsonData, nil)
	if err != nil {
		return "", err
	}

	c.SetJWT(content.JWT)

	return content.JWT, nil
}

// Balance returns the user balance in pennies
func (c *Client) Balance(ctx context.Context) (int, error) {
	content, err := ExecuteRequest[BalancePayload](ctx, c, "GET", "user/balance", nil, nil)
	if err != nil {
		return 0, err
	}
	return content.Balance, nil
}

// TransactionHistory returns the user transaction history on the site
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
	return ExecuteRequest[InventoryPayload](ctx, c, "GET", "user/getTransactionHistory", nil, params)
}

// SalesHistory returns the user sales history on the site
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
	return ExecuteRequest[InventoryPayload](ctx, c, "GET", "user/getSalesHistory", nil, params)
}

// PurchaseHistory returns the user purchase history on the site
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
	return ExecuteRequest[InventoryPayload](ctx, c, "GET", "user/getPurchaseHistory", nil, params)
}
