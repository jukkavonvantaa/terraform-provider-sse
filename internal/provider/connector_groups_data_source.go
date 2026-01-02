package provider

import (
	"context"
	"fmt"

	"github.com/cisco/terraform-provider-sse/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ConnectorGroupsDataSource{}

func NewConnectorGroupsDataSource() datasource.DataSource {
	return &ConnectorGroupsDataSource{}
}

type ConnectorGroupsDataSource struct {
	client *apiclient.APIClient
}

type ConnectorGroupsDataSourceModel struct {
	ConnectorGroups []ConnectorGroupModel `tfsdk:"connector_groups"`
}

type ConnectorGroupModel struct {
	ID                          types.Int64  `tfsdk:"id"`
	Name                        types.String `tfsdk:"name"`
	Location                    types.String `tfsdk:"location"`
	Environment                 types.String `tfsdk:"environment"`
	Status                      types.String `tfsdk:"status"`
	ConnectorsCount             types.Int64  `tfsdk:"connectors_count"`
	ConnectedConnectorsCount    types.Int64  `tfsdk:"connected_connectors_count"`
	DisconnectedConnectorsCount types.Int64  `tfsdk:"disconnected_connectors_count"`
}

func (d *ConnectorGroupsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connector_groups"
}

func (d *ConnectorGroupsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of Resource Connector Groups.",
		Attributes: map[string]schema.Attribute{
			"connector_groups": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"location": schema.StringAttribute{
							Computed: true,
						},
						"environment": schema.StringAttribute{
							Computed: true,
						},
						"status": schema.StringAttribute{
							Computed: true,
						},
						"connectors_count": schema.Int64Attribute{
							Computed: true,
						},
						"connected_connectors_count": schema.Int64Attribute{
							Computed: true,
						},
						"disconnected_connectors_count": schema.Int64Attribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *ConnectorGroupsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ConnectorGroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ConnectorGroupsDataSourceModel

	// Fetch all connector groups
	// The API supports pagination, so we need to loop until we get all of them.
	var allGroups []apiclient.ConnectorGroup
	limit := 100
	offset := 0

	for {
		groups, err := d.client.GetConnectorGroups(limit, offset)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Read Connector Groups",
				err.Error(),
			)
			return
		}

		allGroups = append(allGroups, groups...)

		if len(groups) < limit {
			break
		}
		offset += limit
	}

	// Map response body to model
	for _, group := range allGroups {
		state.ConnectorGroups = append(state.ConnectorGroups, ConnectorGroupModel{
			ID:                          types.Int64Value(int64(group.ID)),
			Name:                        types.StringValue(group.Name),
			Location:                    types.StringValue(group.Location),
			Environment:                 types.StringValue(group.Environment),
			Status:                      types.StringValue(group.Status),
			ConnectorsCount:             types.Int64Value(int64(group.ConnectorsCount)),
			ConnectedConnectorsCount:    types.Int64Value(int64(group.ConnectedConnectorsCount)),
			DisconnectedConnectorsCount: types.Int64Value(int64(group.DisconnectedConnectorsCount)),
		})
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
