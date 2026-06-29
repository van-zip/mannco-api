package mannco

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestItemPricing(t *testing.T) {
	runAPITest(t, testCase[PriceItem]{
		name:           "ItemPricing_success",
		mockStatus:     200,
		mockResponse:   `{"err":false,"success":true,"message":"","content":{"from_cache":true,"item_id":440,"last_updated":"2024-01-01T00:00:00Z","pricing":{"last_sale":{"date":1704067200,"price":5000},"lowest_buy_order":4500,"lowest_sale_price":5000,"steam_price":6000,"suggested_price":5200}}}`,
		expectedPath:   "/item/pricing/440",
		expectedMethod: "GET",
		runTest: func(ctx context.Context, client *Client) (PriceItem, error) {
			return client.ItemPricing(ctx, 440)
		},
		assertResponse: func(t *testing.T, res PriceItem) {
			if res.ItemID != 440 {
				t.Errorf("expected ItemID 440, got %d", res.ItemID)
			}
			if res.Pricing.LowestBuyOrder != 4500 {
				t.Errorf("expected LowestBuyOrder 4500, got %d", res.Pricing.LowestBuyOrder)
			}
		},
	})

	runAPITest(t, testCase[PriceItem]{
		name:           "ItemPricing_not_found",
		mockStatus:     404,
		mockResponse:   `{"err":true,"success":false,"message":"Item not found","content":null}`,
		expectedPath:   "/item/pricing/999",
		expectedMethod: "GET",
		runTest: func(ctx context.Context, client *Client) (PriceItem, error) {
			return client.ItemPricing(ctx, 999)
		},
		assertError: func(t *testing.T, err error) {
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		},
	})
}

func TestItemPricingBulk(t *testing.T) {
	runAPITest(t, testCase[BulkPricingPayload]{
		name:           "ItemPricingBulk_success",
		mockStatus:     200,
		mockResponse:   `{"err":false,"success":true,"message":"","content":{"cached_items":1,"items":[{"from_cache":true,"item_id":440,"last_updated":"2024-01-01T00:00:00Z","pricing":{"last_sale":{"date":1704067200,"price":5000},"lowest_buy_order":4500,"lowest_sale_price":5000,"steam_price":6000,"suggested_price":5200}}],"refreshed_items":1,"total_items":2}}`,
		expectedPath:   "/item/pricing/bulk",
		expectedMethod: "GET",
		runTest: func(ctx context.Context, client *Client) (BulkPricingPayload, error) {
			return client.ItemPricingBulk(ctx, []int{440, 5021})
		},
		assertResponse: func(t *testing.T, res BulkPricingPayload) {
			if len(res.Items) != 1 {
				t.Errorf("expected 1 item, got %d", len(res.Items))
			}
		},
	})

	runAPITest(t, testCase[BulkPricingPayload]{
		name:           "ItemPricingBulk_too_many_ids",
		mockStatus:     200,
		mockResponse:   "",
		expectedPath:   "",
		expectedMethod: "",
		runTest: func(ctx context.Context, client *Client) (BulkPricingPayload, error) {
			ids := make([]int, 101)
			for i := range ids {
				ids[i] = i
			}
			return client.ItemPricingBulk(ctx, ids)
		},
		assertError: func(t *testing.T, err error) {
			if err == nil {
				t.Fatal("expected error for too many IDs, got nil")
			}
		},
	})
}

func TestItemSalesGraph(t *testing.T) {
	runAPITest(t, testCase[PriceHistoryPayload]{
		name:           "ItemSalesGraph_success",
		mockStatus:     200,
		mockResponse:   `{"err":false,"success":true,"message":"","content":{"values":[{"date":"2024-01-01","price":5000,"nb":10},{"date":"2024-01-02","price":5100,"nb":5}]}}`,
		expectedPath:   "/item/salesGraph/440",
		expectedMethod: "GET",
		runTest: func(ctx context.Context, client *Client) (PriceHistoryPayload, error) {
			return client.ItemSalesGraph(ctx, 440, Period1Month)
		},
		assertResponse: func(t *testing.T, res PriceHistoryPayload) {
			if len(res.Values) != 2 {
				t.Errorf("expected 2 values, got %d", len(res.Values))
			}
		},
	})
}

