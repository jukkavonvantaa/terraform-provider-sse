// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apiclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ApplicationCategory struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	CreatedAt         int    `json:"createdAt"`
	ModifiedAt        int    `json:"modifiedAt"`
	ApplicationsCount int    `json:"applicationsCount"`
}

func (c *APIClient) GetApplicationCategories() ([]ApplicationCategory, error) {
	var allCategories []ApplicationCategory
	page := 1
	limit := 100

	for {
		endpoint := fmt.Sprintf("applicationCategories?page=%d&limit=%d", page, limit)
		resp, err := c.Query("policies", endpoint, http.MethodGet, nil)
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
		// fmt.Printf("DEBUG: Body: %s\n", string(body))

		var pageCategories []ApplicationCategory

		// Try unmarshalling as map[string]ApplicationCategory first (observed behavior)
		var mapCategories map[string]ApplicationCategory
		if err := json.Unmarshal(body, &mapCategories); err == nil && len(mapCategories) > 0 {
			fmt.Printf("DEBUG: Found %d categories in map\n", len(mapCategories))
			for _, cat := range mapCategories {
				pageCategories = append(pageCategories, cat)
			}
		} else {
			// Fallback to array or wrapper
			if err := json.Unmarshal(body, &pageCategories); err != nil {
				// Try wrapper
				var wrapper struct {
					Data  []ApplicationCategory `json:"data"`
					Items []ApplicationCategory `json:"items"`
				}
				if err := json.Unmarshal(body, &wrapper); err == nil {
					if len(wrapper.Data) > 0 {
						pageCategories = wrapper.Data
					} else if len(wrapper.Items) > 0 {
						pageCategories = wrapper.Items
					} else {
						// Empty list
						pageCategories = []ApplicationCategory{}
					}
				} else {
					// If all fail, return error
					return nil, fmt.Errorf("failed to decode response: %w", err)
				}
			}
		}

		allCategories = append(allCategories, pageCategories...)

		if len(pageCategories) < limit {
			break
		}
		page++
	}

	return allCategories, nil
}
