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

var _ datasource.DataSource = &IPSProfileDataSource{}

func NewIPSProfileDataSource() datasource.DataSource {
	return &IPSProfileDataSource{}
}

type IPSProfileDataSource struct {
	client *apiclient.APIClient
}

type IPSProfileDataSourceModel struct {
	ID         types.Int64  `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	IsDefault  types.Bool   `tfsdk:"is_default"`
	SystemMode types.String `tfsdk:"system_mode"`
}

func (d *IPSProfileDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ips_profile"
}

func (d *IPSProfileDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single IPS Profile by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the IPS profile to find.",
			},
			"is_default": schema.BoolAttribute{
				Computed: true,
			},
			"system_mode": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *IPSProfileDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *IPSProfileDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state IPSProfileDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	targetName := state.Name.ValueString()

	profiles, err := d.client.GetIPSProfiles()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read IPS Profiles",
			err.Error(),
		)
		return
	}

	var foundProfile *apiclient.IPSProfile
	for _, profile := range profiles {
		if profile.Name == targetName {
			foundProfile = &profile
			break
		}
	}

	if foundProfile == nil {
		resp.Diagnostics.AddError(
			"IPS Profile Not Found",
			fmt.Sprintf("No IPS profile found with name '%s'", targetName),
		)
		return
	}

	state.ID = types.Int64Value(int64(foundProfile.ID))
	state.Name = types.StringValue(foundProfile.Name)
	state.IsDefault = types.BoolValue(foundProfile.IsDefault)
	state.SystemMode = types.StringValue(foundProfile.SystemMode)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
