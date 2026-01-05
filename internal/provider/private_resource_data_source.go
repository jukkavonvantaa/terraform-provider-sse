// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/cisco/terraform-provider-sse/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &PrivateResourceDataSource{}

func NewPrivateResourceDataSource() datasource.DataSource {
	return &PrivateResourceDataSource{}
}

type PrivateResourceDataSource struct {
	client *apiclient.APIClient
}

type PrivateResourceDataSourceModel struct {
	ID               types.Int64  `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	DNSServerID      types.Int64  `tfsdk:"dns_server_id"`
	CertificateID    types.Int64  `tfsdk:"certificate_id"`
	ResourceGroupIDs types.List   `tfsdk:"resource_group_ids"`
}

func (d *PrivateResourceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_private_resource"
}

func (d *PrivateResourceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single Private Resource by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of the private resource.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the private resource to find.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "The description of the private resource.",
			},
			"dns_server_id": schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of the DNS server.",
			},
			"certificate_id": schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of the certificate.",
			},
			"resource_group_ids": schema.ListAttribute{
				Computed:    true,
				ElementType: types.Int64Type,
				Description: "The list of resource group IDs.",
			},
		},
	}
}

func (d *PrivateResourceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PrivateResourceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state PrivateResourceDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	targetName := state.Name.ValueString()

	// 1. Find ID by Name
	id, err := apiclient.GetPrivateResourceIDByName(d.client, targetName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Find Private Resource",
			fmt.Sprintf("Could not find private resource with name '%s': %s", targetName, err.Error()),
		)
		return
	}

	// 2. Get Full Details
	endpoint := fmt.Sprintf("privateResources/%d", id)
	respHTTP, err := d.client.Query("policies", endpoint, "GET", nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Private Resource",
			err.Error(),
		)
		return
	}
	defer respHTTP.Body.Close()

	if respHTTP.StatusCode != 200 {
		body, _ := io.ReadAll(respHTTP.Body)
		resp.Diagnostics.AddError(
			"Unable to Read Private Resource",
			fmt.Sprintf("Status: %s, Body: %s", respHTTP.Status, string(body)),
		)
		return
	}

	body, _ := io.ReadAll(respHTTP.Body)

	var resourceObj struct {
		ID               int    `json:"resourceId"`
		Name             string `json:"name"`
		Description      string `json:"description"`
		DNSServerID      int    `json:"dnsServerId"`
		CertificateID    int    `json:"certificateId"`
		ResourceGroupIDs []int  `json:"resourceGroupIds"`
	}

	if err := json.Unmarshal(body, &resourceObj); err != nil {
		resp.Diagnostics.AddError("Error unmarshalling response", err.Error())
		return
	}

	state.ID = types.Int64Value(int64(resourceObj.ID))
	state.Name = types.StringValue(resourceObj.Name)
	state.Description = types.StringValue(resourceObj.Description)

	if resourceObj.DNSServerID != 0 {
		state.DNSServerID = types.Int64Value(int64(resourceObj.DNSServerID))
	} else {
		state.DNSServerID = types.Int64Null()
	}

	if resourceObj.CertificateID != 0 {
		state.CertificateID = types.Int64Value(int64(resourceObj.CertificateID))
	} else {
		state.CertificateID = types.Int64Null()
	}

	if len(resourceObj.ResourceGroupIDs) > 0 {
		var ids []types.Int64
		for _, id := range resourceObj.ResourceGroupIDs {
			ids = append(ids, types.Int64Value(int64(id)))
		}
		state.ResourceGroupIDs, _ = types.ListValueFrom(ctx, types.Int64Type, ids)
	} else {
		state.ResourceGroupIDs = types.ListNull(types.Int64Type)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