func TestItemListings(t *testing.T) {
	runAPITest(t, testCase[ListingPayload]{
		name:           "ItemListings_success",
		mockStatus:     200,
		mockResponse:   `{"err":false,"success":true,"message":"","content":{"count":1,"values":[{"assetId":"12345","bot":"bot1","game":440,"getImage":"","html":"","id":1,"item_id":440,"killstreaker":"","paint":"","parts":"","price":5000,"sheen":"","spell":"","state":1,"user":"user1","wear":0.1}]}}`,
		expectedPath:   "/item/listing/440",
		expectedMethod: "GET",
		runTest: func(ctx context.Context, client *Client) (ListingPayload, error) {
			return client.ItemListings(ctx, 440, "", nil)
		},
		assertResponse: func(t *testing.T, res ListingPayload) {
			if res.Count != 1 {
				t.Errorf("expected count 1, got %d", res.Count)
			}
		},
	})

	runAPITest(t, testCase[ListingPayload]{
		name:           "ItemListings_with_user_id",
		mockStatus:     200,
		mockResponse:   `{"err":false,"success":true,"message":"","content":{"count":1,"values":[]}}`,
		expectedPath:   "/item/listing/440/user123",
		expectedMethod: "GET",
		runTest: func(ctx context.Context, client *Client) (ListingPayload, error) {
			return client.ItemListings(ctx, 440, "user123", nil)
		},
		assertResponse: func(_ *testing.T, _ ListingPayload) {},
	})

	runAPITest(t, testCase[ListingPayload]{
		name:           "ItemListings_with_options",
		mockStatus:     200,
		mockResponse:   `{"err":false,"success":true,"message":"","content":{"count":1,"values":[]}}`,
		expectedPath:   "/item/listing/440",
		expectedMethod: "GET",
		runTest: func(ctx context.Context, client *Client) (ListingPayload, error) {
			return client.ItemListings(ctx, 440, "", &ListingOptions{Count: 10, Page: 2, Game: 440})
		},
		assertResponse: func(_ *testing.T, _ ListingPayload) {},
	})
}

