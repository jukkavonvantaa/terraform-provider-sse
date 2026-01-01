// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apiclient

import (
	"encoding/json"
	"fmt"
	"io"
)

// NetworkTunnelGroup represents a network tunnel group
type NetworkTunnelGroup struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	OrganizationID int64  `json:"organizationId"`
	DeviceType     string `json:"deviceType"`
	Region         string `json:"region"`
	Status         string `json:"status"`
	CreatedAt      string `json:"createdAt"`
	ModifiedAt     string `json:"modifiedAt"`
}

// NetworkTunnelGroupList represents the response from the network tunnel groups endpoint
type NetworkTunnelGroupList struct {
	Data   []NetworkTunnelGroup `json:"data"`
	Offset int                  `json:"offset"`
	Limit  int                  `json:"limit"`
	Total  int                  `json:"total"`
}

// GetNetworkTunnelGroups retrieves a list of network tunnel groups
func GetNetworkTunnelGroups(client *APIClient) ([]NetworkTunnelGroup, error) {
	resp, err := client.Query("deployments", "networktunnelgroups", "GET", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get network tunnel groups. Status: %s, Body: %s", resp.Status, string(body))
	}

	var list NetworkTunnelGroupList
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return list.Data, nil
}
