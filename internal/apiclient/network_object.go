// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apiclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

// Network Objects API endpoints
const (
	NetworkObjectsEndpoint                 = "objects/networkObjects"
	NetworkObjectDetailsEndpoint           = "objects/networkObjects/%s"
	NetworkObjectsReferencesEndpoint       = "objects/networkObjects/references"
	NetworkObjectReferencesDetailsEndpoint = "objects/networkObjects/%s/references"
)

type NetworkObjectValue struct {
	Addresses []string `json:"addresses"`
	Type      string   `json:"type"`
}

type NetworkObject struct {
	ID          string             `json:"id,omitempty"`
	Name        string             `json:"name"`
	Value       NetworkObjectValue `json:"value"`
	Description string             `json:"description,omitempty"`
}

type NetworkObjectPayload struct {
	Name        string      `json:"name"`
	Value       interface{} `json:"value"`
	Description string      `json:"description,omitempty"`
}

func GetNetworkObjects(client *APIClient) (interface{}, error) {
	resp, err := client.Query(ScopePolicies, NetworkObjectsEndpoint, OperationGet, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get network objects: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get network objects. Status: %d, Response: %s", resp.StatusCode, string(body))
	}

	var data interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	count := extractCount(data)
	fmt.Printf("Success. GET %s, retrieved %d objects\n", NetworkObjectsEndpoint, count)
	return data, nil
}

func GetNetworkObjectIDByName(client *APIClient, name string) (string, error) {
	data, err := GetNetworkObjects(client)
	if err != nil {
		return "", err
	}

	m, ok := data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected response format")
	}

	var results []interface{}
	if r, ok := m["results"].([]interface{}); ok {
		results = r
	} else if d, ok := m["data"].([]interface{}); ok {
		results = d
	} else {
		return "", fmt.Errorf("no results found in response")
	}

	for _, item := range results {
		obj, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		objName, ok := obj["name"].(string)
		if !ok {
			continue
		}

		if objName == name {
			idVal := obj["id"]
			switch v := idVal.(type) {
			case float64:
				return strconv.FormatInt(int64(v), 10), nil
			case string:
				return v, nil
			case int:
				return strconv.Itoa(v), nil
			default:
				return "", fmt.Errorf("unknown ID format for object %s", name)
			}
		}
	}

	return "", fmt.Errorf("network object with name '%s' not found", name)
}

func PostNetworkObject(client *APIClient, name string, value interface{}, description string) (string, error) {
	if err := ValidateNetworkObject(name, value); err != nil {
		return "", fmt.Errorf("validation failed: %w", err)
	}
	payload := NetworkObjectPayload{
		Name:        name,
		Value:       value,
		Description: description,
	}
	resp, err := client.Query(ScopePolicies, NetworkObjectsEndpoint, OperationPost, payload)
	if err != nil {
		return "", fmt.Errorf("failed to create network object: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to create network object '%s'. Status: %d, Response: %s", name, resp.StatusCode, string(body))
	}
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	objectID := extractID(result)
	if objectID == "" {
		return "", fmt.Errorf("no ID returned in response")
	}
	fmt.Printf("Success. Created network object '%s' with ID: %s\n", name, objectID)
	return objectID, nil
}

