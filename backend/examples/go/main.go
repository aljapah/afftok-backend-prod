// AffTok Server-to-Server Integration - Go Example
//
// This example shows how to send postbacks/conversions from your server
// to AffTok using the Server-to-Server API.
package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Config holds the SDK configuration
type Config struct {
	APIKey       string
	AdvertiserID string
	BaseURL      string
	Timeout      time.Duration
}

// ConversionParams represents conversion/postback parameters
type ConversionParams struct {
	OfferID       string                 `json:"offer_id"`
	TransactionID string                 `json:"transaction_id"`
	ClickID       string                 `json:"click_id,omitempty"`
	Amount        *float64               `json:"amount,omitempty"`
	Currency      string                 `json:"currency,omitempty"`
	Status        string                 `json:"status,omitempty"`
	CustomParams  map[string]interface{} `json:"custom_params,omitempty"`
}

// ClickParams represents click tracking parameters
type ClickParams struct {
	OfferID      string                 `json:"offer_id"`
	TrackingCode string                 `json:"tracking_code,omitempty"`
	SubID1       string                 `json:"sub_id_1,omitempty"`
	SubID2       string                 `json:"sub_id_2,omitempty"`
	SubID3       string                 `json:"sub_id_3,omitempty"`
	IP           string                 `json:"ip,omitempty"`
	UserAgent    string                 `json:"user_agent,omitempty"`
	CustomParams map[string]interface{} `json:"custom_params,omitempty"`
}

// Result represents the API response
type Result struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

// AfftokTracker handles communication with AffTok API
type AfftokTracker struct {
	config Config
	client *http.Client
}

// NewAfftokTracker creates a new tracker instance
func NewAfftokTracker(config Config) *AfftokTracker {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.afftok.com"
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &AfftokTracker{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// generateSignature creates HMAC-SHA256 signature
func (t *AfftokTracker) generateSignature(timestamp int64, nonce string) string {
	dataToSign := fmt.Sprintf("%s|%s|%d|%s", t.config.APIKey, t.config.AdvertiserID, timestamp, nonce)
	h := hmac.New(sha256.New, []byte(t.config.APIKey))
	h.Write([]byte(dataToSign))
	return hex.EncodeToString(h.Sum(nil))
}

// generateNonce creates a random nonce
func (t *AfftokTracker) generateNonce(length int) string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	rand.Read(b)
	for i := range b {
		b[i] = chars[int(b[i])%len(chars)]
	}
	return string(b)
}

// SendPostback sends a conversion/postback to AffTok
func (t *AfftokTracker) SendPostback(params ConversionParams) Result {
	timestamp := time.Now().UnixMilli()
	nonce := t.generateNonce(32)
	signature := t.generateSignature(timestamp, nonce)

	payload := map[string]interface{}{
		"api_key":        t.config.APIKey,
		"advertiser_id":  t.config.AdvertiserID,
		"offer_id":       params.OfferID,
		"transaction_id": params.TransactionID,
		"status":         "approved",
		"currency":       "USD",
		"timestamp":      timestamp,
		"nonce":          nonce,
		"signature":      signature,
	}

	if params.Status != "" {
		payload["status"] = params.Status
	}
	if params.Currency != "" {
		payload["currency"] = params.Currency
	}
	if params.ClickID != "" {
		payload["click_id"] = params.ClickID
	}
	if params.Amount != nil {
		payload["amount"] = *params.Amount
	}
	if params.CustomParams != nil {
		payload["custom_params"] = params.CustomParams
	}

	return t.sendRequest("/api/postback", payload)
}

// TrackClick tracks a click event server-side
func (t *AfftokTracker) TrackClick(params ClickParams) Result {
	timestamp := time.Now().UnixMilli()
	nonce := t.generateNonce(32)
	signature := t.generateSignature(timestamp, nonce)

	payload := map[string]interface{}{
		"api_key":       t.config.APIKey,
		"advertiser_id": t.config.AdvertiserID,
		"offer_id":      params.OfferID,
		"timestamp":     timestamp,
		"nonce":         nonce,
		"signature":     signature,
	}

	if params.TrackingCode != "" {
		payload["tracking_code"] = params.TrackingCode
	}
	if params.SubID1 != "" {
		payload["sub_id_1"] = params.SubID1
	}
	if params.SubID2 != "" {
		payload["sub_id_2"] = params.SubID2
	}
	if params.SubID3 != "" {
		payload["sub_id_3"] = params.SubID3
	}
	if params.IP != "" {
		payload["ip"] = params.IP
	}
	if params.UserAgent != "" {
		payload["user_agent"] = params.UserAgent
	}
	if params.CustomParams != nil {
		payload["custom_params"] = params.CustomParams
	}

	return t.sendRequest("/api/sdk/click", payload)
}

// SendBatchPostbacks sends multiple conversions
func (t *AfftokTracker) SendBatchPostbacks(conversions []ConversionParams) []Result {
	results := make([]Result, len(conversions))

	for i, conversion := range conversions {
		results[i] = t.SendPostback(conversion)
		time.Sleep(100 * time.Millisecond) // Rate limit protection
	}

	return results
}

// sendRequest sends HTTP POST request
func (t *AfftokTracker) sendRequest(endpoint string, payload map[string]interface{}) Result {
	url := t.config.BaseURL + endpoint

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return Result{Success: false, Error: err.Error()}
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return Result{Success: false, Error: err.Error()}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", t.config.APIKey)

	resp, err := t.client.Do(req)
	if err != nil {
		return Result{Success: false, Error: err.Error()}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{Success: false, Error: err.Error()}
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		return Result{Success: true, Data: data}
	}

	return Result{Success: false, Error: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body))}
}

