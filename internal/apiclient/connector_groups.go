// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apiclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type ConnectorGroup struct {
	ID                          int    `json:"id"`
	Name                        string `json:"name"`
	Location                    string `json:"location"`
	Environment                 string `json:"environment"`
	ProvisioningKey             string `json:"provisioningKey,omitempty"`
	ProvisioningKeyExpiresAt    string `json:"provisioningKeyExpiresAt,omitempty"`
	BaseImageDownloadURL        string `json:"baseImageDownloadUrl,omitempty"`
	Status                      string `json:"status,omitempty"`
	StatusUpdatedAt             string `json:"statusUpdatedAt,omitempty"`
	ConnectorsCount             int    `json:"connectorsCount,omitempty"`
	ResourceIDs                 []int  `json:"resourceIds,omitempty"`
	CreatedAt                   string `json:"createdAt,omitempty"`
	ModifiedAt                  string `json:"modifiedAt,omitempty"`
	ConnectedConnectorsCount    int    `json:"connectedConnectorsCount,omitempty"`
	DisconnectedConnectorsCount int    `json:"disconnectedConnectorsCount,omitempty"`
}

type ConnectorGroupCreateRequest struct {
	Name        string `json:"name"`
	Location    string `json:"location"`
	Environment string `json:"environment"`
}

type ConnectorGroupUpdateRequest struct {
	Name     string `json:"name"`
	Location string `json:"location"`
	// Environment cannot be updated according to spec (only name and location)
}

type ConnectorGroupsResponse struct {
	Data   []ConnectorGroup `json:"data"`
	Total  int              `json:"total"`
	Limit  int              `json:"limit"`
	Offset int              `json:"offset"`
}

func (c *APIClient) GetConnectorGroups(limit, offset int) ([]ConnectorGroup, error) {
	endpoint := fmt.Sprintf("connectorGroups?limit=%d&offset=%d", limit, offset)
	resp, err := c.Query("deployments", endpoint, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result ConnectorGroupsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Data, nil
}

func (c *APIClient) GetConnectorGroupByName(name string) (*ConnectorGroup, error) {
	// Construct filter JSON
	filter := map[string]string{"name": name}
	filterBytes, err := json.Marshal(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal filter: %w", err)
	}
	filterStr := url.QueryEscape(string(filterBytes))

	endpoint := fmt.Sprintf("connectorGroups?filters=%s", filterStr)
	resp, err := c.Query("deployments", endpoint, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result ConnectorGroupsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	// Return the first match
	return &result.Data[0], nil
}

func (c *APIClient) GetConnectorGroup(id int) (*ConnectorGroup, error) {
	endpoint := fmt.Sprintf("connectorGroups/%d?includeProvisioningKey=true", id)
	resp, err := c.Query("deployments", endpoint, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result ConnectorGroup
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *APIClient) CreateConnectorGroup(req ConnectorGroupCreateRequest) (*ConnectorGroup, error) {
	endpoint := "connectorGroups"
	resp, err := c.Query("deployments", endpoint, http.MethodPost, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result ConnectorGroup
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *APIClient) UpdateConnectorGroup(id int, req ConnectorGroupUpdateRequest) (*ConnectorGroup, error) {
	endpoint := fmt.Sprintf("connectorGroups/%d", id)
	resp, err := c.Query("deployments", endpoint, http.MethodPut, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result ConnectorGroup
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *APIClient) DeleteConnectorGroup(id int) error {
	endpoint := fmt.Sprintf("connectorGroups/%d", id)
	resp, err := c.Query("deployments", endpoint, http.MethodDelete, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
