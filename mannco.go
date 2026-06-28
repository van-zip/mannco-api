// Package mannco is an API client for mannco.store
package mannco

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// BaseURL is the default base URL for the mannco.store api
const BaseURL = "https://api.mannco.store/"

/*
	Datatypes
*/
// APIResponse is the general shape of Mannco.store API responses
type APIResponse[T any] struct {
	Err     bool   `json:"err"`
	Success bool   `json:"success"`
	Message string `json:"message"`
	Content T      `json:"content"`
}

// APIError represents the general shape of API errors
type APIError struct {
	StatusCode int
	Message    string
}

// Error formats APIErrors
func (e *APIError) Error() string {
	return fmt.Sprintf("server error with status code %d: %s", e.StatusCode, e.Message)
}

var (
	// ErrUnauthorized indicates either insufficient permissions or expired authentication
	ErrUnauthorized = errors.New("authentication error")
	// ErrNetwork is any network related failure
	ErrNetwork = errors.New("network error")
	// ErrInternal is any error derived from internal package logic
	ErrInternal = errors.New("internal error")
)

// Client represents the base API client which interacts with Mannco.store
type Client struct {
	httpClient *http.Client
	mu         sync.RWMutex
	baseURL    string
	jwt        string
}

// NewClient instantiates a new API client
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

// SetJWT sets the JWT for a client
func (c *Client) SetJWT(token string) {
	c.mu.Lock()
	c.jwt = token
	c.mu.Unlock()
}

// GetJWT gets the JWT for a client
func (c *Client) GetJWT() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.jwt
}

// SetBaseURL sets the API client base url
func (c *Client) SetBaseURL(url string) {
	c.mu.Lock()
	c.baseURL = url
	c.mu.Unlock()
}

// GetBaseURL gets the API client base url
func (c *Client) GetBaseURL() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.baseURL

}

// executeRequest performs generic parsing, safety handling, and raw IO operations for interactinng with the API
func ExecuteRequest[T any](ctx context.Context, c *Client, method, endpoint string, body []byte, queryParams url.Values) (T, error) {
	var target T

	u, err := url.Parse(c.GetBaseURL() + endpoint)
	if err != nil {
		return target, fmt.Errorf("%w: invalid endpoint url: %w", ErrInternal, err)
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
		return target, fmt.Errorf("%w: failed to construct request object: %w", ErrInternal, err)
	}

	req.Header.Set("Content-Type", "application/json")
	if jwt := c.GetJWT(); jwt != "" {
		req.Header.Set("Authorization", "Bearer "+jwt)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return target, fmt.Errorf("%w: request execution failed: %w", ErrNetwork, err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return target, fmt.Errorf("%w: failed reading raw response bytes: %w", ErrNetwork, &APIError{StatusCode: resp.StatusCode, Message: err.Error()})
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr APIResponse[json.RawMessage]
		msg := ""
		if json.Unmarshal(bodyBytes, &apiErr) == nil {
			msg = apiErr.Message
		}
		httpErr := &APIError{StatusCode: resp.StatusCode, Message: msg}
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return target, fmt.Errorf("%w: %w", ErrUnauthorized, httpErr)
		}
		return target, httpErr
	}
	var apiResponse APIResponse[T]
	if err = json.Unmarshal(bodyBytes, &apiResponse); err != nil {
		return target, fmt.Errorf("%w: failed decoding response JSON: %w", ErrInternal, &APIError{StatusCode: resp.StatusCode, Message: err.Error()})
	}

	if apiResponse.Err || !apiResponse.Success {
		return target, &APIError{StatusCode: resp.StatusCode, Message: apiResponse.Message}
	}
	return apiResponse.Content, nil
}
