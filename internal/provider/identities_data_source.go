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
var _ datasource.DataSource = &IdentitiesDataSource{}

func NewIdentitiesDataSource() datasource.DataSource {
	return &IdentitiesDataSource{}
}

// IdentitiesDataSource defines the data source implementation.
type IdentitiesDataSource struct {
	client *apiclient.APIClient
}

// IdentitiesDataSourceModel describes the data source data model.
type IdentitiesDataSourceModel struct {
	Identities []IdentityModel `tfsdk:"identities"`
}

type IdentityModel struct {
	ID      types.Int64  `tfsdk:"id"`
	Label   types.String `tfsdk:"label"`
	Type    types.String `tfsdk:"type"`
	Deleted types.Bool   `tfsdk:"deleted"`
}

func (d *IdentitiesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_identities"
}

func (d *IdentitiesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Identities data source",

		Attributes: map[string]schema.Attribute{
			"identities": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "The ID of the identity.",
							Computed:            true,
						},
						"label": schema.StringAttribute{
							MarkdownDescription: "The descriptive label for the identity.",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "The type of the identity.",
							Computed:            true,
						},
						"deleted": schema.BoolAttribute{
							MarkdownDescription: "Indicates whether the identity was deleted.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *IdentitiesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *IdentitiesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data IdentitiesDataSourceModel

	// Read API call logic
	endpoint := "identities?limit=100&offset=0"
	httpResp, err := d.client.Query(apiclient.ScopeReports, endpoint, apiclient.OperationGet, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read identities, got error: %s", err),
		)
		return
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read response body, got error: %s", err),
		)
		return
	}

	if httpResp.StatusCode != 200 {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read identities, got status code: %d, body: %s", httpResp.StatusCode, string(body)),
		)
		return
	}

	// Define struct to match API response
	type IdentityType struct {
		ID    int    `json:"id"`
		Type  string `json:"type"`
		Label string `json:"label"`
	}

	type Identity struct {
		ID      int64        `json:"id"`
		Label   string       `json:"label"`
		Type    IdentityType `json:"type"`
		Deleted bool         `json:"deleted"`
	}

	type IdentityList struct {
		Data []Identity `json:"data"`
	}

	var identityList IdentityList
	if err := json.Unmarshal(body, &identityList); err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to unmarshal response body, got error: %s", err),
		)
		return
	}

	// Map response to state
	for _, identity := range identityList.Data {
		data.Identities = append(data.Identities, IdentityModel{
			ID:      types.Int64Value(identity.ID),
			Label:   types.StringValue(identity.Label),
			Type:    types.StringValue(identity.Type.Type),
			Deleted: types.BoolValue(identity.Deleted),
		})
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
