package provider

import (
	"context"
	"fmt"

	"github.com/cisco/terraform-provider-sse/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ApplicationDataSource{}

func NewApplicationDataSource() datasource.DataSource {
	return &ApplicationDataSource{}
}

type ApplicationDataSource struct {
	client *apiclient.APIClient
}

type ApplicationDataSourceModel struct {
	ID      types.Int64  `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Label   types.String `tfsdk:"label"`
	AppType types.String `tfsdk:"app_type"`
}

func (d *ApplicationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (d *ApplicationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single Application by name (label).",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name (label) of the application to find.",
			},
			"label": schema.StringAttribute{
				Computed: true,
			},
			"app_type": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *ApplicationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ApplicationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ApplicationDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	targetName := state.Name.ValueString()

	apps, err := d.client.GetApplications()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Applications",
			err.Error(),
		)
		return
	}

	var foundApp *apiclient.Application
	for _, app := range apps {
		if app.Label == targetName {
			foundApp = &app
			break
		}
	}

	if foundApp == nil {
		resp.Diagnostics.AddError(
			"Application Not Found",
			fmt.Sprintf("No application found with name '%s'", targetName),
		)
		return
	}

	state.ID = types.Int64Value(int64(foundApp.ID))
	state.Label = types.StringValue(foundApp.Label)
	state.AppType = types.StringValue(foundApp.Type)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
