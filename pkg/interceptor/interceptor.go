package interceptor

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/MaplesMcDepth/canopy/pkg/models"
)

// Interceptor wraps an http.RoundTripper to track AI API calls and costs.
type Interceptor struct {
	Store      models.DBStore
	CostCalculator CostCalculatorFunc
	RoundTripper http.RoundTripper
}

// CostCalculatorFunc calculates the cost of an API call based on request and response.
type CostCalculatorFunc func(*http.Request, *http.Response) (float64, error)

// RoundTrip executes a single HTTP transaction, capturing metrics for AI API calls.
func (i *Interceptor) RoundTrip(req *http.Request) (*http.Response, error) {
	// Execute the request
	resp, err := i.RoundTripper.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	// Only process successful responses for cost calculation
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Calculate cost using the provided calculator
		cost, err := i.CostCalculator(req, resp)
		if err != nil {
			// Log error but don't break the request
			// In a real implementation, you might want to log this
			return resp, nil
		}

		// Extract tokens if possible (for OpenAI-compatible APIs)
		var tokensInput, tokensOutput int
		if resp.Body != nil {
			// Read response body to parse token usage
			bodyBytes, err := io.ReadAll(resp.Body)
			resp.Body.Close() // must close
			if err == nil {
				// Restore body for caller
				resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				// Try to parse as JSON for token usage
				var tokenResp struct {
					Usage struct {
						PromptTokens     int `json:"prompt_tokens"`
						CompletionTokens int `json:"completion_tokens"`
					} `json:"usage"`
				}
				if err := json.Unmarshal(bodyBytes, &tokenResp); err == nil {
					tokensInput = tokenResp.Usage.PromptTokens
					tokensOutput = tokenResp.Usage.CompletionTokens
				}
			} else {
				// If we can't read body, create a new reader for the caller
				resp.Body = io.NopCloser(bytes.NewBuffer(nil))
			}
		}

		// Create API call record
		call := &models.APICall{
			Timestamp:   time.Now(),
			APIKey:      extractAPIKey(req),
			Endpoint:    req.URL.String(),
			Model:       extractModel(req),
			TokensInput: tokensInput,
			TokensOutput: tokensOutput,
			Cost:        cost,
		}

		// Store the call (non-blocking in real implementation)
		go func() {
			if err := i.Store.CreateAPICall(call); err != nil {
				// Log error
			}
		}()
	}

	return resp, nil
}

// extractAPIKey attempts to extract API key from request headers.
// Returns a masked version for privacy.
func extractAPIKey(req *http.Request) string {
	// Check Authorization header
	if auth := req.Header.Get("Authorization"); auth != "" {
		// Simple masking: show first 4 and last 4 chars
		if len(auth) > 8 {
			return auth[:4] + "..." + auth[len(auth)-4:]
		}
		return auth
	}
	// Check for API key in custom headers
	if key := req.Header.Get("X-API-Key"); key != "" {
		if len(key) > 8 {
			return key[:4] + "..." + key[len(key)-4:]
		}
		return key
	}
	return "unknown"
}

// extractModel attempts to extract model from request body or URL.
// This is a simplified implementation.
func extractModel(req *http.Request) string {
	// For OpenAI API, model is often in JSON body
	// We'll return a placeholder; in practice, parse the body
	return "unknown"
}