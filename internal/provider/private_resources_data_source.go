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

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &PrivateResourcesDataSource{}

func NewPrivateResourcesDataSource() datasource.DataSource {
	return &PrivateResourcesDataSource{}
}

// PrivateResourcesDataSource defines the data source implementation.
type PrivateResourcesDataSource struct {
	client *apiclient.APIClient
}

// PrivateResourcesDataSourceModel describes the data source data model.
type PrivateResourcesDataSourceModel struct {
	PrivateResources []PrivateResourceModel `tfsdk:"private_resources"`
}

type PrivateResourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (d *PrivateResourcesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_private_resources"
}

func (d *PrivateResourcesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Private Resources data source",

		Attributes: map[string]schema.Attribute{
			"private_resources": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "The ID of the private resource.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the private resource.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *PrivateResourcesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PrivateResourcesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PrivateResourcesDataSourceModel

	// Read API call logic
	offset := 0
	limit := 100
	var allResources []PrivateResourceModel

	for {
		endpoint := fmt.Sprintf("privateResources?limit=%d&offset=%d", limit, offset)
		respHTTP, err := d.client.Query("policies", endpoint, "GET", nil)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Read Private Resources",
				err.Error(),
			)
			return
		}

		if respHTTP.StatusCode != 200 {
			body, _ := io.ReadAll(respHTTP.Body)
			resp.Diagnostics.AddError(
				"Unable to Read Private Resources",
				fmt.Sprintf("Status: %s, Body: %s", respHTTP.Status, string(body)),
			)
			respHTTP.Body.Close()
			return
		}

		var result struct {
			Items []struct {
				ID   int    `json:"resourceId"`
				Name string `json:"name"`
			} `json:"items"`
			Total int `json:"total"`
		}

		if err := json.NewDecoder(respHTTP.Body).Decode(&result); err != nil {
			resp.Diagnostics.AddError(
				"Unable to Decode Private Resources Response",
				err.Error(),
			)
			respHTTP.Body.Close()
			return
		}
		respHTTP.Body.Close()

		for _, item := range result.Items {
			allResources = append(allResources, PrivateResourceModel{
				ID:   types.Int64Value(int64(item.ID)),
				Name: types.StringValue(item.Name),
			})
		}

		if len(result.Items) == 0 || offset+len(result.Items) >= result.Total {
			break
		}
		offset += len(result.Items)
	}

	data.PrivateResources = allResources

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
