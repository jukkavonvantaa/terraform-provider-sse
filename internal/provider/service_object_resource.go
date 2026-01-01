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
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ServiceObjectResource{}
var _ resource.ResourceWithImportState = &ServiceObjectResource{}

func NewServiceObjectResource() resource.Resource {
	return &ServiceObjectResource{}
}

// ServiceObjectResource defines the resource implementation.
type ServiceObjectResource struct {
	client *apiclient.APIClient
}

// ServiceObjectResourceModel describes the resource data model.
type ServiceObjectResourceModel struct {
	ID          types.String   `tfsdk:"id"`
	Name        types.String   `tfsdk:"name"`
	Description types.String   `tfsdk:"description"`
	Protocol    types.String   `tfsdk:"protocol"`
	Ports       []types.String `tfsdk:"ports"`
}

func (r *ServiceObjectResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_object"
}

func (r *ServiceObjectResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Service Object resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Service Object ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Service Object Name",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Service Object Description",
			},
			"protocol": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Protocol (e.g., tcp, udp, icmp)",
			},
			"ports": schema.ListAttribute{
				Required:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "List of ports or port ranges",
			},
		},
	}
}

func (r *ServiceObjectResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ServiceObjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ServiceObjectResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var ports []string
	for _, p := range data.Ports {
		ports = append(ports, p.ValueString())
	}

	payload := apiclient.CreateServiceObjectPayload{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Value: apiclient.ServiceObjectValue{
			Protocol: data.Protocol.ValueString(),
			Ports:    ports,
		},
	}

	object, err := apiclient.CreateServiceObject(r.client, payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating service object",
			"Could not create service object, unexpected error: "+err.Error(),
		)
		return
	}

	data.ID = types.StringValue(strconv.FormatInt(object.ID, 10))
	// Ensure we write back what we got, although for input fields we usually keep plan values unless computed
	// But here we can confirm creation success.

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceObjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ServiceObjectResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(data.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", "Could not parse ID")
		return
	}

	object, err := apiclient.GetServiceObjectDetails(r.client, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading service object",
			"Could not read service object ID "+data.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	data.Name = types.StringValue(object.Name)
	data.Description = types.StringValue(object.Description)
	data.Protocol = types.StringValue(object.Value.Protocol)

	var ports []types.String
	for _, p := range object.Value.Ports {
		ports = append(ports, types.StringValue(p))
	}
	data.Ports = ports

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceObjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ServiceObjectResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(data.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", "Could not parse ID")
		return
	}

	var ports []string
	for _, p := range data.Ports {
		ports = append(ports, p.ValueString())
	}

	payload := apiclient.UpdateServiceObjectPayload{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Value: apiclient.ServiceObjectValue{
			Protocol: data.Protocol.ValueString(),
			Ports:    ports,
		},
	}

	_, err = apiclient.UpdateServiceObject(r.client, id, payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating service object",
			"Could not update service object: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceObjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ServiceObjectResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(data.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", "Could not parse ID")
		return
	}

	err = apiclient.DeleteServiceObject(r.client, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting service object",
			"Could not delete service object ID "+data.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}

func (r *ServiceObjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id := req.ID
	if _, err := strconv.ParseInt(id, 10, 64); err != nil {
		// Not a number, try to find by name
		foundID, err := apiclient.GetServiceObjectIDByName(r.client, id)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error importing service object",
				fmt.Sprintf("Could not find service object with name '%s': %s", id, err.Error()),
			)
			return
		}
		req.ID = fmt.Sprintf("%d", foundID)
	}
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
