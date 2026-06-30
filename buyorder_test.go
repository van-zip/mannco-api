package mannco

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestBuyOrderList(t *testing.T) {
	runAPITest(t, testCase[BuyOrderPayload]{
		name:           "BuyOrderList_success_array",
		mockStatus:     200,
		mockResponse:   `{"err":false,"success":true,"content":{"informations":[{"count":1,"price":191200},{"count":1,"price":191000}]}}`,
		expectedPath:   "/item/buyorderList/174119",
		expectedMethod: "GET",
		runTest: func(ctx context.Context, client *Client) (BuyOrderPayload, error) {
			return client.BuyOrderList(ctx, 174119)
		},
		assertResponse: func(t *testing.T, res BuyOrderPayload) {
			if len(res.BuyOrders) != 2 {
				t.Errorf("expected 2 buy orders, got %d", len(res.BuyOrders))
			}
			if res.BuyOrders[0].Count != 1 || res.BuyOrders[0].Price != 191200 {
				t.Errorf("expected first order count=1 price=191200, got count=%d price=%d", res.BuyOrders[0].Count, res.BuyOrders[0].Price)
			}
		},
	})

	runAPITest(t, testCase[BuyOrderPayload]{
		name:           "BuyOrderList_success_object_with_more",
		mockStatus:     200,
		mockResponse:   `{"err":false,"success":true,"content":{"informations":{"0":{"count":1,"price":4153},"1":{"count":5,"price":4150},"2":{"count":2,"price":4000},"3":{"count":1,"price":3720},"4":{"count":1,"price":3687},"more":{"count":125,"price":3524}}}}`,
		expectedPath:   "/item/buyorderList/371",
		expectedMethod: "GET",
		runTest: func(ctx context.Context, client *Client) (BuyOrderPayload, error) {
			return client.BuyOrderList(ctx, 371)
		},
		assertResponse: func(t *testing.T, res BuyOrderPayload) {
			if len(res.BuyOrders) != 6 {
				t.Errorf("expected 6 buy orders (including more), got %d", len(res.BuyOrders))
			}
		},
	})

	runAPITest(t, testCase[BuyOrderPayload]{
		name:           "BuyOrderList_empty_null",
		mockStatus:     200,
		mockResponse:   `{"err":false,"success":true,"message":"","content":{"informations":null}}`,
		expectedPath:   "/item/buyorderList/958",
		expectedMethod: "GET",
		runTest: func(ctx context.Context, client *Client) (BuyOrderPayload, error) {
			return client.BuyOrderList(ctx, 958)
		},
		assertResponse: func(t *testing.T, res BuyOrderPayload) {
			if res.BuyOrders != nil {
				t.Errorf("expected nil BuyOrders for null response, got %v", res.BuyOrders)
			}
		},
	})

	runAPITest(t, testCase[BuyOrderPayload]{
		name:           "BuyOrderList_empty_object",
		mockStatus:     200,
		mockResponse:   `{"err":false,"success":true,"message":"","content":{"informations":{}}}`,
		expectedPath:   "/item/buyorderList/958",
		expectedMethod: "GET",
		runTest: func(ctx context.Context, client *Client) (BuyOrderPayload, error) {
			return client.BuyOrderList(ctx, 958)
		},
		assertResponse: func(t *testing.T, res BuyOrderPayload) {
			if len(res.BuyOrders) != 0 {
				t.Errorf("expected empty BuyOrders for empty object, got %d", len(res.BuyOrders))
			}
		},
	})

	runAPITest(t, testCase[BuyOrderPayload]{
		name:           "BuyOrderList_not_found",
		mockStatus:     404,
		mockResponse:   `{"err":true,"success":false,"message":"Item not found","content":null}`,
		expectedPath:   "/item/buyorderList/999",
		expectedMethod: "GET",
		runTest: func(ctx context.Context, client *Client) (BuyOrderPayload, error) {
			return client.BuyOrderList(ctx, 999)
		},
		assertError: func(t *testing.T, err error) {
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, ErrInternal) {
				t.Errorf("expected internal error for 404, got: %v", err)
			}
		},
	})
}

