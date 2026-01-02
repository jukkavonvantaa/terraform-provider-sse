// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/cisco/terraform-provider-sse/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = &ScaffoldingProvider{}
var _ provider.ProviderWithFunctions = &ScaffoldingProvider{}
var _ provider.ProviderWithEphemeralResources = &ScaffoldingProvider{}
var _ provider.ProviderWithActions = &ScaffoldingProvider{}

// ScaffoldingProvider defines the provider implementation.
type ScaffoldingProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// ScaffoldingProviderModel describes the provider data model.
type ScaffoldingProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
}

func (p *ScaffoldingProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "sse"
	resp.Version = p.version
}

func (p *ScaffoldingProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Example provider attribute",
				Optional:            true,
			},
		},
	}
}

func (p *ScaffoldingProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ScaffoldingProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	// if data.Endpoint.IsNull() { /* ... */ }

	// Retrieve configuration from environment variables
	clientID := os.Getenv("SSE_CLIENT_KEY")
	if clientID == "" {
		// Fallback to old variable name for backward compatibility or user convenience
		clientID = os.Getenv("SSE_CLIENT_ID")
	}
	clientSecret := os.Getenv("SSE_CLIENT_SECRET")
	tokenURL := os.Getenv("SSE_TOKEN_URL")
	region := os.Getenv("SSE_REGION")
	if region == "" {
		region = "us"
	}

	if tokenURL == "" {
		tokenURL = "https://api.sse.cisco.com/auth/v2/token"
	}

	if clientID == "" || clientSecret == "" {
		resp.Diagnostics.AddError(
			"Missing Configuration",
			"SSE_CLIENT_KEY (or SSE_CLIENT_ID) and SSE_CLIENT_SECRET environment variables must be set.",
		)
		return
	}

	scopes := []string{
		"policies.destinationlists:read", "policies.destinationlists:write",
		"policies.objects.networkObjects:read", "policies.objects.networkObjects:write",
		"policies.securityProfiles:read",
		"policies.objects.serviceObjects:read", "policies.objects.serviceObjects:write",
		"policies.rules:read", "policies.rules:write",
		"policies.privateresources:read", "policies.privateresources:write",
		"policies.privateresourcegroups:read", "policies.privateresourcegroups:write",
		"deployments.privateresources:read", "deployments.privateresources:write",
		"deployments.identities:read",
		"deployments.networktunnelgroups:read",
		"reports.utilities:read",
		"admin.users:read",
		"deployments.roamingcomputers:read",
	}

	// Create the API client
	client := apiclient.NewAPIClient(tokenURL, clientID, clientSecret, scopes, region)
	if client == nil {
		resp.Diagnostics.AddError(
			"Client Creation Failed",
			"Failed to create API client. Check your configuration.",
		)
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *ScaffoldingProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewNetworkObjectResource,
		NewAccessRuleResource,
		NewDestinationListResource,
		NewServiceObjectResource,
		NewPrivateResourceGroupResource,
		NewPrivateResourceResource,
	}
}

func (p *ScaffoldingProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{}
}

func (p *ScaffoldingProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewNetworkTunnelGroupsDataSource,
		NewIdentitiesDataSource,
	}
}

func (p *ScaffoldingProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func (p *ScaffoldingProvider) Actions(ctx context.Context) []func() action.Action {
	return []func() action.Action{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ScaffoldingProvider{
			version: version,
		}
	}
}
