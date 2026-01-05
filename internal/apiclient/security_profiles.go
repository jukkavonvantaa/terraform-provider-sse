// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apiclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type SecurityProfile struct {
	ID             int    `json:"id"`
	OrganizationID int    `json:"organizationId"`
	IsDefault      bool   `json:"isDefault"`
	Name           string `json:"name"`
	CreatedAt      int64  `json:"createdAt"`
	ModifiedAt     int64  `json:"modifiedAt"`
	Priority       int    `json:"priority"`
}

func (c *APIClient) GetSecurityProfiles() ([]SecurityProfile, error) {
	endpoint := "securityProfiles"
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

	// The API returns a direct array of objects
	var profiles []SecurityProfile
	if err := json.Unmarshal(body, &profiles); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return profiles, nil
}
