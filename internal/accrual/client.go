// Package accrual - клиент для работы с сервисом начислений.
package accrual

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// Client - простой HTTP клиент для обращения к системе начислений.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// Response - ответ от сервиса начислений по заказу.
type Response struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}

// NewClient - создаёт новый клиент с базовым URL.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GetOrder - запрашивает статус и начисление по номеру заказа.
// Возвращает ответ, HTTP статус, retryAfter (если 429) и ошибку.
func (c *Client) GetOrder(number string) (*Response, int, int, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.baseURL, number)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil, resp.StatusCode, 0, nil
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := 60
		if value := resp.Header.Get("Retry-After"); value != "" {
			if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
				retryAfter = parsed
			}
		}
		return nil, resp.StatusCode, retryAfter, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, 0, nil
	}

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, resp.StatusCode, 0, err
	}

	return &result, resp.StatusCode, 0, nil
}
