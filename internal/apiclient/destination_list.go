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
	DestinationListsEndpoint        = "destinationlists"
	DestinationListsDetailsEndpoint = "destinationlists/%d"
	DestinationsDetailsEndpoint     = "destinationlists/%d/destinations"
	DestinationsRemoveEndpoint      = "destinationlists/%d/destinations/remove"
)

type Meta struct {
	DestinationCount int `json:"destinationCount"`
	DomainCount      int `json:"domainCount"`
	URLCount         int `json:"urlCount"`
	IPv4Count        int `json:"ipv4Count"`
	IPv6Count        int `json:"ipv6Count"`
	ApplicationCount int `json:"applicationCount"`
}

type DestinationList struct {
	ID                   int64  `json:"id"`
	OrganizationID       int64  `json:"organizationId"`
	Access               string `json:"access"`
	IsGlobal             bool   `json:"isGlobal"`
	Name                 string `json:"name"`
	ThirdpartyCategoryID int    `json:"thirdpartyCategoryId"`
	CreatedAt            int64  `json:"createdAt"`
	ModifiedAt           int64  `json:"modifiedAt"`
	IsMspDefault         bool   `json:"isMspDefault"`
	MarkedForDeletion    bool   `json:"markedForDeletion"`
	BundleTypeID         int    `json:"bundleTypeId"`
	Meta                 *Meta  `json:"meta,omitempty"`
}

type DestinationListResponse struct {
	Status struct {
		Code int    `json:"code"`
		Text string `json:"text"`
	} `json:"status"`
	Meta struct {
		Page  int `json:"page"`
		Limit int `json:"limit"`
		Total int `json:"total"`
	} `json:"meta"`
	//Data []int64 `json:"data"`
	Data []DestinationList `json:"data"`
}

type DestinationListDetailsResponse struct {
	Status struct {
		Code int    `json:"code"`
		Text string `json:"text"`
	} `json:"status"`
	Data DestinationList `json:"data"`
}

type Destination struct {
	ID          string `json:"id,omitempty"`
	Destination string `json:"destination"`
	Type        string `json:"type"`
	Comment     string `json:"comment,omitempty"`
}

type DestinationsResponse struct {
	Status struct {
		Code int    `json:"code"`
		Text string `json:"text"`
	} `json:"status"`
	Meta struct {
		Page  int `json:"page"`
		Limit int `json:"limit"`
		Total int `json:"total"`
	} `json:"meta"`
	Data []Destination `json:"data"`
}

type CreateDestinationListPayload struct {
	Access       string `json:"access"`
	IsGlobal     bool   `json:"isGlobal"`
	Name         string `json:"name"`
	BundleTypeID int    `json:"bundleTypeId"`
}

type UpdateDestinationListPayload struct {
	Name string `json:"name"`
}

type DestinationIdsList struct {
	DestinationIds []int64 `json:"destinationIds"`
}

func GetAllDestinationLists(client *APIClient) ([]int64, error) {
	var allLists []DestinationList
	page := 1
	hasMore := true

	for hasMore {
		endpoint := fmt.Sprintf("%s?page=%d&limit=100", DestinationListsEndpoint, page)
		resp, err := client.Query(ScopePolicies, endpoint, OperationGet, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to query destination lists: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("failed to get destination lists. Status code: %d, Response: %s", resp.StatusCode, string(body))
		}

		var result DestinationListResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		if len(result.Data) > 0 {
			allLists = append(allLists, result.Data...)
			// fmt.Printf("Success. GET %s, retrieved %d items (page %d)\n", endpoint, len(result.Data), page)
			page++
		} else {
			hasMore = false
		}
	}

	ids := make([]int64, len(allLists))
	for i, list := range allLists {
		ids[i] = list.ID
	}
	// fmt.Printf("Total destination lists retrieved: %d\n", len(ids))
	return ids, nil
}

