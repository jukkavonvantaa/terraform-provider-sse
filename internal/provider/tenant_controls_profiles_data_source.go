package provider

import (
	"context"
	"fmt"

	"github.com/cisco/terraform-provider-sse/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource = &TenantControlsProfilesDataSource{}
	_ datasource.DataSource = &TenantControlsProfileDataSource{}
)

func NewTenantControlsProfilesDataSource() datasource.DataSource {
	return &TenantControlsProfilesDataSource{}
}

type TenantControlsProfilesDataSource struct {
	client *apiclient.APIClient
}

type TenantControlsProfilesDataSourceModel struct {
	Profiles []TenantControlsProfileModel `tfsdk:"profiles"`
}

type TenantControlsProfileModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	IsDefault   types.Bool   `tfsdk:"is_default"`
}

func (d *TenantControlsProfilesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant_controls_profiles"
}

func (d *TenantControlsProfilesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all Tenant Controls Profiles.",
		Attributes: map[string]schema.Attribute{
			"profiles": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed:    true,
							Description: "The ID of the Tenant Controls Profile.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the Tenant Controls Profile.",
						},
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "The description of the Tenant Controls Profile.",
						},
						"is_default": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether this is the default profile.",
						},
					},
				},
			},
		},
	}
}

func (d *TenantControlsProfilesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TenantControlsProfilesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state TenantControlsProfilesDataSourceModel

	profiles, err := apiclient.GetTenantControlsProfiles(d.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Tenant Controls Profiles",
			err.Error(),
		)
		return
	}

	for _, profile := range profiles {
		state.Profiles = append(state.Profiles, TenantControlsProfileModel{
			ID:          types.Int64Value(profile.ID),
			Name:        types.StringValue(profile.Name),
			Description: types.StringValue(profile.Description),
			IsDefault:   types.BoolValue(profile.IsDefault),
		})
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Single Data Source

func NewTenantControlsProfileDataSource() datasource.DataSource {
	return &TenantControlsProfileDataSource{}
}

type TenantControlsProfileDataSource struct {
	client *apiclient.APIClient
}

type TenantControlsProfileDataSourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	IsDefault   types.Bool   `tfsdk:"is_default"`
}

func (d *TenantControlsProfileDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant_controls_profile"
}

func (d *TenantControlsProfileDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single Tenant Controls Profile by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of the Tenant Controls Profile.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the Tenant Controls Profile to fetch.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "The description of the Tenant Controls Profile.",
			},
			"is_default": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether this is the default profile.",
			},
		},
	}
}

func (d *TenantControlsProfileDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TenantControlsProfileDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state TenantControlsProfileDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	profiles, err := apiclient.GetTenantControlsProfiles(d.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Tenant Controls Profiles",
			err.Error(),
		)
		return
	}

	var foundProfile *apiclient.TenantControlsProfile
	for _, p := range profiles {
		if p.Name == name {
			foundProfile = &p
			break
		}
	}

	if foundProfile == nil {
		resp.Diagnostics.AddError(
			"Tenant Controls Profile Not Found",
			fmt.Sprintf("No Tenant Controls Profile found with name '%s'", name),
		)
		return
	}

	state.ID = types.Int64Value(foundProfile.ID)
	state.Description = types.StringValue(foundProfile.Description)
	state.IsDefault = types.BoolValue(foundProfile.IsDefault)

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
