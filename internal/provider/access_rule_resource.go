// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/cisco/terraform-provider-sse/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &AccessRuleResource{}
var _ resource.ResourceWithImportState = &AccessRuleResource{}

func NewAccessRuleResource() resource.Resource {
	return &AccessRuleResource{}
}

// AccessRuleResource defines the resource implementation.
type AccessRuleResource struct {
	client *apiclient.APIClient
}

// AccessRuleResourceModel describes the resource data model.
type AccessRuleResourceModel struct {
	ID             types.Int64  `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Action         types.String `tfsdk:"action"`
	Priority       types.Int64  `tfsdk:"priority"`
	IsEnabled      types.Bool   `tfsdk:"is_enabled"`
	RuleConditions types.Set    `tfsdk:"rule_conditions"`
	RuleSettings   types.Set    `tfsdk:"rule_settings"`
}

type RuleCondition struct {
	AttributeName     types.String `tfsdk:"attribute_name"`
	AttributeValue    types.String `tfsdk:"attribute_value"`
	AttributeOperator types.String `tfsdk:"attribute_operator"`
}

type RuleSetting struct {
	SettingName  types.String `tfsdk:"setting_name"`
	SettingValue types.String `tfsdk:"setting_value"`
}

var ruleConditionAttrTypes = map[string]attr.Type{
	"attribute_name":     types.StringType,
	"attribute_value":    types.StringType,
	"attribute_operator": types.StringType,
}

var ruleSettingAttrTypes = map[string]attr.Type{
	"setting_name":  types.StringType,
	"setting_value": types.StringType,
}

func (r *AccessRuleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_rule"
}

func (r *AccessRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Access Rule resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Access Rule ID",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Access Rule Name",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Access Rule Description",
			},
			"action": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Access Rule Action (allow, block)",
			},
			"priority": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Access Rule Priority. Must be between 1 and the total number of rules + 1.",
				PlanModifiers: []planmodifier.Int64{
					useStateForUnknownOrNull{},
				},
			},
			"is_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Is Access Rule Enabled",
				PlanModifiers: []planmodifier.Bool{
					useStateForUnknownOrNull{},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"rule_conditions": schema.SetNestedBlock{
				MarkdownDescription: "List of rule conditions",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"attribute_name": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Attribute Name",
							PlanModifiers: []planmodifier.String{
								caseInsensitiveNormalizer{},
							},
						},
						"attribute_value": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Attribute Value",
							PlanModifiers: []planmodifier.String{
								jsonNormalizer{},
							},
						},
						"attribute_operator": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Attribute Operator",
							PlanModifiers: []planmodifier.String{
								caseInsensitiveNormalizer{},
							},
						},
					},
				},
			},
			"rule_settings": schema.SetNestedBlock{
				MarkdownDescription: "List of rule settings",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"setting_name": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Setting Name",
							PlanModifiers: []planmodifier.String{
								caseInsensitiveNormalizer{},
							},
						},
						"setting_value": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Setting Value",
							PlanModifiers: []planmodifier.String{
								jsonNormalizer{},
							},
						},
					},
				},
			},
		},
	}
}

func (r *AccessRuleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*apiclient.APIClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *apiclient.APIClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *AccessRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AccessRuleResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var ruleSettings []RuleSetting
	resp.Diagnostics.Append(data.RuleSettings.ElementsAs(ctx, &ruleSettings, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(ruleSettings) == 0 {
		resp.Diagnostics.AddError("Missing Required Argument", "At least one rule_settings block is required.")
		return
	}

	var ruleConditions []RuleCondition
	resp.Diagnostics.Append(data.RuleConditions.ElementsAs(ctx, &ruleConditions, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform model to API request
	conditions := make([]apiclient.RuleCondition, len(ruleConditions))
	for i, c := range ruleConditions {
		valStr := c.AttributeValue.ValueString()
		var val interface{} = valStr

		// Try to unmarshal as JSON if it looks like a JSON object or array
		if strings.HasPrefix(strings.TrimSpace(valStr), "{") || strings.HasPrefix(strings.TrimSpace(valStr), "[") {
			var jsonVal interface{}
			if err := json.Unmarshal([]byte(valStr), &jsonVal); err == nil {
				val = jsonVal
			}
		} else if valStr == "true" || valStr == "false" {
			// Handle booleans
			if b, err := strconv.ParseBool(valStr); err == nil {
				val = b
			}
		}

		conditions[i] = apiclient.RuleCondition{
			AttributeName:     c.AttributeName.ValueString(),
			AttributeValue:    val,
			AttributeOperator: c.AttributeOperator.ValueString(),
		}
	}

	settings := make([]apiclient.RuleSetting, len(ruleSettings))
	for i, s := range ruleSettings {
		valStr := s.SettingValue.ValueString()
		var val interface{} = valStr

		// Try to unmarshal as JSON if it looks like a JSON object or array
		if strings.HasPrefix(strings.TrimSpace(valStr), "{") || strings.HasPrefix(strings.TrimSpace(valStr), "[") {
			var jsonVal interface{}
			if err := json.Unmarshal([]byte(valStr), &jsonVal); err == nil {
				val = jsonVal
			}
		} else if valStr == "true" || valStr == "false" {
			// Handle booleans
			if b, err := strconv.ParseBool(valStr); err == nil {
				val = b
			}
		} else if iVal, err := strconv.Atoi(valStr); err == nil {
			// Handle integers (some settings might be ints)
			// But be careful, some might be strings that look like ints.
			// For now, let's stick to string unless it's clearly JSON or boolean,
			// or maybe we should check if the API expects int.
			// The API client uses interface{}, so it depends on what the API expects.
			// In the user's example: "settingValue": 14843764 (int)
			// So we should probably try to parse as int if possible.
			val = iVal
		}

		settings[i] = apiclient.RuleSetting{
			SettingName:  s.SettingName.ValueString(),
			SettingValue: val,
		}
	}

	reqPayload := apiclient.RuleRequest{
		RuleName:        data.Name.ValueString(),
		RuleDescription: data.Description.ValueString(),
		RuleAction:      data.Action.ValueString(),
		RulePriority:    int(data.Priority.ValueInt64()),
		RuleIsEnabled:   data.IsEnabled.ValueBool(),
		RuleConditions:  conditions,
		RuleSettings:    settings,
	}

	createdRule, err := apiclient.CreateAccessRule(r.client, reqPayload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create access rule, got error: %s", err))
		return
	}

	data.ID = types.Int64Value(int64(createdRule.RuleID))

	// Only update priority from API if it wasn't specified in the plan (i.e. computed/unknown)
	// This prevents "Provider produced inconsistent result after apply" errors if the API
	// adjusts the priority (e.g. shifting rules).
	if data.Priority.IsUnknown() || data.Priority.IsNull() {
		data.Priority = types.Int64Value(int64(createdRule.RulePriority))
	}

	data.IsEnabled = types.BoolValue(createdRule.RuleIsEnabled)
	// Update other fields if API modifies them (e.g. default values)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AccessRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AccessRuleResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := apiclient.GetAccessRuleDetails(r.client, int(data.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read access rule, got error: %s", err))
		return
	}

	// DEBUG: Print rule details
	fmt.Printf("[DEBUG] AccessRule Read ID=%d Priority=%d\n", rule.RuleID, rule.RulePriority)
	for i, c := range rule.RuleConditions {
		fmt.Printf("[DEBUG] Condition[%d]: Name=%s Op=%s ValueType=%T Value=%v\n", i, c.AttributeName, c.AttributeOperator, c.AttributeValue, c.AttributeValue)
	}
	for i, s := range rule.RuleSettings {
		fmt.Printf("[DEBUG] Setting[%d]: Name=%s ValueType=%T Value=%v\n", i, s.SettingName, s.SettingValue, s.SettingValue)
	}

	data.Name = types.StringValue(rule.RuleName)
	if rule.RuleDescription != "" {
		data.Description = types.StringValue(rule.RuleDescription)
	} else {
		data.Description = types.StringNull()
	}
	data.Action = types.StringValue(rule.RuleAction)
	data.Priority = types.Int64Value(int64(rule.RulePriority))
	data.IsEnabled = types.BoolValue(rule.RuleIsEnabled)

	// Map conditions and settings back to Terraform model
	if len(rule.RuleConditions) > 0 {
		conditions := make([]RuleCondition, len(rule.RuleConditions))
		for i, c := range rule.RuleConditions {
			valStr := ""
			switch v := c.AttributeValue.(type) {
			case string:
				valStr = v
			case int, int32, int64:
				valStr = fmt.Sprintf("%d", v)
			case float64:
				valStr = fmt.Sprintf("%.0f", v) // Assuming integer values for now
			case bool:
				valStr = fmt.Sprintf("%t", v)
			case []interface{}, map[string]interface{}:
				b, err := json.Marshal(v)
				if err == nil {
					valStr = string(b)
				} else {
					valStr = fmt.Sprintf("%v", v)
				}
			default:
				valStr = fmt.Sprintf("%v", v)
			}

			conditions[i] = RuleCondition{
				AttributeName:     types.StringValue(c.AttributeName),
				AttributeValue:    types.StringValue(valStr),
				AttributeOperator: types.StringValue(c.AttributeOperator),
			}
		}
		var diags diag.Diagnostics
		data.RuleConditions, diags = types.SetValueFrom(ctx, types.ObjectType{AttrTypes: ruleConditionAttrTypes}, conditions)
		resp.Diagnostics.Append(diags...)
	} else {
		data.RuleConditions = types.SetNull(types.ObjectType{AttrTypes: ruleConditionAttrTypes})
	}

	if len(rule.RuleSettings) > 0 {
		settings := make([]RuleSetting, len(rule.RuleSettings))
		for i, s := range rule.RuleSettings {
			valStr := ""
			switch v := s.SettingValue.(type) {
			case string:
				valStr = v
			case int, int32, int64:
				valStr = fmt.Sprintf("%d", v)
			case float64:
				valStr = fmt.Sprintf("%.0f", v)
			case bool:
				valStr = fmt.Sprintf("%t", v)
			case []interface{}, map[string]interface{}:
				b, err := json.Marshal(v)
				if err == nil {
					valStr = string(b)
				} else {
					valStr = fmt.Sprintf("%v", v)
				}
			default:
				valStr = fmt.Sprintf("%v", v)
			}

			settings[i] = RuleSetting{
				SettingName:  types.StringValue(s.SettingName),
				SettingValue: types.StringValue(valStr),
			}
		}
		var diags diag.Diagnostics
		data.RuleSettings, diags = types.SetValueFrom(ctx, types.ObjectType{AttrTypes: ruleSettingAttrTypes}, settings)
		resp.Diagnostics.Append(diags...)
	} else {
		data.RuleSettings = types.SetNull(types.ObjectType{AttrTypes: ruleSettingAttrTypes})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AccessRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AccessRuleResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var ruleSettings []RuleSetting
	resp.Diagnostics.Append(data.RuleSettings.ElementsAs(ctx, &ruleSettings, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(ruleSettings) == 0 {
		resp.Diagnostics.AddError("Missing Required Argument", "At least one rule_settings block is required.")
		return
	}

	var ruleConditions []RuleCondition
	resp.Diagnostics.Append(data.RuleConditions.ElementsAs(ctx, &ruleConditions, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform model to API request
	conditions := make([]apiclient.RuleCondition, len(ruleConditions))
	for i, c := range ruleConditions {
		valStr := c.AttributeValue.ValueString()
		var val interface{} = valStr

		if strings.HasPrefix(strings.TrimSpace(valStr), "{") || strings.HasPrefix(strings.TrimSpace(valStr), "[") {
			var jsonVal interface{}
			if err := json.Unmarshal([]byte(valStr), &jsonVal); err == nil {
				val = jsonVal
			}
		} else if valStr == "true" || valStr == "false" {
			if b, err := strconv.ParseBool(valStr); err == nil {
				val = b
			}
		}

		conditions[i] = apiclient.RuleCondition{
			AttributeName:     c.AttributeName.ValueString(),
			AttributeValue:    val,
			AttributeOperator: c.AttributeOperator.ValueString(),
		}
	}

	settings := make([]apiclient.RuleSetting, len(ruleSettings))
	for i, s := range ruleSettings {
		valStr := s.SettingValue.ValueString()
		var val interface{} = valStr

		if strings.HasPrefix(strings.TrimSpace(valStr), "{") || strings.HasPrefix(strings.TrimSpace(valStr), "[") {
			var jsonVal interface{}
			if err := json.Unmarshal([]byte(valStr), &jsonVal); err == nil {
				val = jsonVal
			}
		} else if valStr == "true" || valStr == "false" {
			if b, err := strconv.ParseBool(valStr); err == nil {
				val = b
			}
		} else if iVal, err := strconv.Atoi(valStr); err == nil {
			val = iVal
		}

		settings[i] = apiclient.RuleSetting{
			SettingName:  s.SettingName.ValueString(),
			SettingValue: val,
		}
	}

	reqPayload := apiclient.RuleRequestUpdate{
		RuleName:        data.Name.ValueString(),
		RuleDescription: data.Description.ValueString(),
		RuleAction:      data.Action.ValueString(),
		RulePriority:    int(data.Priority.ValueInt64()),
		RuleIsEnabled:   data.IsEnabled.ValueBool(),
		RuleConditions:  conditions,
		RuleSettings:    settings,
	}

	updatedRule, err := apiclient.UpdateAccessRule(r.client, int(data.ID.ValueInt64()), reqPayload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update access rule, got error: %s", err))
		return
	}

	// Only update priority from API if it wasn't specified in the plan (i.e. computed/unknown)
	if data.Priority.IsUnknown() || data.Priority.IsNull() {
		data.Priority = types.Int64Value(int64(updatedRule.RulePriority))
	}

	data.IsEnabled = types.BoolValue(updatedRule.RuleIsEnabled)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AccessRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AccessRuleResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := apiclient.DeleteAccessRule(r.client, int(data.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete access rule, got error: %s", err))
		return
	}
}

func (r *AccessRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id := req.ID
	if _, err := strconv.ParseInt(id, 10, 64); err != nil {
		// Not a number, try to find by name
		foundID, err := apiclient.GetAccessRuleIDByName(r.client, id)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error importing access rule",
				fmt.Sprintf("Could not find access rule with name '%s': %s", id, err.Error()),
			)
			return
		}
		req.ID = fmt.Sprintf("%d", foundID)
	}
	idInt, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error importing access rule",
			fmt.Sprintf("Could not parse ID '%s' as int64: %s", req.ID, err.Error()),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idInt)...)
}

// jsonNormalizer normalizes JSON strings to compact format to avoid diffs due to whitespace.
type jsonNormalizer struct{}

func (m jsonNormalizer) Description(ctx context.Context) string {
	return "Normalizes JSON strings to compact format."
}

func (m jsonNormalizer) MarkdownDescription(ctx context.Context) string {
	return "Normalizes JSON strings to compact format."
}

func (m jsonNormalizer) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// If the value is unknown or null, we don't need to do anything.
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	val := req.ConfigValue.ValueString()

	// Try to unmarshal as generic JSON
	var tmp interface{}
	if err := json.Unmarshal([]byte(val), &tmp); err != nil {
		// Not valid JSON, leave as is
		return
	}

	// Marshal back to compact JSON
	compact, err := json.Marshal(tmp)
	if err != nil {
		// Should not happen if Unmarshal succeeded, but safe to ignore
		return
	}

	resp.PlanValue = types.StringValue(string(compact))
}

// caseInsensitiveNormalizer normalizes string values to match state if they differ only by case.
type caseInsensitiveNormalizer struct{}

func (m caseInsensitiveNormalizer) Description(ctx context.Context) string {
	return "Normalizes string values to match state if they differ only by case."
}

func (m caseInsensitiveNormalizer) MarkdownDescription(ctx context.Context) string {
	return "Normalizes string values to match state if they differ only by case."
}

func (m caseInsensitiveNormalizer) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// If config is null or unknown, do nothing
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	// If state is null or unknown, we can't normalize against it
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}

	configVal := req.ConfigValue.ValueString()
	stateVal := req.StateValue.ValueString()

	if strings.EqualFold(configVal, stateVal) {
		resp.PlanValue = req.StateValue
	}
}

// useStateForUnknownOrNull copies the state value to the plan if the config value is unknown or null.
// This is useful for Computed fields that should persist their state if removed from config.
type useStateForUnknownOrNull struct{}

func (m useStateForUnknownOrNull) Description(ctx context.Context) string {
	return "Copies the state value to the plan if the config value is unknown or null."
}

func (m useStateForUnknownOrNull) MarkdownDescription(ctx context.Context) string {
	return "Copies the state value to the plan if the config value is unknown or null."
}

func (m useStateForUnknownOrNull) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		if !req.StateValue.IsUnknown() && !req.StateValue.IsNull() {
			resp.PlanValue = req.StateValue
		}
	}
}

func (m useStateForUnknownOrNull) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		if !req.StateValue.IsUnknown() && !req.StateValue.IsNull() {
			resp.PlanValue = req.StateValue
		}
	}
}
