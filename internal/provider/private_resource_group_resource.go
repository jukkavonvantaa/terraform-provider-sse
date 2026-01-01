// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/cisco/terraform-provider-sse/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &PrivateResourceGroupResource{}
var _ resource.ResourceWithImportState = &PrivateResourceGroupResource{}

func NewPrivateResourceGroupResource() resource.Resource {
	return &PrivateResourceGroupResource{}
}

// PrivateResourceGroupResource defines the resource implementation.
type PrivateResourceGroupResource struct {
	client *apiclient.APIClient
}

// PrivateResourceGroupResourceModel describes the resource data model.
type PrivateResourceGroupResourceModel struct {
	ID          types.String  `tfsdk:"id"`
	Name        types.String  `tfsdk:"name"`
	Description types.String  `tfsdk:"description"`
	ResourceIDs []types.Int64 `tfsdk:"resource_ids"`
}

func (r *PrivateResourceGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_private_resource_group"
}

func (r *PrivateResourceGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Private Resource Group resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Private Resource Group ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Private Resource Group Name",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Private Resource Group Description",
			},
			"resource_ids": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.Int64Type,
				MarkdownDescription: "List of Resource IDs",
			},
		},
	}
}

func (r *PrivateResourceGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PrivateResourceGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PrivateResourceGroupResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resourceIDs := make([]int, 0)
	for _, id := range data.ResourceIDs {
		resourceIDs = append(resourceIDs, int(id.ValueInt64()))
	}

	reqBody := map[string]interface{}{
		"name":        data.Name.ValueString(),
		"description": data.Description.ValueString(),
		"resourceIds": resourceIDs,
	}

	respHTTP, err := r.client.Query("policies", "privateResourceGroups", "POST", reqBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating private resource group",
			"Could not create private resource group, unexpected error: "+err.Error(),
		)
		return
	}
	defer respHTTP.Body.Close()

	body, _ := io.ReadAll(respHTTP.Body)

	if respHTTP.StatusCode != 200 && respHTTP.StatusCode != 201 {
		resp.Diagnostics.AddError(
			"Error creating private resource group",
			fmt.Sprintf("Could not create private resource group. Status: %s, Body: %s", respHTTP.Status, string(body)),
		)
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		resp.Diagnostics.AddError(
			"Error unmarshalling response",
			"Could not unmarshal response body: "+err.Error(),
		)
		return
	}

	// Handle wrapped response if necessary, but based on private_resources.go it seems to return the object directly or wrapped?
	// private_resources.go CreatePrivateResourceGroup unmarshals into PrivateResourceGroup.
	// Let's assume it returns the object.

	var createdGroup struct {
		ID          int    `json:"resourceGroupId"`
		Name        string `json:"name"`
		Description string `json:"description"`
		ResourceIDs []int  `json:"resourceIds"`
	}
	if err := json.Unmarshal(body, &createdGroup); err != nil {
		resp.Diagnostics.AddError("Error unmarshalling response", err.Error())
		return
	}

	data.ID = types.StringValue(strconv.Itoa(createdGroup.ID))
	data.Name = types.StringValue(createdGroup.Name)
	data.Description = types.StringValue(createdGroup.Description)

	var resIDs []types.Int64
	for _, id := range createdGroup.ResourceIDs {
		resIDs = append(resIDs, types.Int64Value(int64(id)))
	}
	data.ResourceIDs = resIDs

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PrivateResourceGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PrivateResourceGroupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	respHTTP, err := r.client.Query("policies", "privateResourceGroups/"+data.ID.ValueString(), "GET", nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading private resource group",
			"Could not read private resource group: "+err.Error(),
		)
		return
	}
	defer respHTTP.Body.Close()

	if respHTTP.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}

	if respHTTP.StatusCode != 200 {
		body, _ := io.ReadAll(respHTTP.Body)
		resp.Diagnostics.AddError(
			"Error reading private resource group",
			fmt.Sprintf("Status: %s, Body: %s", respHTTP.Status, string(body)),
		)
		return
	}

	body, _ := io.ReadAll(respHTTP.Body)
	var group struct {
		ID          int    `json:"resourceGroupId"`
		Name        string `json:"name"`
		Description string `json:"description"`
		ResourceIDs []int  `json:"resourceIds"`
	}
	if err := json.Unmarshal(body, &group); err != nil {
		resp.Diagnostics.AddError("Error unmarshalling response", err.Error())
		return
	}

	data.Name = types.StringValue(group.Name)
	data.Description = types.StringValue(group.Description)

	if len(group.ResourceIDs) > 0 {
		var resIDs []types.Int64
		for _, id := range group.ResourceIDs {
			resIDs = append(resIDs, types.Int64Value(int64(id)))
		}
		data.ResourceIDs = resIDs
	} else {
		data.ResourceIDs = nil
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PrivateResourceGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PrivateResourceGroupResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resourceIDs := make([]int, 0)
	for _, id := range data.ResourceIDs {
		resourceIDs = append(resourceIDs, int(id.ValueInt64()))
	}

	reqBody := map[string]interface{}{
		"name":        data.Name.ValueString(),
		"description": data.Description.ValueString(),
		"resourceIds": resourceIDs,
	}

	respHTTP, err := r.client.Query("policies", "privateResourceGroups/"+data.ID.ValueString(), "PUT", reqBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating private resource group",
			"Could not update private resource group: "+err.Error(),
		)
		return
	}
	defer respHTTP.Body.Close()

	if respHTTP.StatusCode != 200 {
		body, _ := io.ReadAll(respHTTP.Body)
		resp.Diagnostics.AddError(
			"Error updating private resource group",
			fmt.Sprintf("Status: %s, Body: %s", respHTTP.Status, string(body)),
		)
		return
	}

	body, _ := io.ReadAll(respHTTP.Body)
	var group struct {
		ID          int    `json:"resourceGroupId"`
		Name        string `json:"name"`
		Description string `json:"description"`
		ResourceIDs []int  `json:"resourceIds"`
	}
	if err := json.Unmarshal(body, &group); err != nil {
		resp.Diagnostics.AddError("Error unmarshalling response", err.Error())
		return
	}

	data.Name = types.StringValue(group.Name)
	data.Description = types.StringValue(group.Description)

	var resIDs []types.Int64
	for _, id := range group.ResourceIDs {
		resIDs = append(resIDs, types.Int64Value(int64(id)))
	}
	data.ResourceIDs = resIDs

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PrivateResourceGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PrivateResourceGroupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	respHTTP, err := r.client.Query("policies", "privateResourceGroups/"+data.ID.ValueString(), "DELETE", nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting private resource group",
			"Could not delete private resource group: "+err.Error(),
		)
		return
	}
	defer respHTTP.Body.Close()

	if respHTTP.StatusCode != 200 && respHTTP.StatusCode != 204 {
		body, _ := io.ReadAll(respHTTP.Body)
		resp.Diagnostics.AddError(
			"Error deleting private resource group",
			fmt.Sprintf("Status: %s, Body: %s", respHTTP.Status, string(body)),
		)
		return
	}
}

func (r *PrivateResourceGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id := req.ID
	if _, err := strconv.ParseInt(id, 10, 64); err != nil {
		// Not a number, try to find by name
		foundID, err := apiclient.GetPrivateResourceGroupIDByName(r.client, id)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error importing private resource group",
				fmt.Sprintf("Could not find private resource group with name '%s': %s", id, err.Error()),
			)
			return
		}
		req.ID = fmt.Sprintf("%d", foundID)
	}
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
