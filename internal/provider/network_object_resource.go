// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cisco/terraform-provider-sse/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &NetworkObjectResource{}
var _ resource.ResourceWithImportState = &NetworkObjectResource{}

func NewNetworkObjectResource() resource.Resource {
	return &NetworkObjectResource{}
}

// NetworkObjectResource defines the resource implementation.
type NetworkObjectResource struct {
	client *apiclient.APIClient
}

// NetworkObjectResourceModel describes the resource data model.
type NetworkObjectResourceModel struct {
	ID          types.String   `tfsdk:"id"`
	ObjectID    types.Int64    `tfsdk:"object_id"`
	Name        types.String   `tfsdk:"name"`
	Description types.String   `tfsdk:"description"`
	Type        types.String   `tfsdk:"type"`
	Addresses   []types.String `tfsdk:"addresses"`
}

func (r *NetworkObjectResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_object"
}

func (r *NetworkObjectResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Network Object resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Network Object ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"object_id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Network Object ID (Integer), useful for JSON encoding in rules.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Network Object Name",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Network Object Description",
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Network Object Type (e.g., host, network, range, fqdn)",
			},
			"addresses": schema.ListAttribute{
				Required:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "List of addresses",
			},
		},
	}
}

func (r *NetworkObjectResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
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

func (r *NetworkObjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NetworkObjectResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var addresses []string
	for _, addr := range data.Addresses {
		addresses = append(addresses, addr.ValueString())
	}

	value := apiclient.NetworkObjectValue{
		Type:      data.Type.ValueString(),
		Addresses: addresses,
	}

	id, err := apiclient.PostNetworkObject(r.client, data.Name.ValueString(), value, data.Description.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating network object",
			"Could not create network object, unexpected error: "+err.Error(),
		)
		return
	}

	data.ID = types.StringValue(id)
	if idInt, err := strconv.ParseInt(id, 10, 64); err == nil {
		data.ObjectID = types.Int64Value(idInt)
	} else {
		// If ID is not an integer, set to null or handle accordingly
		// For now, assuming IDs are numeric strings
		data.ObjectID = types.Int64Null()
	}

	// Ensure description is known (if it was computed/optional and not set)
	if data.Description.IsUnknown() {
		data.Description = types.StringValue(data.Description.ValueString())
	}

	// Write logs using the tflog package
	// tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NetworkObjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NetworkObjectResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	obj, err := apiclient.GetNetworkObjectDetails(r.client, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading network object",
			"Could not read network object ID "+data.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	if obj == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Map response body to model
	objMap, ok := obj.(map[string]interface{})
	if !ok {
		resp.Diagnostics.AddError("Error reading network object", "Unexpected response format")
		return
	}

	if id, ok := objMap["id"].(string); ok {
		if idInt, err := strconv.ParseInt(id, 10, 64); err == nil {
			data.ObjectID = types.Int64Value(idInt)
		}
	} else if id, ok := objMap["id"].(float64); ok {
		data.ObjectID = types.Int64Value(int64(id))
	}

	if name, ok := objMap["name"].(string); ok {
		tflog.Debug(ctx, "Read name from API", map[string]interface{}{"api_name": name})
		data.Name = types.StringValue(name)
	}
	if desc, ok := objMap["description"].(string); ok {
		data.Description = types.StringValue(desc)
	}
	if val, ok := objMap["value"].(map[string]interface{}); ok {
		if t, ok := val["type"].(string); ok {
			data.Type = types.StringValue(t)
		}
		if addrs, ok := val["addresses"].([]interface{}); ok {
			var addresses []types.String
			for _, addr := range addrs {
				if s, ok := addr.(string); ok {
					addresses = append(addresses, types.StringValue(s))
				}
			}
			data.Addresses = addresses
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NetworkObjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data NetworkObjectResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var addresses []string
	for _, addr := range data.Addresses {
		addresses = append(addresses, addr.ValueString())
	}

	value := apiclient.NetworkObjectValue{
		Type:      data.Type.ValueString(),
		Addresses: addresses,
	}

	err := apiclient.PutNetworkObjectDetails(r.client, data.ID.ValueString(), data.Name.ValueString(), value, data.Description.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating network object",
			"Could not update network object, unexpected error: "+err.Error(),
		)
		return
	}

	// Ensure description is known (if it was computed/optional and not set)
	if data.Description.IsUnknown() {
		data.Description = types.StringValue(data.Description.ValueString())
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NetworkObjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data NetworkObjectResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := apiclient.DeleteNetworkObject(r.client, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting network object",
			"Could not delete network object, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *NetworkObjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Try to determine if the ID is a name or an ID
	// If it's a name, look up the ID
	// If it's an ID, pass it through

	// Simple heuristic: if it contains non-numeric characters, treat as name (unless it's a UUID, but network object IDs seem to be integers)
	// Or just try to look it up by name first, if fails, assume it's an ID.
	// But looking up by name involves fetching all objects, which is expensive.
	// Let's try to parse as int.

	id := req.ID
	isNumeric := true
	for _, c := range id {
		if c < '0' || c > '9' {
			isNumeric = false
			break
		}
	}

	if !isNumeric {
		// Assume it's a name
		foundID, err := apiclient.GetNetworkObjectIDByName(r.client, id)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error importing network object",
				fmt.Sprintf("Could not find network object with name '%s': %s", id, err.Error()),
			)
			return
		}
		// resp.Diagnostics.AddInfo("Import", fmt.Sprintf("Resolved name '%s' to ID '%s'", id, foundID))
		req.ID = foundID
	}

	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
