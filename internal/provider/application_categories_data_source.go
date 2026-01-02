package provider

import (
	"context"
	"fmt"

	"github.com/cisco/terraform-provider-sse/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ApplicationCategoriesDataSource{}

func NewApplicationCategoriesDataSource() datasource.DataSource {
	return &ApplicationCategoriesDataSource{}
}

type ApplicationCategoriesDataSource struct {
	client *apiclient.APIClient
}

type ApplicationCategoriesDataSourceModel struct {
	Categories []ApplicationCategoryModel `tfsdk:"application_categories"`
}

type ApplicationCategoryModel struct {
	ID                types.Int64  `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	ApplicationsCount types.Int64  `tfsdk:"applications_count"`
}

func (d *ApplicationCategoriesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_categories"
}

func (d *ApplicationCategoriesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of Application Categories.",
		Attributes: map[string]schema.Attribute{
			"application_categories": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"description": schema.StringAttribute{
							Computed: true,
						},
						"applications_count": schema.Int64Attribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *ApplicationCategoriesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ApplicationCategoriesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ApplicationCategoriesDataSourceModel

	categories, err := d.client.GetApplicationCategories()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Application Categories",
			err.Error(),
		)
		return
	}

	for _, cat := range categories {
		state.Categories = append(state.Categories, ApplicationCategoryModel{
			ID:                types.Int64Value(int64(cat.ID)),
			Name:              types.StringValue(cat.Name),
			Description:       types.StringValue(cat.Description),
			ApplicationsCount: types.Int64Value(int64(cat.ApplicationsCount)),
		})
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
