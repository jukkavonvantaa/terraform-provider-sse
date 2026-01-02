package provider

import (
	"context"
	"fmt"

	"github.com/cisco/terraform-provider-sse/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &SecurityProfileDataSource{}

func NewSecurityProfileDataSource() datasource.DataSource {
	return &SecurityProfileDataSource{}
}

type SecurityProfileDataSource struct {
	client *apiclient.APIClient
}

type SecurityProfileDataSourceModel struct {
	ID        types.Int64  `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	IsDefault types.Bool   `tfsdk:"is_default"`
	Priority  types.Int64  `tfsdk:"priority"`
}

func (d *SecurityProfileDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_profile"
}

func (d *SecurityProfileDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single Security Profile by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the security profile to find.",
			},
			"is_default": schema.BoolAttribute{
				Computed: true,
			},
			"priority": schema.Int64Attribute{
				Computed: true,
			},
		},
	}
}

func (d *SecurityProfileDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SecurityProfileDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state SecurityProfileDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	targetName := state.Name.ValueString()

	profiles, err := d.client.GetSecurityProfiles()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Security Profiles",
			err.Error(),
		)
		return
	}

	var foundProfile *apiclient.SecurityProfile
	for _, profile := range profiles {
		if profile.Name == targetName {
			foundProfile = &profile
			break
		}
	}

	if foundProfile == nil {
		resp.Diagnostics.AddError(
			"Security Profile Not Found",
			fmt.Sprintf("No security profile found with name '%s'", targetName),
		)
		return
	}

	state.ID = types.Int64Value(int64(foundProfile.ID))
	state.Name = types.StringValue(foundProfile.Name)
	state.IsDefault = types.BoolValue(foundProfile.IsDefault)
	state.Priority = types.Int64Value(int64(foundProfile.Priority))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