func main() {
	fmt.Println("AffTok Server-to-Server Integration Example (Go)\n")

	// Configuration
	apiKey := os.Getenv("AFFTOK_API_KEY")
	if apiKey == "" {
		apiKey = "your_api_key"
	}
	advertiserID := os.Getenv("AFFTOK_ADVERTISER_ID")
	if advertiserID == "" {
		advertiserID = "your_advertiser_id"
	}

	tracker := NewAfftokTracker(Config{
		APIKey:       apiKey,
		AdvertiserID: advertiserID,
	})

	// Example 1: Send a simple conversion
	fmt.Println("1. Sending a simple conversion...")
	amount := 29.99
	result := tracker.SendPostback(ConversionParams{
		OfferID:       "offer_123",
		TransactionID: fmt.Sprintf("txn_%d", time.Now().Unix()),
		Amount:        &amount,
		Status:        "approved",
	})
	fmt.Printf("Result: %+v\n\n", result)

	// Example 2: Send a conversion with click attribution
	fmt.Println("2. Sending a conversion with click attribution...")
	amount2 := 49.99
	result = tracker.SendPostback(ConversionParams{
		OfferID:       "offer_123",
		TransactionID: fmt.Sprintf("txn_%d_2", time.Now().Unix()),
		ClickID:       "click_abc123",
		Amount:        &amount2,
		Currency:      "EUR",
		Status:        "approved",
		CustomParams: map[string]interface{}{
			"product_id": "prod_456",
			"category":   "electronics",
		},
	})
	fmt.Printf("Result: %+v\n\n", result)

	// Example 3: Track a server-side click
	fmt.Println("3. Tracking a server-side click...")
	result = tracker.TrackClick(ClickParams{
		OfferID:      "offer_123",
		TrackingCode: "campaign_summer_2024",
		SubID1:       "source_google",
		IP:           "192.168.1.1",
		UserAgent:    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
	})
	fmt.Printf("Result: %+v\n\n", result)

	// Example 4: Batch send conversions
	fmt.Println("4. Batch sending conversions...")
	amount10 := 10.00
	amount20 := 20.00
	amount30 := 30.00
	batchResults := tracker.SendBatchPostbacks([]ConversionParams{
		{OfferID: "offer_123", TransactionID: fmt.Sprintf("batch_1_%d", time.Now().Unix()), Amount: &amount10},
		{OfferID: "offer_123", TransactionID: fmt.Sprintf("batch_2_%d", time.Now().Unix()), Amount: &amount20},
		{OfferID: "offer_123", TransactionID: fmt.Sprintf("batch_3_%d", time.Now().Unix()), Amount: &amount30},
	})
	fmt.Printf("Batch results: %+v\n", batchResults)
}

