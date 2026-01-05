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

var _ datasource.DataSource = &ContentCategoryListsDataSource{}

func NewContentCategoryListsDataSource() datasource.DataSource {
	return &ContentCategoryListsDataSource{}
}

type ContentCategoryListsDataSource struct {
	client *apiclient.APIClient
}

type ContentCategoryListsDataSourceModel struct {
	ContentCategoryLists []ContentCategoryListModel `tfsdk:"content_category_lists"`
}

type ContentCategoryListModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (d *ContentCategoryListsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_content_category_lists"
}

func (d *ContentCategoryListsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of Content Category Lists.",
		Attributes: map[string]schema.Attribute{
			"content_category_lists": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *ContentCategoryListsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ContentCategoryListsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ContentCategoryListsDataSourceModel

	categories, err := d.client.GetContentCategories()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Content Category Lists",
			err.Error(),
		)
		return
	}

	for _, cat := range categories {
		state.ContentCategoryLists = append(state.ContentCategoryLists, ContentCategoryListModel{
			ID:   types.Int64Value(int64(cat.ID)),
			Name: types.StringValue(cat.Name),
		})
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
