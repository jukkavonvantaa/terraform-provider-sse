// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apiclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Scopes
const (
	ScopePolicies    = "policies"
	ScopeReports     = "reports"
	ScopeAdmin       = "admin"
	ScopeDeployments = "deployments"
)

// Operations
const (
	OperationGet    = "GET"
	OperationPost   = "POST"
	OperationPut    = "PUT"
	OperationPatch  = "PATCH"
	OperationDelete = "DELETE"
)

// Token represents an OAuth2 token response
type Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	IssuedAt    time.Time
}

// IsExpired checks if the token is expired
// Returns true if expired or close to expiring (within 60 seconds)
func (t *Token) IsExpired() bool {
	if t.IssuedAt.IsZero() {
		return true
	}
	// Expire 60 seconds early to prevent edge cases
	expiryBuffer := time.Duration(60) * time.Second
	expiryTime := time.Duration(t.ExpiresIn)*time.Second - expiryBuffer
	return time.Since(t.IssuedAt) >= expiryTime
}

// APIClient represents the Cisco Secure Access API client
type APIClient struct {
	TokenURL     string
	ClientID     string
	ClientSecret string
	Scopes       []string
	Token        *Token
	HTTPClient   *http.Client
	Region       string
}

// NewAPIClient creates a new API client instance
func NewAPIClient(tokenURL, clientID, clientSecret string, scopes []string, region string) *APIClient {
	if tokenURL == "" || clientID == "" || clientSecret == "" {
		return nil
	}

	return &APIClient{
		TokenURL:     tokenURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       scopes,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		Region: region,
	}
}

// GetToken fetches a new OAuth token
func (c *APIClient) GetToken() error {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", c.ClientID)
	data.Set("client_secret", c.ClientSecret)

	if len(c.Scopes) > 0 {
		data.Set("scope", strings.Join(c.Scopes, " "))
	}

	req, err := http.NewRequest("POST", c.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(c.ClientID, c.ClientSecret)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to request token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to obtain token, status: %d, response: %s", resp.StatusCode, string(body))
	}

	var token Token
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return fmt.Errorf("failed to decode token response: %w", err)
	}

	token.IssuedAt = time.Now()
	c.Token = &token
	return nil
}

// ensureToken ensures we have a valid token
func (c *APIClient) ensureToken() error {
	if c.Token == nil || c.Token.IsExpired() {
		return c.GetToken()
	}
	return nil
}

// Query executes an API request with automatic token refresh
func (c *APIClient) Query(scope, endpoint, operation string, requestData interface{}) (*http.Response, error) {
	baseURI := fmt.Sprintf("https://api.sse.cisco.com/%s/v2", scope)
	if scope == ScopeReports && c.Region != "" {
		baseURI = fmt.Sprintf("https://api.sse.cisco.com/%s.%s/v2", scope, c.Region)
	}

	// Handle full URLs or relative paths
	var url string
	if strings.HasPrefix(endpoint, "http") {
		url = endpoint
	} else {
		// Remove leading slash if present to avoid double slashes
		endpoint = strings.TrimPrefix(endpoint, "/")
		url = fmt.Sprintf("%s/%s", baseURI, endpoint)
	}

	maxRetries := 10
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Ensure we have a valid token
		if err := c.ensureToken(); err != nil {
			return nil, fmt.Errorf("failed to obtain token: %w", err)
		}

		// Create request body
		var body io.Reader
		if requestData != nil {
			jsonData, err := json.Marshal(requestData)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request data: %w", err)
			}
			body = bytes.NewBuffer(jsonData)
		}

		req, err := http.NewRequest(operation, url, body)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Set headers
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token.AccessToken))
		req.Header.Set("Content-Type", "application/json")

		// Execute request
		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}

		// Check for token expiration (401 Unauthorized)
		if resp.StatusCode == http.StatusUnauthorized && attempt < maxRetries-1 {
			resp.Body.Close()
			// Force token refresh
			c.Token = nil
			continue
		}

		// Check for Rate Limit (429)
		if resp.StatusCode == http.StatusTooManyRequests && attempt < maxRetries-1 {
			resp.Body.Close()
			// Wait for rate limit to reset (simple fixed backoff for now)
			time.Sleep(10 * time.Second)
			continue
		}

		// Check for Conflict (409) - specifically for locked ruleset
		if resp.StatusCode == http.StatusConflict && attempt < maxRetries-1 {
			bodyBytes, readErr := io.ReadAll(resp.Body)
			resp.Body.Close()
			if readErr != nil {
				// If we can't read the body, we can't check for the lock message.
				// Return the original response (re-constructed? no, we closed it).
				// Just fail or retry? Let's retry blindly if read fails? No, safer to error.
				return nil, fmt.Errorf("failed to read 409 response body: %w", readErr)
			}

			bodyStr := string(bodyBytes)
			if strings.Contains(bodyStr, "locked") {
				// It's a lock error, retry after delay
				time.Sleep(5 * time.Second)
				continue
			}

			// If not locked, restore the body and return the response
			resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			return resp, nil
		}

		// Add delay for state-changing operations to avoid 409 Conflict (locked ruleset)
		if operation == OperationPost || operation == OperationPut || operation == OperationPatch || operation == OperationDelete {
			time.Sleep(2 * time.Second)
		}

		// Return response (caller should check status code)
		return resp, nil
	}

	return nil, fmt.Errorf("all retry attempts failed: %w", lastErr)
}

// Helper functions

// extractID extracts an ID from a response map, handling various types
func extractID(data map[string]interface{}) string {
	if id, ok := data["id"]; ok {
		switch v := id.(type) {
		case string:
			return v
		case float64:
			return fmt.Sprintf("%.0f", v)
		case json.Number:
			return v.String()
		}
	}
	return ""
}

// extractCount attempts to extract item count from response data
func extractCount(data interface{}) int {
	switch v := data.(type) {
	case map[string]interface{}:
		// Check for count field
		if count, ok := v["count"].(float64); ok {
			return int(count)
		}
		// Check for data/items arrays
		if items, ok := v["data"].([]interface{}); ok {
			return len(items)
		}
		if items, ok := v["items"].([]interface{}); ok {
			return len(items)
		}
	case []interface{}:
		return len(v)
	}
	return 0
}