func TestCreateBuyOrder(t *testing.T) {
	runAPITest(t, testCase[json.RawMessage]{
		name:           "CreateBuyOrder_success",
		mockStatus:     200,
		mockResponse:   `{"err":false,"success":true,"message":"Buy order created","content":{}}`,
		expectedPath:   "/item/buyorder",
		expectedMethod: "POST",
		runTest: func(ctx context.Context, client *Client) (json.RawMessage, error) {
			return nil, client.CreateBuyOrder(ctx, 958, 1000, 5)
		},
		assertResponse: func(_ *testing.T, _ json.RawMessage) {},
	})

	runAPITest(t, testCase[json.RawMessage]{
		name:           "CreateBuyOrder_unauthorized",
		mockStatus:     403,
		mockResponse:   `{"err":true,"success":false,"message":"Invalid API key","content":null}`,
		expectedPath:   "/item/buyorder",
		expectedMethod: "POST",
		runTest: func(ctx context.Context, client *Client) (json.RawMessage, error) {
			return nil, client.CreateBuyOrder(ctx, 958, 1000, 5)
		},
		assertError: func(t *testing.T, err error) {
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, ErrUnauthorized) {
				t.Errorf("expected unauthorized error, got: %v", err)
			}
		},
	})
}

func TestCreateBuyOrderErrorPaths(t *testing.T) {
	// Test server error (500)
	runAPITest(t, testCase[json.RawMessage]{
		name:           "CreateBuyOrder_server_error",
		mockStatus:     500,
		mockResponse:   `{"err":true,"success":false,"message":"Internal server error","content":null}`,
		expectedPath:   "/item/buyorder",
		expectedMethod: "POST",
		runTest: func(ctx context.Context, client *Client) (json.RawMessage, error) {
			return nil, client.CreateBuyOrder(ctx, 958, 1000, 5)
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
	runAPITest(t, testCase[json.RawMessage]{
		name:           "CreateBuyOrder_not_found",
		mockStatus:     404,
		mockResponse:   `{"err":true,"success":false,"message":"Item not found","content":null}`,
		expectedPath:   "/item/buyorder",
		expectedMethod: "POST",
		runTest: func(ctx context.Context, client *Client) (json.RawMessage, error) {
			return nil, client.CreateBuyOrder(ctx, 999999, 1000, 5)
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

	// Test malformed JSON response
	runAPITest(t, testCase[json.RawMessage]{
		name:           "CreateBuyOrder_malformed_response",
		mockStatus:     200,
		mockResponse:   `not valid json`,
		expectedPath:   "/item/buyorder",
		expectedMethod: "POST",
		runTest: func(ctx context.Context, client *Client) (json.RawMessage, error) {
			return nil, client.CreateBuyOrder(ctx, 958, 1000, 5)
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
}

func TestBuyOrderPayloadUnmarshalJSON(t *testing.T) {
	t.Run("empty_null_informations", func(t *testing.T) {
		var payload BuyOrderPayload
		err := json.Unmarshal([]byte(`{"informations":null}`), &payload)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if payload.BuyOrders != nil {
			t.Errorf("expected nil BuyOrders, got %v", payload.BuyOrders)
		}
	})

	t.Run("empty_object_informations", func(t *testing.T) {
		var payload BuyOrderPayload
		err := json.Unmarshal([]byte(`{"informations":{}}`), &payload)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(payload.BuyOrders) != 0 {
			t.Errorf("expected empty BuyOrders, got %d", len(payload.BuyOrders))
		}
	})

	t.Run("empty_string_informations", func(t *testing.T) {
		var payload BuyOrderPayload
		// Empty string for informations is not a valid type, but we test the error path
		err := json.Unmarshal([]byte(`{"informations":""}`), &payload)
		if err == nil {
			t.Fatal("expected error for string informations, got nil")
		}
		// Should fail because string cannot unmarshal into map
	})

	t.Run("invalid_json_envelope", func(t *testing.T) {
		var payload BuyOrderPayload
		err := json.Unmarshal([]byte(`not valid json`), &payload)
		if err == nil {
			t.Fatal("expected error for invalid JSON, got nil")
		}
	})

	t.Run("array_with_invalid_element", func(t *testing.T) {
		var payload BuyOrderPayload
		// Test with an element that has wrong type for count (string instead of int)
		err := json.Unmarshal([]byte(`{"informations":[{"count":"not_int","price":100}]}`), &payload)
		if err == nil {
			t.Fatal("expected error for invalid count type, got nil")
		}
	})

	t.Run("map_with_non_numeric_key", func(t *testing.T) {
		var payload BuyOrderPayload
		err := json.Unmarshal([]byte(`{"informations":{"invalid_key":{"count":1,"price":100}}}`), &payload)
		if err == nil {
			t.Fatal("expected error for non-numeric key, got nil")
		}
		if !strings.Contains(err.Error(), "unexpected non-numeric") {
			t.Errorf("expected non-numeric key error, got %v", err)
		}
	})

	t.Run("map_unmarshal_error", func(t *testing.T) {
		var payload BuyOrderPayload
		// This has a valid structure but invalid type for price (string instead of int)
		err := json.Unmarshal([]byte(`{"informations":{"0":{"count":1,"price":"not_a_number"}}}`), &payload)
		if err == nil {
			t.Fatal("expected error for invalid price type, got nil")
		}
	})
}

func TestUserItemBuyOrder(t *testing.T) {
	runAPITest(t, testCase[UserItemBuyOrderPayload]{
		name:           "UserItemBuyOrder_success",
		mockStatus:     200,
		mockResponse:   `{"err":false,"success":true,"content":{"informations":{"id":98765,"steamid":"76561198000000000","itemid":12345,"price":15000,"amount":3,"timestamp":"1706745600"}}}`,
		expectedPath:   "/user/buyorder/98765",
		expectedMethod: "GET",
		runTest: func(ctx context.Context, client *Client) (UserItemBuyOrderPayload, error) {
			return client.UserItemBuyOrder(ctx, 98765)
		},
		assertResponse: func(t *testing.T, res UserItemBuyOrderPayload) {
			if res.ID != 98765 {
				t.Errorf("expected id 98765, got %d", res.ID)
			}
			if res.Price != 15000 {
				t.Errorf("expected price 15000, got %d", res.Price)
			}
			if res.Amount != 3 {
				t.Errorf("expected amount 3, got %d", res.Amount)
			}
			if res.Timestamp != time.Unix(1706745600, 0) {
				t.Errorf("expected timestamp 1706745600, got %v", res.Timestamp)
			}
		},
	})
}

func TestGetUserBuyOrders(t *testing.T) {
	runAPITest(t, testCase[UserBuyOrdersPayload]{
		name:           "GetUserBuyOrders_success",
		mockStatus:     200,
		mockResponse:   `{"err":false,"success":true,"content":{"values":[{"id":98765,"itemid":12345,"price":15000,"amount":3,"name":"Burning Flames Team Captain","game":440}],"count":{"nb":15}}}`,
		expectedPath:   "/user/getBuyorder",
		expectedMethod: "GET",
		runTest: func(ctx context.Context, client *Client) (UserBuyOrdersPayload, error) {
			return client.GetUserBuyOrders(ctx)
		},
		assertResponse: func(t *testing.T, res UserBuyOrdersPayload) {
			if len(res.Values) != 1 {
				t.Errorf("expected 1 buy order, got %d", len(res.Values))
			}
			if res.Values[0].ID != 98765 {
				t.Errorf("expected id 98765, got %d", res.Values[0].ID)
			}
			if res.Values[0].ItemID != 12345 {
				t.Errorf("expected itemid 12345, got %d", res.Values[0].ItemID)
			}
			if res.Values[0].Price != 15000 {
				t.Errorf("expected price 15000, got %d", res.Values[0].Price)
			}
			if res.Values[0].Amount != 3 {
				t.Errorf("expected amount 3, got %d", res.Values[0].Amount)
			}
			if res.Values[0].Name != "Burning Flames Team Captain" {
				t.Errorf("expected name 'Burning Flames Team Captain', got %q", res.Values[0].Name)
			}
			if res.Values[0].Game != 440 {
				t.Errorf("expected game 440, got %d", res.Values[0].Game)
			}
			if res.Count.Count != 15 {
				t.Errorf("expected count 15, got %d", res.Count.Count)
			}
		},
	})
}

func TestUpdateBuyOrder(t *testing.T) {
	runAPITest(t, testCase[json.RawMessage]{
		name:           "UpdateBuyOrder_success",
		mockStatus:     200,
		mockResponse:   `{"err":false,"success":true,"message":"Buy order updated","content":{}}`,
		expectedPath:   "/item/buyorder/update",
		expectedMethod: "POST",
		runTest: func(ctx context.Context, client *Client) (json.RawMessage, error) {
			return nil, client.UpdateBuyOrder(ctx, 12345, 16000, 5)
		},
		assertResponse: func(_ *testing.T, _ json.RawMessage) {},
	})

	runAPITest(t, testCase[json.RawMessage]{
		name:           "UpdateBuyOrder_unauthorized",
		mockStatus:     403,
		mockResponse:   `{"err":true,"success":false,"message":"Invalid API key","content":null}`,
		expectedPath:   "/item/buyorder/update",
		expectedMethod: "POST",
		runTest: func(ctx context.Context, client *Client) (json.RawMessage, error) {
			return nil, client.UpdateBuyOrder(ctx, 12345, 16000, 5)
		},
		assertError: func(t *testing.T, err error) {
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, ErrUnauthorized) {
				t.Errorf("expected unauthorized error, got: %v", err)
			}
		},
	})

	runAPITest(t, testCase[json.RawMessage]{
		name:           "UpdateBuyOrder_business_error",
		mockStatus:     300,
		mockResponse:   `{"err":true,"success":false,"message":"You don't have a buy order for this item","content":null}`,
		expectedPath:   "/item/buyorder/update",
		expectedMethod: "POST",
		runTest: func(ctx context.Context, client *Client) (json.RawMessage, error) {
			return nil, client.UpdateBuyOrder(ctx, 12345, 16000, 5)
		},
		assertError: func(t *testing.T, err error) {
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), "You don't have a buy order") {
				t.Errorf("expected business error about missing buy order, got: %v", err)
			}
		},
	})
}

func TestRemoveBuyOrder(t *testing.T) {
	runAPITest(t, testCase[json.RawMessage]{
		name:           "RemoveBuyOrder_success",
		mockStatus:     200,
		mockResponse:   `{"err":false,"success":true,"message":"Removed","content":{}}`,
		expectedPath:   "/item/buyorder/remove",
		expectedMethod: "POST",
		runTest: func(ctx context.Context, client *Client) (json.RawMessage, error) {
			return nil, client.RemoveBuyOrder(ctx, 12345)
		},
		assertResponse: func(_ *testing.T, _ json.RawMessage) {},
	})

	runAPITest(t, testCase[json.RawMessage]{
		name:           "RemoveBuyOrder_unauthorized",
		mockStatus:     403,
		mockResponse:   `{"err":true,"success":false,"message":"Invalid API key","content":null}`,
		expectedPath:   "/item/buyorder/remove",
		expectedMethod: "POST",
		runTest: func(ctx context.Context, client *Client) (json.RawMessage, error) {
			return nil, client.RemoveBuyOrder(ctx, 12345)
		},
		assertError: func(t *testing.T, err error) {
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, ErrUnauthorized) {
				t.Errorf("expected unauthorized error, got: %v", err)
			}
		},
	})

	runAPITest(t, testCase[json.RawMessage]{
		name:           "RemoveBuyOrder_not_found",
		mockStatus:     300,
		mockResponse:   `{"err":true,"success":false,"message":"No buy order found","content":null}`,
		expectedPath:   "/item/buyorder/remove",
		expectedMethod: "POST",
		runTest: func(ctx context.Context, client *Client) (json.RawMessage, error) {
			return nil, client.RemoveBuyOrder(ctx, 12345)
		},
		assertError: func(t *testing.T, err error) {
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), "No buy order found") {
				t.Errorf("expected 'No buy order found' error, got: %v", err)
			}
		},
	})
}

func TestBulkBuyOrders(t *testing.T) {
	runAPITest(t, testCase[BulkBuyOrdersContent]{
		name:           "BulkBuyOrders_success",
		mockStatus:     200,
		mockResponse:   `{"err":false,"success":true,"content":{"total":3,"processed":2,"errors":1,"results":[{"itemid":123,"status":"inserted","message":"Inserted"},{"itemid":456,"status":"updated","message":"Buy order updated"},{"itemid":321,"status":"error","message":"No buy order found"}]}}`,
		expectedPath:   "/item/buyorder/bulk",
		expectedMethod: "POST",
		runTest: func(ctx context.Context, client *Client) (BulkBuyOrdersContent, error) {
			return client.BulkBuyOrders(ctx, []BulkBuyOrderEntry{
				{ItemID: 123, Value: 50, Amount: 2},
				{ItemID: 456, Value: 10, Amount: 5},
				{ItemID: 321, Value: 0, Amount: 0},
			})
		},
		assertResponse: func(t *testing.T, res BulkBuyOrdersContent) {
			if res.Total != 3 {
				t.Errorf("expected total=3, got %d", res.Total)
			}
			if res.Processed != 2 {
				t.Errorf("expected processed=2, got %d", res.Processed)
			}
			if res.Errors != 1 {
				t.Errorf("expected errors=1, got %d", res.Errors)
			}
			if len(res.Results) != 3 {
				t.Errorf("expected 3 results, got %d", len(res.Results))
			}
			if res.Results[0].ItemID != 123 || res.Results[0].Status != "inserted" {
				t.Errorf("expected first result itemid=123 status=inserted, got itemid=%d status=%s", res.Results[0].ItemID, res.Results[0].Status)
			}
		},
	})

	runAPITest(t, testCase[BulkBuyOrdersContent]{
		name:           "BulkBuyOrders_unauthorized",
		mockStatus:     403,
		mockResponse:   `{"err":true,"success":false,"message":"Invalid API key","content":null}`,
		expectedPath:   "/item/buyorder/bulk",
		expectedMethod: "POST",
		runTest: func(ctx context.Context, client *Client) (BulkBuyOrdersContent, error) {
			return client.BulkBuyOrders(ctx, []BulkBuyOrderEntry{
				{ItemID: 123, Value: 50, Amount: 2},
			})
		},
		assertError: func(t *testing.T, err error) {
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, ErrUnauthorized) {
				t.Errorf("expected unauthorized error, got: %v", err)
			}
		},
	})

	runAPITest(t, testCase[BulkBuyOrdersContent]{
		name:           "BulkBuyOrders_rate_limited",
		mockStatus:     429,
		mockResponse:   `{"err":true,"success":false,"message":"Rate limit exceeded. Retry after 60 seconds."}`,
		expectedPath:   "/item/buyorder/bulk",
		expectedMethod: "POST",
		runTest: func(ctx context.Context, client *Client) (BulkBuyOrdersContent, error) {
			return client.BulkBuyOrders(ctx, []BulkBuyOrderEntry{
				{ItemID: 123, Value: 50, Amount: 2},
			})
		},
		assertError: func(t *testing.T, err error) {
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), "Rate limit exceeded") {
				t.Errorf("expected rate limit error, got: %v", err)
			}
		},
	})

	runAPITest(t, testCase[BulkBuyOrdersContent]{
		name:           "BulkBuyOrders_too_many_orders",
		mockStatus:     200,
		mockResponse:   `{"err":false,"success":true,"content":{}}`,
		expectedPath:   "/item/buyorder/bulk",
		expectedMethod: "POST",
		runTest: func(ctx context.Context, client *Client) (BulkBuyOrdersContent, error) {
			orders := make([]BulkBuyOrderEntry, 101)
			for i := range orders {
				orders[i] = BulkBuyOrderEntry{ItemID: i + 1, Value: 10, Amount: 1}
			}
			return client.BulkBuyOrders(ctx, orders)
		},
		assertError: func(t *testing.T, err error) {
			if err == nil {
				t.Fatal("expected error for >100 orders, got nil")
			}
			if !strings.Contains(err.Error(), "maximum 100 orders allowed") {
				t.Errorf("expected max orders error, got: %v", err)
			}
		},
	})
}
