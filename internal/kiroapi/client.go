// Package kiroapi implements native Kiro API auth — bypassing the kiro-cli subprocess.
// Status: work in progress. Requires traffic capture via mitmproxy to finalize
// the ksk_ → JWT exchange flow against runtime.us-east-1.kiro.dev.
package kiroapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	RuntimeEndpoint = "https://runtime.us-east-1.kiro.dev"
	MessagesPath    = "/v1/messages"
)

// Client talks directly to the Kiro native API.
type Client struct {
	httpClient *http.Client
	apiKey     string // ksk_ key
}

func New(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// Complete sends a messages request and returns the full response body.
// The caller is responsible for parsing streaming vs. non-streaming.
func (c *Client) Complete(reqBody []byte, stream bool) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, RuntimeEndpoint+MessagesPath, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	if stream {
		req.Header.Set("Accept", "text/event-stream")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("kiro api request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("kiro api status %d: %s", resp.StatusCode, body)
	}

	return body, nil
}

// TODO: implement JWT exchange
// Flow (captured via mitmproxy — to be completed):
//  1. POST /auth/token  { "api_key": "ksk_..." }  → { "token": "<jwt>" }
//  2. Use JWT in Authorization: Bearer <jwt> for subsequent requests

type authRequest struct {
	APIKey string `json:"api_key"`
}

type authResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

func (c *Client) exchangeToken() (*authResponse, error) {
	payload, _ := json.Marshal(authRequest{APIKey: c.apiKey})
	req, err := http.NewRequest(http.MethodPost, RuntimeEndpoint+"/auth/token", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var ar authResponse
	if err := json.NewDecoder(resp.Body).Decode(&ar); err != nil {
		return nil, err
	}
	return &ar, nil
}