func GetDestinationListIDByName(client *APIClient, name string) (int64, error) {
	page := 1
	hasMore := true

	for hasMore {
		endpoint := fmt.Sprintf("%s?page=%d&limit=100", DestinationListsEndpoint, page)
		resp, err := client.Query(ScopePolicies, endpoint, OperationGet, nil)
		if err != nil {
			return 0, fmt.Errorf("failed to query destination lists: %w", err)
		}
		// We must close the body in each iteration
		// But defer schedules it for function exit.
		// To avoid leaking file descriptors in a loop, we should wrap the body handling in a func or close explicitly.
		// For simplicity here, I'll close explicitly.

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return 0, fmt.Errorf("failed to get destination lists. Status code: %d, Response: %s", resp.StatusCode, string(body))
		}

		var result DestinationListResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return 0, fmt.Errorf("failed to decode response: %w", err)
		}
		resp.Body.Close()

		for _, list := range result.Data {
			if list.Name == name {
				return list.ID, nil
			}
		}

		if len(result.Data) > 0 {
			page++
		} else {
			hasMore = false
		}
	}

	return 0, fmt.Errorf("destination list with name '%s' not found", name)
}

func GetDestinations(client *APIClient, destinationListID int64) ([]int64, error) {
	var allDestinations []int64
	page := 1
	hasMore := true

	for hasMore {
		endpoint := fmt.Sprintf("%s?page=%d&limit=100", fmt.Sprintf(DestinationsDetailsEndpoint, destinationListID), page)
		// fmt.Println("GetDestinations endpoint", endpoint)
		resp, err := client.Query(ScopePolicies, endpoint, OperationGet, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to query destinations: %w", err)
		}
		// Read the response body into a buffer so we can print and decode it multiple times
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		// fmt.Printf("GetDestinations raw response: %s\n", string(bodyBytes))

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to get destinations in list %d. Status code: %d, Response: %s", destinationListID, resp.StatusCode, string(bodyBytes))
		}

		// The response may be a JSON array of string IDs, or an object with Data field (array of objects)
		var stringIDs []string
		if err := json.Unmarshal(bodyBytes, &stringIDs); err == nil {
			// Case: plain array of string IDs
			// stringIDs already set
		} else {
			// Try to decode as object with Data field (array of objects)
			type dataObj struct {
				Data []struct {
					ID string `json:"id"`
				} `json:"data"`
			}
			var obj dataObj
			if err2 := json.Unmarshal(bodyBytes, &obj); err2 == nil {
				for _, entry := range obj.Data {
					stringIDs = append(stringIDs, entry.ID)
				}
			} else {
				return nil, fmt.Errorf("failed to decode response as array or object-with-id: %w / %v", err, err2)
			}
		}

		if len(stringIDs) > 0 {
			for _, sid := range stringIDs {
				var id int64
				_, err := fmt.Sscan(sid, &id)
				if err != nil {
					return nil, fmt.Errorf("failed to convert destination ID '%s' to int64: %w", sid, err)
				}
				allDestinations = append(allDestinations, id)
			}
			// fmt.Printf("Success. GET %s, retrieved %d destinations (page %d)\n", endpoint, len(stringIDs), page)
			page++
		} else {
			hasMore = false
		}
	}

	// fmt.Printf("Total destinations retrieved: %d\n", len(allDestinations))
	// fmt.Println("allDestinations:", allDestinations)
	return allDestinations, nil
}

