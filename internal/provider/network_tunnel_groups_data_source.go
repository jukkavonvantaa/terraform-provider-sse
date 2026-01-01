// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/cisco/terraform-provider-sse/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &NetworkTunnelGroupsDataSource{}

func NewNetworkTunnelGroupsDataSource() datasource.DataSource {
	return &NetworkTunnelGroupsDataSource{}
}

// NetworkTunnelGroupsDataSource defines the data source implementation.
type NetworkTunnelGroupsDataSource struct {
	client *apiclient.APIClient
}

// NetworkTunnelGroupsDataSourceModel describes the data source data model.
type NetworkTunnelGroupsDataSourceModel struct {
	NetworkTunnelGroups []NetworkTunnelGroupModel `tfsdk:"network_tunnel_groups"`
}

type NetworkTunnelGroupModel struct {
	ID             types.Int64  `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	OrganizationID types.Int64  `tfsdk:"organization_id"`
	DeviceType     types.String `tfsdk:"device_type"`
	Region         types.String `tfsdk:"region"`
	Status         types.String `tfsdk:"status"`
}

func (d *NetworkTunnelGroupsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_tunnel_groups"
}

func (d *NetworkTunnelGroupsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Network Tunnel Groups data source",

		Attributes: map[string]schema.Attribute{
			"network_tunnel_groups": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"organization_id": schema.Int64Attribute{
							Computed: true,
						},
						"device_type": schema.StringAttribute{
							Computed: true,
						},
						"region": schema.StringAttribute{
							Computed: true,
						},
						"status": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *NetworkTunnelGroupsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*apiclient.APIClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *apiclient.APIClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *NetworkTunnelGroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NetworkTunnelGroupsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	groups, err := apiclient.GetNetworkTunnelGroups(d.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Network Tunnel Groups",
			err.Error(),
		)
		return
	}

	for _, group := range groups {
		data.NetworkTunnelGroups = append(data.NetworkTunnelGroups, NetworkTunnelGroupModel{
			ID:             types.Int64Value(group.ID),
			Name:           types.StringValue(group.Name),
			OrganizationID: types.Int64Value(group.OrganizationID),
			DeviceType:     types.StringValue(group.DeviceType),
			Region:         types.StringValue(group.Region),
			Status:         types.StringValue(group.Status),
		})
	}

	// Write logs using the tflog package
	// tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
