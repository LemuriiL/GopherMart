package accrual

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	baseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{baseURL: baseURL}
}

type Response struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}

func (c *Client) GetOrder(number string) (*Response, int, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.baseURL, number)

	resp, err := http.Get(url)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil, resp.StatusCode, nil
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, resp.StatusCode, nil
	}

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, resp.StatusCode, err
	}

	return &result, resp.StatusCode, nil
}
