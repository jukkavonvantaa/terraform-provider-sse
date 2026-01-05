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

var _ datasource.DataSource = &IdentityDataSource{}

func NewIdentityDataSource() datasource.DataSource {
	return &IdentityDataSource{}
}

type IdentityDataSource struct {
	client *apiclient.APIClient
}

type IdentityDataSourceModel struct {
	ID      types.Int64  `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Label   types.String `tfsdk:"label"`
	Type    types.String `tfsdk:"type"`
	Deleted types.Bool   `tfsdk:"deleted"`
}

func (d *IdentityDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_identity"
}

func (d *IdentityDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single Identity by name (label).",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name (label) of the identity to find.",
			},
			"label": schema.StringAttribute{
				Computed: true,
			},
			"type": schema.StringAttribute{
				Computed: true,
			},
			"deleted": schema.BoolAttribute{
				Computed: true,
			},
		},
	}
}

func (d *IdentityDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *IdentityDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state IdentityDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	targetName := state.Name.ValueString()

	identities, err := d.client.GetIdentities()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Identities",
			err.Error(),
		)
		return
	}

	var foundIdentity *apiclient.Identity
	for _, identity := range identities {
		if identity.Label == targetName {
			foundIdentity = &identity
			break
		}
	}

	if foundIdentity == nil {
		resp.Diagnostics.AddError(
			"Identity Not Found",
			fmt.Sprintf("No identity found with name '%s'", targetName),
		)
		return
	}

	state.ID = types.Int64Value(int64(foundIdentity.ID))
	state.Label = types.StringValue(foundIdentity.Label)
	state.Type = types.StringValue(foundIdentity.Type.Type)
	state.Deleted = types.BoolValue(foundIdentity.Deleted)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
