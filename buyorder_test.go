package mannco

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestBuyOrderList(t *testing.T) {
	runAPITest(t, testCase[BuyOrderPayload]{
		name:       "BuyOrderList_success_array",
		mockStatus: 200,
		mockResponse: `{
  "err": false,
  "success": true,
  "content": {
    "informations": [
      {
        "count": 1,
        "price": 191200
      },
      {
        "count": 1,
        "price": 191000
      }
    ]
  }
}`,
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
		name:       "BuyOrderList_success_object_with_more",
		mockStatus: 200,
		mockResponse: `{
  "err": false,
  "success": true,
  "content": {
    "informations": {
      "0": {
        "count": 1,
        "price": 4153
      },
      "1": {
        "count": 5,
        "price": 4150
      },
      "2": {
        "count": 2,
        "price": 4000
      },
      "3": {
        "count": 1,
        "price": 3720
      },
      "4": {
        "count": 1,
        "price": 3687
      },
      "more": {
        "count": 125,
        "price": 3524
      }
    }
  }
}`,
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

