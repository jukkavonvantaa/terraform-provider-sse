package apiclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ContentCategory struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	CreatedAt  int    `json:"createdAt"`
	ModifiedAt int    `json:"modifiedAt"`
}

func (c *APIClient) GetContentCategories() ([]ContentCategory, error) {
	// The API supports pagination but the spec says "Get all Content Category settings".
	// Let's assume we can fetch them all or loop if needed.
	// Spec says limit default 10, max 100. So we MUST loop.

	var allCategories []ContentCategory
	page := 1
	limit := 100

	for {
		endpoint := fmt.Sprintf("categorySettings?page=%d&limit=%d", page, limit)
		resp, err := c.Query("policies", endpoint, http.MethodGet, nil)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
		}

		var pageCategories []ContentCategory
		if err := json.NewDecoder(resp.Body).Decode(&pageCategories); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		allCategories = append(allCategories, pageCategories...)

		if len(pageCategories) < limit {
			break
		}
		page++
	}

	return allCategories, nil
}