func TestItemSalesGraphErrorPaths(t *testing.T) {
	// Test server error (500)
	runAPITest(t, testCase[PriceHistoryPayload]{
		name:           "ItemSalesGraph_server_error",
		mockStatus:     500,
		mockResponse:   `{"err":true,"success":false,"message":"Internal server error","content":null}`,
		expectedPath:   "/item/salesGraph/440",
		expectedMethod: "GET",
		runTest: func(ctx context.Context, client *Client) (PriceHistoryPayload, error) {
			return client.ItemSalesGraph(ctx, 440, Period1Month)
		},
		assertError: func(t *testing.T, err error) {
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			var apiErr *APIError
			if !errors.As(err, &apiErr) {
				t.Errorf("expected *APIError, got %T: %v", err, err)
			}
			if apiErr.StatusCode != http.StatusInternalServerError {
				t.Errorf("expected status 500, got %d", apiErr.StatusCode)
			}
		},
	})

	// Test not found (404)
	runAPITest(t, testCase[PriceHistoryPayload]{
		name:           "ItemSalesGraph_not_found",
		mockStatus:     404,
		mockResponse:   `{"err":true,"success":false,"message":"Item not found","content":null}`,
		expectedPath:   "/item/salesGraph/999",
		expectedMethod: "GET",
		runTest: func(ctx context.Context, client *Client) (PriceHistoryPayload, error) {
			return client.ItemSalesGraph(ctx, 999, Period1Month)
		},
		assertError: func(t *testing.T, err error) {
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			var apiErr *APIError
			if !errors.As(err, &apiErr) {
				t.Errorf("expected *APIError, got %T: %v", err, err)
			}
			if apiErr.StatusCode != http.StatusNotFound {
				t.Errorf("expected status 404, got %d", apiErr.StatusCode)
			}
		},
	})

	// Test unauthorized (401)
	runAPITest(t, testCase[PriceHistoryPayload]{
		name:           "ItemSalesGraph_unauthorized",
		mockStatus:     401,
		mockResponse:   `{"err":true,"success":false,"message":"Unauthorized","content":null}`,
		expectedPath:   "/item/salesGraph/440",
		expectedMethod: "GET",
		runTest: func(ctx context.Context, client *Client) (PriceHistoryPayload, error) {
			return client.ItemSalesGraph(ctx, 440, Period1Month)
		},
		assertError: func(t *testing.T, err error) {
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, ErrUnauthorized) {
				t.Errorf("expected ErrUnauthorized, got %v", err)
			}
		},
	})

	// Test malformed JSON response
	runAPITest(t, testCase[PriceHistoryPayload]{
		name:           "ItemSalesGraph_malformed_response",
		mockStatus:     200,
		mockResponse:   `not valid json`,
		expectedPath:   "/item/salesGraph/440",
		expectedMethod: "GET",
		runTest: func(ctx context.Context, client *Client) (PriceHistoryPayload, error) {
			return client.ItemSalesGraph(ctx, 440, Period1Month)
		},
		assertError: func(t *testing.T, err error) {
			if err == nil {
				t.Fatal("expected JSON decode error, got nil")
			}
			if !strings.Contains(err.Error(), "failed decoding response JSON") {
				t.Errorf("expected JSON decode error, got %v", err)
			}
		},
	})

	// Test API response with err=true
	runAPITest(t, testCase[PriceHistoryPayload]{
		name:           "ItemSalesGraph_api_err_true",
		mockStatus:     200,
		mockResponse:   `{"err":true,"success":false,"message":"API error","content":null}`,
		expectedPath:   "/item/salesGraph/440",
		expectedMethod: "GET",
		runTest: func(ctx context.Context, client *Client) (PriceHistoryPayload, error) {
			return client.ItemSalesGraph(ctx, 440, Period1Month)
		},
		assertError: func(t *testing.T, err error) {
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), "API error") {
				t.Errorf("expected API error message, got %v", err)
			}
		},
	})

	// Test API response with success=false
	runAPITest(t, testCase[PriceHistoryPayload]{
		name:           "ItemSalesGraph_api_success_false",
		mockStatus:     200,
		mockResponse:   `{"err":false,"success":false,"message":"Operation failed","content":null}`,
		expectedPath:   "/item/salesGraph/440",
		expectedMethod: "GET",
		runTest: func(ctx context.Context, client *Client) (PriceHistoryPayload, error) {
			return client.ItemSalesGraph(ctx, 440, Period1Month)
		},
		assertError: func(t *testing.T, err error) {
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), "Operation failed") {
				t.Errorf("expected operation failed message, got %v", err)
			}
		},
	})
}

func TestItemSalesGraphDefaultPeriod(t *testing.T) {
	// Test that empty period defaults to Period1Month
	runAPITest(t, testCase[PriceHistoryPayload]{
		name:           "ItemSalesGraph_default_period",
		mockStatus:     200,
		mockResponse:   `{"err":false,"success":true,"message":"","content":{"values":[]}}`,
		expectedPath:   "/item/salesGraph/440",
		expectedMethod: "GET",
		runTest: func(ctx context.Context, client *Client) (PriceHistoryPayload, error) {
			// Pass empty period to test default
			return client.ItemSalesGraph(ctx, 440, "")
		},
		assertResponse: func(_ *testing.T, _ PriceHistoryPayload) {},
	})
}

