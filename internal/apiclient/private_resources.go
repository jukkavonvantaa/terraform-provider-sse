// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apiclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	PrivateResourcesEndpoint      = "privateResources"
	PrivateResourceGroupsEndpoint = "privateResourceGroups"
)

type PrivateResource struct {
	ID   int    `json:"resourceId"`
	Name string `json:"name"`
}

type PrivateResourceGroup struct {
	ID   int    `json:"resourceGroupId"`
	Name string `json:"name"`
}

type PrivateResourcesResponse struct {
	Items  []PrivateResource `json:"items"`
	Offset int               `json:"offset"`
	Limit  int               `json:"limit"`
	Total  int               `json:"total"`
}

type PrivateResourceGroupsResponse struct {
	Items  []PrivateResourceGroup `json:"items"`
	Offset int                    `json:"offset"`
	Limit  int                    `json:"limit"`
	Total  int                    `json:"total"`
}

func GetPrivateResourceIDByName(client *APIClient, name string) (int, error) {
	offset := 0
	limit := 100

	for {
		endpoint := fmt.Sprintf("%s?offset=%d&limit=%d", PrivateResourcesEndpoint, offset, limit)
		resp, err := client.Query(ScopePolicies, endpoint, OperationGet, nil)
		if err != nil {
			return 0, fmt.Errorf("failed to get private resources: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return 0, fmt.Errorf("failed to get private resources. Status: %d, Response: %s", resp.StatusCode, string(body))
		}

		var result PrivateResourcesResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return 0, fmt.Errorf("failed to decode response: %w", err)
		}
		resp.Body.Close()

		for _, item := range result.Items {
			if item.Name == name {
				return item.ID, nil
			}
		}

		if len(result.Items) == 0 || offset+len(result.Items) >= result.Total {
			break
		}
		offset += len(result.Items)
	}

	return 0, fmt.Errorf("private resource with name '%s' not found", name)
}

func GetPrivateResourceGroupIDByName(client *APIClient, name string) (int, error) {
	offset := 0
	limit := 100

	for {
		endpoint := fmt.Sprintf("%s?offset=%d&limit=%d", PrivateResourceGroupsEndpoint, offset, limit)
		resp, err := client.Query(ScopePolicies, endpoint, OperationGet, nil)
		if err != nil {
			return 0, fmt.Errorf("failed to get private resource groups: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return 0, fmt.Errorf("failed to get private resource groups. Status: %d, Response: %s", resp.StatusCode, string(body))
		}

		var result PrivateResourceGroupsResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return 0, fmt.Errorf("failed to decode response: %w", err)
		}
		resp.Body.Close()

		for _, item := range result.Items {
			if item.Name == name {
				return item.ID, nil
			}
		}

		if len(result.Items) == 0 || offset+len(result.Items) >= result.Total {
			break
		}
		offset += len(result.Items)
	}

	return 0, fmt.Errorf("private resource group with name '%s' not found", name)
}
