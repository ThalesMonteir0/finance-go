package verboo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gofinance/dto/ia"
	"io"
	"net/http"
	"time"
)

const baseURL = "https://code.verboo.ai/router/v1"

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func New(apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *Client) ChatCompletion(ctx context.Context, req ia.ChatCompletionRequest) (*ia.ChatCompletionResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("verboo: status inesperado %d: %s", resp.StatusCode, string(respBody))
	}

	var chatResp ia.ChatCompletionResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, err
	}

	return &chatResp, nil
}
