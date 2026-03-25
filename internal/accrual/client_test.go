package accrual

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientGetOrderNoContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL)

	resp, statusCode, retryAfter, err := client.GetOrder("123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp != nil {
		t.Fatal("expected nil response")
	}

	if statusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", statusCode)
	}

	if retryAfter != 0 {
		t.Fatalf("expected retryAfter 0, got %d", retryAfter)
	}
}

func TestClientGetOrderTooManyRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "15")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewClient(server.URL)

	resp, statusCode, retryAfter, err := client.GetOrder("123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp != nil {
		t.Fatal("expected nil response")
	}

	if statusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", statusCode)
	}

	if retryAfter != 15 {
		t.Fatalf("expected retryAfter 15, got %d", retryAfter)
	}
}

func TestClientGetOrderTooManyRequestsDefaultRetryAfter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewClient(server.URL)

	resp, statusCode, retryAfter, err := client.GetOrder("123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp != nil {
		t.Fatal("expected nil response")
	}

	if statusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", statusCode)
	}

	if retryAfter != 60 {
		t.Fatalf("expected retryAfter 60, got %d", retryAfter)
	}
}

func TestClientGetOrderTooManyRequestsInvalidRetryAfter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "abc")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewClient(server.URL)

	resp, statusCode, retryAfter, err := client.GetOrder("123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp != nil {
		t.Fatal("expected nil response")
	}

	if statusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", statusCode)
	}

	if retryAfter != 60 {
		t.Fatalf("expected retryAfter 60, got %d", retryAfter)
	}
}

func TestClientGetOrderOK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"order":"123","status":"PROCESSED","accrual":500}`)
	}))
	defer server.Close()

	client := NewClient(server.URL)

	resp, statusCode, retryAfter, err := client.GetOrder("123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if statusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", statusCode)
	}

	if retryAfter != 0 {
		t.Fatalf("expected retryAfter 0, got %d", retryAfter)
	}

	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	if resp.Order != "123" {
		t.Fatalf("expected order 123, got %s", resp.Order)
	}

	if resp.Status != "PROCESSED" {
		t.Fatalf("expected status PROCESSED, got %s", resp.Status)
	}

	if resp.Accrual == nil || *resp.Accrual != 500 {
		t.Fatalf("expected accrual 500, got %v", resp.Accrual)
	}
}

func TestClientGetOrderOKWithoutAccrual(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"order":"123","status":"PROCESSING"}`)
	}))
	defer server.Close()

	client := NewClient(server.URL)

	resp, statusCode, retryAfter, err := client.GetOrder("123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if statusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", statusCode)
	}

	if retryAfter != 0 {
		t.Fatalf("expected retryAfter 0, got %d", retryAfter)
	}

	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	if resp.Order != "123" {
		t.Fatalf("expected order 123, got %s", resp.Order)
	}

	if resp.Status != "PROCESSING" {
		t.Fatalf("expected status PROCESSING, got %s", resp.Status)
	}

	if resp.Accrual != nil {
		t.Fatalf("expected nil accrual, got %v", resp.Accrual)
	}
}

func TestClientGetOrderBadJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{"order":`)
	}))
	defer server.Close()

	client := NewClient(server.URL)

	_, statusCode, _, err := client.GetOrder("123")
	if err == nil {
		t.Fatal("expected error")
	}

	if statusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", statusCode)
	}
}

func TestClientGetOrderInternalServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL)

	resp, statusCode, retryAfter, err := client.GetOrder("123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp != nil {
		t.Fatal("expected nil response")
	}

	if statusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", statusCode)
	}

	if retryAfter != 0 {
		t.Fatalf("expected retryAfter 0, got %d", retryAfter)
	}
}
