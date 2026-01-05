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

var _ datasource.DataSource = &IPSProfilesDataSource{}

func NewIPSProfilesDataSource() datasource.DataSource {
	return &IPSProfilesDataSource{}
}

type IPSProfilesDataSource struct {
	client *apiclient.APIClient
}

type IPSProfilesDataSourceModel struct {
	IPSProfiles []IPSProfileModel `tfsdk:"ips_profiles"`
}

type IPSProfileModel struct {
	ID         types.Int64  `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	IsDefault  types.Bool   `tfsdk:"is_default"`
	SystemMode types.String `tfsdk:"system_mode"`
}

func (d *IPSProfilesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ips_profiles"
}

func (d *IPSProfilesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of IPS Profiles.",
		Attributes: map[string]schema.Attribute{
			"ips_profiles": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"is_default": schema.BoolAttribute{
							Computed: true,
						},
						"system_mode": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *IPSProfilesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *IPSProfilesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state IPSProfilesDataSourceModel

	profiles, err := d.client.GetIPSProfiles()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read IPS Profiles",
			err.Error(),
		)
		return
	}

	for _, profile := range profiles {
		state.IPSProfiles = append(state.IPSProfiles, IPSProfileModel{
			ID:         types.Int64Value(int64(profile.ID)),
			Name:       types.StringValue(profile.Name),
			IsDefault:  types.BoolValue(profile.IsDefault),
			SystemMode: types.StringValue(profile.SystemMode),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