// GetDestinationsDetails fetches the full destination objects, not just IDs
func GetDestinationsDetails(client *APIClient, destinationListID int64) ([]Destination, error) {
	var allDestinations []Destination
	page := 1
	hasMore := true

	for hasMore {
		endpoint := fmt.Sprintf("%s?page=%d&limit=100", fmt.Sprintf(DestinationsDetailsEndpoint, destinationListID), page)
		resp, err := client.Query(ScopePolicies, endpoint, OperationGet, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to query destinations: %w", err)
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to get destinations in list %d. Status code: %d, Response: %s", destinationListID, resp.StatusCode, string(bodyBytes))
		}

		var result DestinationsResponse
		if err := json.Unmarshal(bodyBytes, &result); err == nil && result.Data != nil {
			if len(result.Data) > 0 {
				allDestinations = append(allDestinations, result.Data...)
				page++
			} else {
				hasMore = false
			}
		} else {
			// Fallback or error handling if structure is different
			// For now assume it matches DestinationsResponse
			hasMore = false
		}
	}
	return allDestinations, nil
}

func GetDestinationListDetails(client *APIClient, destinationListID int64) (*DestinationList, error) {
	endpoint := fmt.Sprintf(DestinationListsDetailsEndpoint, destinationListID)
	resp, err := client.Query(ScopePolicies, endpoint, OperationGet, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query destination list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get destination list %d. Status code: %d, Response: %s", destinationListID, resp.StatusCode, string(body))
	}

	var result DestinationListDetailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// fmt.Printf("Success. Retrieved destination list %d\n", destinationListID)
	return &result.Data, nil
}

func PatchDestinationList(client *APIClient, destinationListID int64, payload UpdateDestinationListPayload) (*DestinationList, error) {
	endpoint := fmt.Sprintf(DestinationListsDetailsEndpoint, destinationListID)
	resp, err := client.Query(ScopePolicies, endpoint, OperationPatch, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to update destination list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to update destination list %d. Status code: %d, Response: %s", destinationListID, resp.StatusCode, string(body))
	}

	var result DestinationListDetailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// fmt.Printf("Success. Updated destination list %d\n", destinationListID)
	return &result.Data, nil
}

func PostDestinationList(client *APIClient, payload CreateDestinationListPayload) (*DestinationList, error) {
	resp, err := client.Query(ScopePolicies, DestinationListsEndpoint, OperationPost, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create destination list. Status code: %d, Response: %s", resp.StatusCode, string(body))
	}

	var result DestinationListDetailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// fmt.Printf("Success. Created destination list with ID: %d\n", result.Data.ID)
	return &result.Data, nil
}

func PostDestinations(client *APIClient, destinationListID int64, destinations []Destination) error {
	endpoint := fmt.Sprintf(DestinationsDetailsEndpoint, destinationListID)
	resp, err := client.Query(ScopePolicies, endpoint, OperationPost, destinations)
	if err != nil {
		return fmt.Errorf("failed to add destinations: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to add destinations to list %d. Status code: %d, Response: %s", destinationListID, resp.StatusCode, string(body))
	}

	// fmt.Printf("Success. Added %d destination(s) to list %d\n", len(destinations), destinationListID)
	return nil
}

func DeleteDestinations(client *APIClient, destinationListID int64, destinationIDs []int64) error {
	endpoint := fmt.Sprintf(DestinationsRemoveEndpoint, destinationListID)
	resp, err := client.Query(ScopePolicies, endpoint, OperationDelete, destinationIDs)
	if err != nil {
		return fmt.Errorf("failed to delete destinations: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete destinations from list %d. Status code: %d, Response: %s", destinationListID, resp.StatusCode, string(body))
	}

	// fmt.Printf("Success. Deleted %d destination(s) from list %d\n", len(destinationIDs), destinationListID)
	return nil
}

func DeleteDestinationList(client *APIClient, destinationListID int64) error {
	endpoint := fmt.Sprintf(DestinationListsDetailsEndpoint, destinationListID)
	resp, err := client.Query(ScopePolicies, endpoint, OperationDelete, nil)
	if err != nil {
		return fmt.Errorf("failed to delete destination list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete destination list %d. Status code: %d, Response: %s", destinationListID, resp.StatusCode, string(body))
	}

	// fmt.Printf("Successfully deleted destination list %d\n", destinationListID)
	return nil
}