func GetNetworkObjectDetails(client *APIClient, objectID string) (interface{}, error) {
	endpoint := fmt.Sprintf(NetworkObjectDetailsEndpoint, objectID)
	resp, err := client.Query(ScopePolicies, endpoint, OperationGet, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get network object details: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode == http.StatusForbidden {
		// Try to read body to see if there is a useful error message
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("access forbidden (403) for object ID %s. Body: %s", objectID, string(body))
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get network object %s. Status: %d, Response: %s", objectID, resp.StatusCode, string(body))
	}
	var data interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	fmt.Printf("Success. Retrieved network object %s\n", objectID)
	return data, nil
}

func PutNetworkObjectDetails(client *APIClient, objectID, name string, value interface{}, description string) error {
	if err := ValidateNetworkObject(name, value); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	endpoint := fmt.Sprintf(NetworkObjectDetailsEndpoint, objectID)
	payload := NetworkObjectPayload{
		Name:        name,
		Value:       value,
		Description: description,
	}
	resp, err := client.Query(ScopePolicies, endpoint, OperationPut, payload)
	if err != nil {
		return fmt.Errorf("failed to update network object: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update network object %s. Status: %d, Response: %s", objectID, resp.StatusCode, string(body))
	}
	fmt.Printf("Success. Updated network object %s\n", objectID)
	return nil
}

func DeleteNetworkObject(client *APIClient, objectID string) error {
	endpoint := fmt.Sprintf(NetworkObjectDetailsEndpoint, objectID)
	resp, err := client.Query(ScopePolicies, endpoint, OperationDelete, nil)
	if err != nil {
		return fmt.Errorf("failed to delete network object: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete network object %s. Status: %d, Response: %s", objectID, resp.StatusCode, string(body))
	}
	fmt.Printf("Successfully deleted network object with ID: %s\n", objectID)
	return nil
}

func GetNetworkObjectsReferences(client *APIClient) (interface{}, error) {
	resp, err := client.Query(ScopePolicies, NetworkObjectsReferencesEndpoint, OperationGet, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get network objects references: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get references for network objects. Status: %d, Response: %s", resp.StatusCode, string(body))
	}
	var data interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	fmt.Println("Success. Retrieved network objects references")
	return data, nil
}

func GetNetworkObjectReferences(client *APIClient, objectID string) (interface{}, error) {
	endpoint := fmt.Sprintf(NetworkObjectReferencesDetailsEndpoint, objectID)
	resp, err := client.Query(ScopePolicies, endpoint, OperationGet, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get network object references: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get references for network object %s. Status: %d, Response: %s", objectID, resp.StatusCode, string(body))
	}
	var data interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	fmt.Printf("Success. Retrieved references for network object %s\n", objectID)
	return data, nil
}

// ValidateNetworkObject is a stub for validation logic
func ValidateNetworkObject(name string, value interface{}) error {
	if name == "" {
		return fmt.Errorf("name is required")
	}
	// Add more validation as needed
	return nil
}

// GetNetworkObjectNames retrieves and returns all network object names using GetNetworkObjects().
func GetNetworkObjectNames(client *APIClient) ([]string, error) {
	data, err := GetNetworkObjects(client)
	if err != nil {
		return nil, err
	}

	var names []string
	m, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format: not a map")
	}

	// Try to get multiple objects from "results" key
	if results, ok := m["results"].([]interface{}); ok {
		for _, obj := range results {
			objMap, ok := obj.(map[string]interface{})
			if !ok {
				continue
			}
			if name, ok := objMap["name"].(string); ok {
				names = append(names, name)
			}
		}
		return names, nil
	}

	// Fallback: handle single object (no "results" key, just a single object)
	if name, ok := m["name"].(string); ok {
		names = append(names, name)
		return names, nil
	}

	// Fallback: try to extract from a slice at the top level
	if arr, ok := data.([]interface{}); ok {
		for _, obj := range arr {
			objMap, ok := obj.(map[string]interface{})
			if !ok {
				continue
			}
			if name, ok := objMap["name"].(string); ok {
				names = append(names, name)
			}
		}
		return names, nil
	}

	return nil, fmt.Errorf("could not extract network object names from response")
}

// =========================
// Bulk Delete Utility
// =========================
// DeleteNetworkObjectByName deletes all network objects whose name starts with the given prefix.
func DeleteNetworkObjectByName(client *APIClient, prefix string) error {
	data, err := GetNetworkObjects(client)
	if err != nil {
		return fmt.Errorf("failed to get network objects: %w", err)
	}

	m, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected response format: not a map")
	}

	results, ok := m["results"].([]interface{})
	if !ok {
		return fmt.Errorf("unexpected response format: missing or invalid 'results' field")
	}

	var deletedCount int
	for _, obj := range results {
		objMap, ok := obj.(map[string]interface{})
		if !ok {
			continue
		}
		name, ok := objMap["name"].(string)
		if !ok || len(name) < len(prefix) || name[:len(prefix)] != prefix {
			continue
		}
		id, ok := objMap["id"].(string)
		if !ok {
			// Try if id is float64 (sometimes numbers are decoded as float64)
			if idNum, ok2 := objMap["id"].(float64); ok2 {
				id = fmt.Sprintf("%.0f", idNum)
			} else {
				fmt.Printf("[DeleteNetworkObjectByName] Could not get id for object with name %s\n", name)
				continue
			}
		}
		err := DeleteNetworkObject(client, id)
		if err != nil {
			fmt.Printf("[DeleteNetworkObjectByName] Failed to delete object ID %s (%s): %v\n", id, name, err)
		} else {
			fmt.Printf("[DeleteNetworkObjectByName] Deleted object ID %s (%s)\n", id, name)
			deletedCount++
		}
	}
	fmt.Printf("[DeleteNetworkObjectByName] Total deleted: %d\n", deletedCount)
	return nil
}
