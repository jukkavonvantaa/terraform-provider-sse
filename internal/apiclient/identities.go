package apiclient

import (
	"encoding/json"
	"fmt"
	"io"
)

type IdentityType struct {
	ID    int    `json:"id"`
	Type  string `json:"type"`
	Label string `json:"label"`
}

type Identity struct {
	ID      int64        `json:"id"`
	Label   string       `json:"label"`
	Type    IdentityType `json:"type"`
	Deleted bool         `json:"deleted"`
}

type IdentityList struct {
	Data []Identity `json:"data"`
}

func (c *APIClient) GetIdentities() ([]Identity, error) {
	endpoint := "identities?limit=100&offset=0"
	resp, err := c.Query(ScopeReports, endpoint, OperationGet, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result IdentityList
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Data, nil
}
