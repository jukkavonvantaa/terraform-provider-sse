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

var _ datasource.DataSource = &SecurityProfilesDataSource{}

func NewSecurityProfilesDataSource() datasource.DataSource {
	return &SecurityProfilesDataSource{}
}

type SecurityProfilesDataSource struct {
	client *apiclient.APIClient
}

type SecurityProfilesDataSourceModel struct {
	SecurityProfiles []SecurityProfileModel `tfsdk:"security_profiles"`
}

type SecurityProfileModel struct {
	ID        types.Int64  `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	IsDefault types.Bool   `tfsdk:"is_default"`
	Priority  types.Int64  `tfsdk:"priority"`
}

func (d *SecurityProfilesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_profiles"
}

func (d *SecurityProfilesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of Security Profiles.",
		Attributes: map[string]schema.Attribute{
			"security_profiles": schema.ListNestedAttribute{
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
						"priority": schema.Int64Attribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *SecurityProfilesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SecurityProfilesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state SecurityProfilesDataSourceModel

	profiles, err := d.client.GetSecurityProfiles()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Security Profiles",
			err.Error(),
		)
		return
	}

	for _, profile := range profiles {
		state.SecurityProfiles = append(state.SecurityProfiles, SecurityProfileModel{
			ID:        types.Int64Value(int64(profile.ID)),
			Name:      types.StringValue(profile.Name),
			IsDefault: types.BoolValue(profile.IsDefault),
			Priority:  types.Int64Value(int64(profile.Priority)),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
