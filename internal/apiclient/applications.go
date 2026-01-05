// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apiclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Application struct {
	ID    int    `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type"`
}

type ApplicationsResponse struct {
	Data struct {
		Applications []Application `json:"applications"`
	} `json:"data"`
}

func (c *APIClient) GetApplications() ([]Application, error) {
	// Use reports/v2/applications endpoint which returns integer IDs
	endpoint := "applications"
	resp, err := c.Query("reports", endpoint, http.MethodGet, nil)
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

	var result ApplicationsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Data.Applications, nil
}
