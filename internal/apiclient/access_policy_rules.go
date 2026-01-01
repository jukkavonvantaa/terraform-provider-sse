// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apiclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Access Rules API endpoints
const (
	AccessRulesEndpoint       = "rules"
	AccessRuleDetailsEndpoint = "rules/%d"
)

// RuleSetting represents a setting on a rule
type RuleSetting struct {
	SettingName  string      `json:"settingName"`
	SettingValue interface{} `json:"settingValue"`
	CreatedAt    string      `json:"createdAt,omitempty"`
	ModifiedAt   string      `json:"modifiedAt,omitempty"`
}

// RuleCondition represents a condition on a rule
type RuleCondition struct {
	AttributeName     string      `json:"attributeName"`
	AttributeValue    interface{} `json:"attributeValue"`
	AttributeOperator string      `json:"attributeOperator"`
}

// Rule represents an access rule
type Rule struct {
	OrganizationID  int             `json:"organizationId,omitempty"`
	RuleID          int             `json:"ruleId,omitempty"`
	RuleName        string          `json:"ruleName"`
	RuleDescription string          `json:"ruleDescription,omitempty"`
	RuleAction      string          `json:"ruleAction"`
	RulePriority    int             `json:"rulePriority,omitempty"`
	RuleIsDefault   bool            `json:"ruleIsDefault,omitempty"`
	RuleIsEnabled   bool            `json:"ruleIsEnabled,omitempty"`
	RuleConditions  []RuleCondition `json:"ruleConditions"`
	RuleSettings    []RuleSetting   `json:"ruleSettings"`
	ModifiedBy      string          `json:"modifiedBy,omitempty"`
	ModifiedAt      string          `json:"modifiedAt,omitempty"`
	CreatedAt       string          `json:"createdAt,omitempty"`
}

// RuleRequest represents the payload to create a rule
type RuleRequest struct {
	RuleName        string          `json:"ruleName"`
	RuleDescription string          `json:"ruleDescription,omitempty"`
	RuleAction      string          `json:"ruleAction"`
	RulePriority    int             `json:"rulePriority"`
	RuleIsEnabled   bool            `json:"ruleIsEnabled"`
	RuleConditions  []RuleCondition `json:"ruleConditions"`
	RuleSettings    []RuleSetting   `json:"ruleSettings"`
}

// RuleRequestUpdate represents the payload to update a rule
type RuleRequestUpdate struct {
	RuleName        string          `json:"ruleName"`
	RuleDescription string          `json:"ruleDescription,omitempty"`
	RuleAction      string          `json:"ruleAction"`
	RulePriority    int             `json:"rulePriority"`
	RuleIsEnabled   bool            `json:"ruleIsEnabled"`
	RuleConditions  []RuleCondition `json:"ruleConditions"`
	RuleSettings    []RuleSetting   `json:"ruleSettings"`
}

// RulesResponse represents the list of rules response
type RulesResponse struct {
	Count   int    `json:"count"`
	Result  []Rule `json:"result"`
	Data    []Rule `json:"data"`
	Items   []Rule `json:"items"`
	Results []Rule `json:"results"`
}

// GetAccessRules lists all access rules
func GetAccessRules(client *APIClient) ([]Rule, error) {
	resp, err := client.Query(ScopePolicies, AccessRulesEndpoint, OperationGet, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get access rules: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get access rules. Status: %d, Response: %s", resp.StatusCode, string(body))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var data RulesResponse
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Handle inconsistent API response fields
	if len(data.Result) == 0 {
		if len(data.Data) > 0 {
			data.Result = data.Data
		} else if len(data.Items) > 0 {
			data.Result = data.Items
		} else if len(data.Results) > 0 {
			data.Result = data.Results
		}
	}

	if data.Count > 0 && len(data.Result) == 0 {
		fmt.Printf("[DEBUG] Response body: %s\n", string(bodyBytes))
	}

	fmt.Printf("Success. GET %s, retrieved %d rules\n", AccessRulesEndpoint, data.Count)
	return data.Result, nil
}

// CreateAccessRule creates a new access rule
func CreateAccessRule(client *APIClient, rule RuleRequest) (*Rule, error) {
	resp, err := client.Query(ScopePolicies, AccessRulesEndpoint, OperationPost, rule)
	if err != nil {
		return nil, fmt.Errorf("failed to create access rule: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create access rule. Status: %d, Response: %s", resp.StatusCode, string(body))
	}

	var createdRule Rule
	if err := json.NewDecoder(resp.Body).Decode(&createdRule); err != nil {
		return nil, fmt.Errorf("failed to deciode response: %w", err)
	}

	fmt.Printf("Success. Created access rule '%s' with ID: %d\n", createdRule.RuleName, createdRule.RuleID)
	return &createdRule, nil
}

// GetAccessRuleDetails gets details of a specific access rule
func GetAccessRuleDetails(client *APIClient, ruleID int) (*Rule, error) {
	endpoint := fmt.Sprintf(AccessRuleDetailsEndpoint, ruleID)
	resp, err := client.Query(ScopePolicies, endpoint, OperationGet, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get access rule details: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get access rule %d. Status: %d, Response: %s", ruleID, resp.StatusCode, string(body))
	}

	var rule Rule
	if err := json.NewDecoder(resp.Body).Decode(&rule); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	fmt.Printf("Success. Retrieved access rule %d\n", ruleID)
	return &rule, nil
}

// UpdateAccessRule updates an existing access rule
func UpdateAccessRule(client *APIClient, ruleID int, rule RuleRequestUpdate) (*Rule, error) {
	endpoint := fmt.Sprintf(AccessRuleDetailsEndpoint, ruleID)
	resp, err := client.Query(ScopePolicies, endpoint, OperationPut, rule)
	if err != nil {
		return nil, fmt.Errorf("failed to update access rule: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to update access rule %d. Status: %d, Response: %s", ruleID, resp.StatusCode, string(body))
	}

	var updatedRule Rule
	if err := json.NewDecoder(resp.Body).Decode(&updatedRule); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	fmt.Printf("Success. Updated access rule %d\n", ruleID)
	return &updatedRule, nil
}

// DeleteAccessRule deletes an access rule
func DeleteAccessRule(client *APIClient, ruleID int) error {
	endpoint := fmt.Sprintf(AccessRuleDetailsEndpoint, ruleID)
	resp, err := client.Query(ScopePolicies, endpoint, OperationDelete, nil)
	if err != nil {
		return fmt.Errorf("failed to delete access rule: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete access rule %d. Status: %d, Response: %s", ruleID, resp.StatusCode, string(body))
	}

	fmt.Printf("Successfully deleted access rule with ID: %d\n", ruleID)
	return nil
}

func GetAccessRuleIDByName(client *APIClient, name string) (int, error) {
	resp, err := client.Query(ScopePolicies, AccessRulesEndpoint, OperationGet, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to get access rules: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("failed to get access rules. Status: %d, Response: %s", resp.StatusCode, string(body))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body: %w", err)
	}

	var data RulesResponse
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	// Handle inconsistent API response fields
	if len(data.Result) == 0 {
		if len(data.Data) > 0 {
			data.Result = data.Data
		} else if len(data.Items) > 0 {
			data.Result = data.Items
		} else if len(data.Results) > 0 {
			data.Result = data.Results
		}
	}

	for _, rule := range data.Result {
		if rule.RuleName == name {
			return rule.RuleID, nil
		}
	}

	return 0, fmt.Errorf("access rule with name '%s' not found", name)
}
