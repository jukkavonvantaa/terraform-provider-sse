package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cisco/terraform-provider-sse/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ConnectorGroupResource{}
var _ resource.ResourceWithImportState = &ConnectorGroupResource{}

func NewConnectorGroupResource() resource.Resource {
	return &ConnectorGroupResource{}
}

type ConnectorGroupResource struct {
	client *apiclient.APIClient
}

type ConnectorGroupResourceModel struct {
	ID                       types.Int64  `tfsdk:"id"`
	Name                     types.String `tfsdk:"name"`
	Location                 types.String `tfsdk:"location"`
	Environment              types.String `tfsdk:"environment"`
	ProvisioningKey          types.String `tfsdk:"provisioning_key"`
	ProvisioningKeyExpiresAt types.String `tfsdk:"provisioning_key_expires_at"`
	BaseImageDownloadURL     types.String `tfsdk:"base_image_download_url"`
	Status                   types.String `tfsdk:"status"`
}

func (r *ConnectorGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connector_group"
}

func (r *ConnectorGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Resource Connector Group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of the Connector Group.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the Connector Group.",
			},
			"location": schema.StringAttribute{
				Required:    true,
				Description: "The region where the Resource Connector Group is available (e.g., us-west-2).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"environment": schema.StringAttribute{
				Required:    true,
				Description: "The type of cloud-native runtime environment (e.g., aws, azure, container, esx).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"provisioning_key": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The provisioning key for the Connector Group.",
			},
			"provisioning_key_expires_at": schema.StringAttribute{
				Computed:    true,
				Description: "The expiration time of the provisioning key.",
			},
			"base_image_download_url": schema.StringAttribute{
				Computed:    true,
				Description: "The URL to download the base image.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The status of the Connector Group.",
			},
		},
	}
}

func (r *ConnectorGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ConnectorGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ConnectorGroupResourceModel

	// Read Terraform plan data into the model
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create API request
	apiReq := apiclient.ConnectorGroupCreateRequest{
		Name:        plan.Name.ValueString(),
		Location:    plan.Location.ValueString(),
		Environment: plan.Environment.ValueString(),
	}

	// Call API
	group, err := r.client.CreateConnectorGroup(apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Connector Group",
			"Could not create connector group, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response to model
	plan.ID = types.Int64Value(int64(group.ID))
	plan.Name = types.StringValue(group.Name)
	plan.Location = types.StringValue(group.Location)
	plan.Environment = types.StringValue(group.Environment)
	plan.ProvisioningKey = types.StringValue(group.ProvisioningKey)
	plan.ProvisioningKeyExpiresAt = types.StringValue(group.ProvisioningKeyExpiresAt)
	plan.BaseImageDownloadURL = types.StringValue(group.BaseImageDownloadURL)
	plan.Status = types.StringValue(group.Status)

	// Save data into Terraform state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *ConnectorGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ConnectorGroupResourceModel

	// Read Terraform prior state data into the model
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get ID from state
	id := int(state.ID.ValueInt64())

	// Call API
	group, err := r.client.GetConnectorGroup(id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Connector Group",
			"Could not read connector group ID "+strconv.Itoa(id)+": "+err.Error(),
		)
		return
	}

	if group == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Map response to model
	state.Name = types.StringValue(group.Name)
	state.Location = types.StringValue(group.Location)
	state.Environment = types.StringValue(group.Environment)
	state.ProvisioningKey = types.StringValue(group.ProvisioningKey)
	state.ProvisioningKeyExpiresAt = types.StringValue(group.ProvisioningKeyExpiresAt)
	state.BaseImageDownloadURL = types.StringValue(group.BaseImageDownloadURL)
	state.Status = types.StringValue(group.Status)

	// Save updated data into Terraform state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *ConnectorGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ConnectorGroupResourceModel

	// Read Terraform plan data into the model
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get ID from state
	id := int(plan.ID.ValueInt64())

	// Create API request
	apiReq := apiclient.ConnectorGroupUpdateRequest{
		Name:     plan.Name.ValueString(),
		Location: plan.Location.ValueString(),
	}

	// Call API
	group, err := r.client.UpdateConnectorGroup(id, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Connector Group",
			"Could not update connector group ID "+strconv.Itoa(id)+": "+err.Error(),
		)
		return
	}

	// Map response to model
	plan.Name = types.StringValue(group.Name)
	plan.Location = types.StringValue(group.Location)
	// Environment is not updated by API, but we keep it from plan
	plan.Environment = types.StringValue(group.Environment)
	plan.ProvisioningKey = types.StringValue(group.ProvisioningKey)
	plan.ProvisioningKeyExpiresAt = types.StringValue(group.ProvisioningKeyExpiresAt)
	plan.BaseImageDownloadURL = types.StringValue(group.BaseImageDownloadURL)
	plan.Status = types.StringValue(group.Status)

	// Save updated data into Terraform state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *ConnectorGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ConnectorGroupResourceModel

	// Read Terraform prior state data into the model
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get ID from state
	id := int(state.ID.ValueInt64())

	// Call API
	err := r.client.DeleteConnectorGroup(id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Connector Group",
			"Could not delete connector group ID "+strconv.Itoa(id)+": "+err.Error(),
		)
		return
	}
}

func (r *ConnectorGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Try to parse the ID as an integer first
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		// If it's not an integer, assume it's a name and try to look it up
		group, err := r.client.GetConnectorGroupByName(req.ID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error importing Connector Group by name",
				fmt.Sprintf("Could not find connector group with name %q: %s", req.ID, err),
			)
			return
		}
		if group == nil {
			resp.Diagnostics.AddError(
				"Connector Group Not Found",
				fmt.Sprintf("No connector group found with name %q", req.ID),
			)
			return
		}
		id = int64(group.ID)
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
