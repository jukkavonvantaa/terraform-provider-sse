package apiclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	TenantControlsProfilesEndpoint = "tenantControls/profiles"
)

type TenantControlsProfile struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description,omitempty"`
	IsDefault      bool   `json:"is_default"`
	OrganizationID int64  `json:"org_id"`
	CreatedAt      string `json:"created_at"`
	ModifiedAt     string `json:"modified_at"`
}

func GetTenantControlsProfiles(client *APIClient) ([]TenantControlsProfile, error) {
	var allProfiles []TenantControlsProfile
	page := 1
	limit := 100
	hasMore := true

	for hasMore {
		endpoint := fmt.Sprintf("%s?page=%d&limit=%d", TenantControlsProfilesEndpoint, page, limit)
		resp, err := client.Query(ScopePolicies, endpoint, OperationGet, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to query tenant controls profiles: %w", err)
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to get tenant controls profiles. Status code: %d, Response: %s", resp.StatusCode, string(bodyBytes))
		}

		// The response is a direct array of objects
		var pageProfiles []TenantControlsProfile
		if err := json.Unmarshal(bodyBytes, &pageProfiles); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		if len(pageProfiles) > 0 {
			allProfiles = append(allProfiles, pageProfiles...)
			if len(pageProfiles) < limit {
				hasMore = false
			} else {
				page++
			}
		} else {
			hasMore = false
		}
	}

	return allProfiles, nil
}
