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
	ServiceObjectsEndpoint       = "objects/serviceObjects"
	ServiceObjectDetailsEndpoint = "objects/serviceObjects/%d"
)

type ServiceObjectValue struct {
	Protocol string   `json:"protocol"`
	Ports    []string `json:"ports"`
}

type ServiceObject struct {
	ID          int64              `json:"id,omitempty"`
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Value       ServiceObjectValue `json:"value"`
	URL         string             `json:"url,omitempty"`
	CreatedAt   string             `json:"created_at,omitempty"`
	ModifiedAt  string             `json:"modified_at,omitempty"`
	ModifiedBy  string             `json:"modified_by,omitempty"`
}

type ServiceObjectResponse struct {
	Count   int             `json:"count"`
	Limit   int             `json:"limit"`
	Offset  int             `json:"offset"`
	Results []ServiceObject `json:"results"`
}

type CreateServiceObjectPayload struct {
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Value       ServiceObjectValue `json:"value"`
}

type UpdateServiceObjectPayload struct {
	Name        string             `json:"name,omitempty"`
	Description string             `json:"description,omitempty"`
	Value       ServiceObjectValue `json:"value,omitempty"`
}

// GetServiceObjects retrieves all service objects
func GetServiceObjects(client *APIClient) ([]ServiceObject, error) {
	var allObjects []ServiceObject
	offset := 0
	limit := 100

	for {
		endpoint := fmt.Sprintf("%s?offset=%d&limit=%d", ServiceObjectsEndpoint, offset, limit)
		resp, err := client.Query(ScopePolicies, endpoint, OperationGet, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get service objects: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("failed to get service objects. Status: %d, Response: %s", resp.StatusCode, string(body))
		}

		var result ServiceObjectResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		resp.Body.Close()

		allObjects = append(allObjects, result.Results...)

		if len(result.Results) == 0 || offset+len(result.Results) >= result.Count {
			break
		}
		offset += len(result.Results)
	}

	return allObjects, nil
}

// CreateServiceObject creates a new service object
func CreateServiceObject(client *APIClient, payload CreateServiceObjectPayload) (*ServiceObject, error) {
	resp, err := client.Query(ScopePolicies, ServiceObjectsEndpoint, OperationPost, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create service object: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create service object. Status: %d, Response: %s", resp.StatusCode, string(body))
	}

	var object ServiceObject
	if err := json.NewDecoder(resp.Body).Decode(&object); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &object, nil
}

// GetServiceObjectDetails retrieves details of a specific service object
func GetServiceObjectDetails(client *APIClient, objectID int64) (*ServiceObject, error) {
	endpoint := fmt.Sprintf(ServiceObjectDetailsEndpoint, objectID)
	resp, err := client.Query(ScopePolicies, endpoint, OperationGet, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get service object details: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get service object %d. Status: %d, Response: %s", objectID, resp.StatusCode, string(body))
	}

	var object ServiceObject
	if err := json.NewDecoder(resp.Body).Decode(&object); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &object, nil
}

// UpdateServiceObject updates a service object
func UpdateServiceObject(client *APIClient, objectID int64, payload UpdateServiceObjectPayload) (*ServiceObject, error) {
	endpoint := fmt.Sprintf(ServiceObjectDetailsEndpoint, objectID)
	resp, err := client.Query(ScopePolicies, endpoint, OperationPut, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to update service object: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to update service object %d. Status: %d, Response: %s", objectID, resp.StatusCode, string(body))
	}

	var object ServiceObject
	if err := json.NewDecoder(resp.Body).Decode(&object); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &object, nil
}

// DeleteServiceObject deletes a service object
func DeleteServiceObject(client *APIClient, objectID int64) error {
	endpoint := fmt.Sprintf(ServiceObjectDetailsEndpoint, objectID)
	resp, err := client.Query(ScopePolicies, endpoint, OperationDelete, nil)
	if err != nil {
		return fmt.Errorf("failed to delete service object: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete service object %d. Status: %d, Response: %s", objectID, resp.StatusCode, string(body))
	}

	return nil
}

func GetServiceObjectIDByName(client *APIClient, name string) (int64, error) {
	offset := 0
	limit := 100

	for {
		endpoint := fmt.Sprintf("%s?offset=%d&limit=%d", ServiceObjectsEndpoint, offset, limit)
		resp, err := client.Query(ScopePolicies, endpoint, OperationGet, nil)
		if err != nil {
			return 0, fmt.Errorf("failed to get service objects: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return 0, fmt.Errorf("failed to get service objects. Status: %d, Response: %s", resp.StatusCode, string(body))
		}

		var result ServiceObjectResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return 0, fmt.Errorf("failed to decode response: %w", err)
		}
		resp.Body.Close()

		for _, obj := range result.Results {
			if obj.Name == name {
				return obj.ID, nil
			}
		}

		if len(result.Results) == 0 || offset+len(result.Results) >= result.Count {
			break
		}
		offset += len(result.Results)
	}

	return 0, fmt.Errorf("service object with name '%s' not found", name)
}
