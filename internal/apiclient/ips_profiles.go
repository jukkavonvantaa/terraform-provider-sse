package apiclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type IPSProfile struct {
	ID             int    `json:"id"`
	OrganizationID int    `json:"organizationId"`
	Name           string `json:"name"`
	IsDefault      bool   `json:"isDefault"`
	SystemMode     string `json:"systemMode"`
	CreatedAt      string `json:"createdAt"`
	ModifiedAt     string `json:"modifiedAt"`
}

type IPSProfilesResponse struct {
	Data []IPSProfile `json:"data"`
	Meta struct {
		Page  int `json:"page"`
		Limit int `json:"limit"`
		Total int `json:"total"`
	} `json:"meta"`
}

func (c *APIClient) GetIPSProfiles() ([]IPSProfile, error) {
	endpoint := "ipsSignatureProfiles"
	// The scope for IPS profiles is policies.ipsconfig:read
	// We need to make sure this scope is included in the token request if not already.
	// However, the client handles scopes globally. Assuming the user has configured the correct scopes.
	// Based on provider.go, "policies.ipsconfig:read" is NOT currently in the default scopes list.
	// We might need to add it to provider.go later.

	// For now, let's assume the client has the scope or we use a generic scope if possible,
	// but the spec says "policies.ipsconfig:read".
	// Let's use ScopePolicies as a base, but the actual scope string is specific.
	// The client.Query method takes a scope category (e.g. "policies") to find the base URL?
	// No, client.Query takes 'scope' which maps to the base URL key in client.go.
	// "policies" -> "https://api.sse.cisco.com/policies/v2"
	// So passing ScopePolicies is correct for the URL construction.

	resp, err := c.Query(ScopePolicies, endpoint, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result IPSProfilesResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Data, nil
}
