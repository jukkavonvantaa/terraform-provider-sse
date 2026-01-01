// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/cisco/terraform-provider-sse/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &PrivateResourceResource{}
var _ resource.ResourceWithImportState = &PrivateResourceResource{}

func NewPrivateResourceResource() resource.Resource {
	return &PrivateResourceResource{}
}

type PrivateResourceResource struct {
	client *apiclient.APIClient
}

type PrivateResourceResourceModel struct {
	ID                types.String           `tfsdk:"id"`
	Name              types.String           `tfsdk:"name"`
	Description       types.String           `tfsdk:"description"`
	DNSServerID       types.Int64            `tfsdk:"dns_server_id"`
	CertificateID     types.Int64            `tfsdk:"certificate_id"`
	AccessTypes       []AccessTypeModel      `tfsdk:"access_types"`
	ResourceAddresses []ResourceAddressModel `tfsdk:"resource_addresses"`
	ResourceGroupIDs  []types.Int64          `tfsdk:"resource_group_ids"`
}

type AccessTypeModel struct {
	Type                   types.String   `tfsdk:"type"`
	ExternalFQDNPrefix     types.String   `tfsdk:"external_fqdn_prefix"`
	ExternalFQDN           types.String   `tfsdk:"external_fqdn"`
	Protocol               types.String   `tfsdk:"protocol"`
	SNI                    types.String   `tfsdk:"sni"`
	SSLVerificationEnabled types.Bool     `tfsdk:"ssl_verification_enabled"`
	ReachableAddresses     []types.String `tfsdk:"reachable_addresses"`
}

type ResourceAddressModel struct {
	DestinationAddr []types.String      `tfsdk:"destination_addr"`
	ProtocolPorts   []ProtocolPortModel `tfsdk:"protocol_ports"`
}

type ProtocolPortModel struct {
	Protocol types.String `tfsdk:"protocol"`
	Ports    types.String `tfsdk:"ports"`
}

func (r *PrivateResourceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_private_resource"
}

func (r *PrivateResourceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Private Resource resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Private Resource ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Private Resource Name",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Private Resource Description",
			},
			"dns_server_id": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "DNS Server ID",
			},
			"certificate_id": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Certificate ID",
			},
			"resource_group_ids": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.Int64Type,
				MarkdownDescription: "List of Resource Group IDs",
			},
		},
		Blocks: map[string]schema.Block{
			"access_types": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Required: true,
						},
						"external_fqdn_prefix": schema.StringAttribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"external_fqdn": schema.StringAttribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"protocol": schema.StringAttribute{
							Optional: true,
						},
						"sni": schema.StringAttribute{
							Optional: true,
						},
						"ssl_verification_enabled": schema.BoolAttribute{
							Optional: true,
						},
						"reachable_addresses": schema.ListAttribute{
							Optional:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
			"resource_addresses": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"destination_addr": schema.ListAttribute{
							Required:    true,
							ElementType: types.StringType,
						},
					},
					Blocks: map[string]schema.Block{
						"protocol_ports": schema.ListNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"protocol": schema.StringAttribute{
										Required: true,
									},
									"ports": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *PrivateResourceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*apiclient.APIClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *apiclient.APIClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *PrivateResourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PrivateResourceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Map model to request body
	reqBody := map[string]interface{}{
		"name":        data.Name.ValueString(),
		"description": data.Description.ValueString(),
	}

	if !data.DNSServerID.IsNull() {
		reqBody["dnsServerId"] = data.DNSServerID.ValueInt64()
	}
	if !data.CertificateID.IsNull() {
		reqBody["certificateId"] = data.CertificateID.ValueInt64()
	}

	if len(data.ResourceGroupIDs) > 0 {
		ids := make([]int, len(data.ResourceGroupIDs))
		for i, id := range data.ResourceGroupIDs {
			ids[i] = int(id.ValueInt64())
		}
		reqBody["resourceGroupIds"] = ids
	}

	accessTypes := make([]map[string]interface{}, len(data.AccessTypes))
	for i, at := range data.AccessTypes {
		atMap := map[string]interface{}{
			"type": at.Type.ValueString(),
		}
		if !at.ExternalFQDNPrefix.IsNull() {
			atMap["externalFQDNPrefix"] = at.ExternalFQDNPrefix.ValueString()
		}
		if !at.ExternalFQDN.IsNull() {
			atMap["externalFQDN"] = at.ExternalFQDN.ValueString()
		}
		if !at.Protocol.IsNull() {
			atMap["protocol"] = at.Protocol.ValueString()
		}
		if !at.SNI.IsNull() {
			atMap["sni"] = at.SNI.ValueString()
		}
		if !at.SSLVerificationEnabled.IsNull() {
			atMap["sslVerificationEnabled"] = at.SSLVerificationEnabled.ValueBool()
		}
		if len(at.ReachableAddresses) > 0 {
			addrs := make([]string, len(at.ReachableAddresses))
			for j, addr := range at.ReachableAddresses {
				addrs[j] = addr.ValueString()
			}
			atMap["reachableAddresses"] = addrs
		}
		accessTypes[i] = atMap
	}
	reqBody["accessTypes"] = accessTypes

	resourceAddresses := make([]map[string]interface{}, len(data.ResourceAddresses))
	for i, ra := range data.ResourceAddresses {
		raMap := map[string]interface{}{}

		dests := make([]string, len(ra.DestinationAddr))
		for j, d := range ra.DestinationAddr {
			dests[j] = d.ValueString()
		}
		raMap["destinationAddr"] = dests

		pps := make([]map[string]interface{}, len(ra.ProtocolPorts))
		for j, pp := range ra.ProtocolPorts {
			pps[j] = map[string]interface{}{
				"protocol": pp.Protocol.ValueString(),
				"ports":    pp.Ports.ValueString(),
			}
		}
		raMap["protocolPorts"] = pps
		resourceAddresses[i] = raMap
	}
	reqBody["resourceAddresses"] = resourceAddresses

	respHTTP, err := r.client.Query("policies", "privateResources", "POST", reqBody)
	if err != nil {
		resp.Diagnostics.AddError("Error creating private resource", err.Error())
		return
	}
	defer respHTTP.Body.Close()

	if respHTTP.StatusCode != 200 && respHTTP.StatusCode != 201 {
		body, _ := io.ReadAll(respHTTP.Body)
		resp.Diagnostics.AddError("Error creating private resource", fmt.Sprintf("Status: %s, Body: %s", respHTTP.Status, string(body)))
		return
	}

	body, _ := io.ReadAll(respHTTP.Body)
	var created struct {
		ID int `json:"resourceId"`
	}
	if err := json.Unmarshal(body, &created); err != nil {
		resp.Diagnostics.AddError("Error unmarshalling response", err.Error())
		return
	}

	data.ID = types.StringValue(strconv.Itoa(created.ID))

	// Ensure computed fields are known
	for i := range data.AccessTypes {
		if data.AccessTypes[i].ExternalFQDN.IsUnknown() {
			data.AccessTypes[i].ExternalFQDN = types.StringNull()
		}
		if data.AccessTypes[i].ExternalFQDNPrefix.IsUnknown() {
			data.AccessTypes[i].ExternalFQDNPrefix = types.StringNull()
		}
	}

	// Ideally we should read back the full object to populate state, but for now we can just set the ID and assume the rest matches plan
	// Or we can call Read. Let's just set ID and let Read handle it if needed, or just set what we have.
	// Setting what we have is safer for now to avoid another API call, but we should be careful about computed fields.
	// Since we don't have many computed fields other than ID, this is fine.

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PrivateResourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PrivateResourceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	respHTTP, err := r.client.Query("policies", "privateResources/"+data.ID.ValueString(), "GET", nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading private resource", err.Error())
		return
	}
	defer respHTTP.Body.Close()

	if respHTTP.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}

	if respHTTP.StatusCode != 200 {
		body, _ := io.ReadAll(respHTTP.Body)
		resp.Diagnostics.AddError("Error reading private resource", fmt.Sprintf("Status: %s, Body: %s", respHTTP.Status, string(body)))
		return
	}

	body, _ := io.ReadAll(respHTTP.Body)

	// Define struct to unmarshal response
	var resourceObj struct {
		ID               int    `json:"resourceId"`
		Name             string `json:"name"`
		Description      string `json:"description"`
		DNSServerID      int    `json:"dnsServerId"`
		CertificateID    int    `json:"certificateId"`
		ResourceGroupIDs []int  `json:"resourceGroupIds"`
		AccessTypes      []struct {
			Type                   string   `json:"type"`
			ExternalFQDNPrefix     string   `json:"externalFQDNPrefix"`
			ExternalFQDN           string   `json:"externalFQDN"`
			Protocol               string   `json:"protocol"`
			SNI                    string   `json:"sni"`
			SSLVerificationEnabled bool     `json:"sslVerificationEnabled"`
			ReachableAddresses     []string `json:"reachableAddresses"`
		} `json:"accessTypes"`
		ResourceAddresses []struct {
			DestinationAddr []string `json:"destinationAddr"`
			ProtocolPorts   []struct {
				Protocol string `json:"protocol"`
				Ports    string `json:"ports"`
			} `json:"protocolPorts"`
		} `json:"resourceAddresses"`
	}

	if err := json.Unmarshal(body, &resourceObj); err != nil {
		resp.Diagnostics.AddError("Error unmarshalling response", err.Error())
		return
	}

	data.Name = types.StringValue(resourceObj.Name)
	data.Description = types.StringValue(resourceObj.Description)

	if resourceObj.DNSServerID != 0 {
		data.DNSServerID = types.Int64Value(int64(resourceObj.DNSServerID))
	} else {
		data.DNSServerID = types.Int64Null()
	}

	if resourceObj.CertificateID != 0 {
		data.CertificateID = types.Int64Value(int64(resourceObj.CertificateID))
	} else {
		data.CertificateID = types.Int64Null()
	}

	if len(resourceObj.ResourceGroupIDs) > 0 {
		var resGroupIDs []types.Int64
		for _, id := range resourceObj.ResourceGroupIDs {
			resGroupIDs = append(resGroupIDs, types.Int64Value(int64(id)))
		}
		data.ResourceGroupIDs = resGroupIDs
	} else {
		data.ResourceGroupIDs = nil
	}

	// Process AccessTypes with sticky ordering
	var apiAccessTypes []AccessTypeModel
	for _, at := range resourceObj.AccessTypes {
		if at.Type == "branch" {
			continue
		}
		model := AccessTypeModel{
			Type:                   types.StringValue(at.Type),
			SSLVerificationEnabled: types.BoolValue(at.SSLVerificationEnabled),
		}
		if at.ExternalFQDNPrefix != "" {
			model.ExternalFQDNPrefix = types.StringValue(at.ExternalFQDNPrefix)
		} else {
			model.ExternalFQDNPrefix = types.StringNull()
		}
		if at.ExternalFQDN != "" {
			model.ExternalFQDN = types.StringValue(at.ExternalFQDN)
		} else {
			model.ExternalFQDN = types.StringNull()
		}
		if at.Protocol != "" {
			model.Protocol = types.StringValue(at.Protocol)
		} else {
			model.Protocol = types.StringNull()
		}
		if at.SNI != "" {
			model.SNI = types.StringValue(at.SNI)
		} else {
			model.SNI = types.StringNull()
		}

		var reachAddrs []types.String
		for _, ra := range at.ReachableAddresses {
			reachAddrs = append(reachAddrs, types.StringValue(ra))
		}
		model.ReachableAddresses = reachAddrs

		apiAccessTypes = append(apiAccessTypes, model)
	}

	// Reorder apiAccessTypes to match data.AccessTypes (state)
	var orderedAccessTypes []AccessTypeModel
	usedIndices := make(map[int]bool)

	// First pass: find matches for existing state items
	for _, stateItem := range data.AccessTypes {
		for i, apiItem := range apiAccessTypes {
			if usedIndices[i] {
				continue
			}
			// Match logic: Type must match.
			if stateItem.Type.ValueString() == apiItem.Type.ValueString() {
				orderedAccessTypes = append(orderedAccessTypes, apiItem)
				usedIndices[i] = true
				break
			}
		}
	}

	// Second pass: add remaining API items (new ones)
	for i, apiItem := range apiAccessTypes {
		if !usedIndices[i] {
			orderedAccessTypes = append(orderedAccessTypes, apiItem)
		}
	}
	data.AccessTypes = orderedAccessTypes

	var resAddrs []ResourceAddressModel
	for _, ra := range resourceObj.ResourceAddresses {
		var dests []types.String
		for _, d := range ra.DestinationAddr {
			dests = append(dests, types.StringValue(d))
		}

		var pps []ProtocolPortModel
		for _, pp := range ra.ProtocolPorts {
			pps = append(pps, ProtocolPortModel{
				Protocol: types.StringValue(strings.ToLower(pp.Protocol)),
				Ports:    types.StringValue(pp.Ports),
			})
		}

		resAddrs = append(resAddrs, ResourceAddressModel{
			DestinationAddr: dests,
			ProtocolPorts:   pps,
		})
	}
	data.ResourceAddresses = resAddrs

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PrivateResourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PrivateResourceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Map model to request body
	reqBody := map[string]interface{}{
		"name":        data.Name.ValueString(),
		"description": data.Description.ValueString(),
	}

	if !data.DNSServerID.IsNull() {
		reqBody["dnsServerId"] = data.DNSServerID.ValueInt64()
	}
	if !data.CertificateID.IsNull() {
		reqBody["certificateId"] = data.CertificateID.ValueInt64()
	}

	if len(data.ResourceGroupIDs) > 0 {
		ids := make([]int, len(data.ResourceGroupIDs))
		for i, id := range data.ResourceGroupIDs {
			ids[i] = int(id.ValueInt64())
		}
		reqBody["resourceGroupIds"] = ids
	}

	accessTypes := make([]map[string]interface{}, len(data.AccessTypes))
	for i, at := range data.AccessTypes {
		atMap := map[string]interface{}{
			"type": at.Type.ValueString(),
		}
		if !at.ExternalFQDNPrefix.IsNull() {
			atMap["externalFQDNPrefix"] = at.ExternalFQDNPrefix.ValueString()
		}
		if !at.ExternalFQDN.IsNull() {
			atMap["externalFQDN"] = at.ExternalFQDN.ValueString()
		}
		if !at.Protocol.IsNull() {
			atMap["protocol"] = at.Protocol.ValueString()
		}
		if !at.SNI.IsNull() {
			atMap["sni"] = at.SNI.ValueString()
		}
		if !at.SSLVerificationEnabled.IsNull() {
			atMap["sslVerificationEnabled"] = at.SSLVerificationEnabled.ValueBool()
		}
		if len(at.ReachableAddresses) > 0 {
			addrs := make([]string, len(at.ReachableAddresses))
			for j, addr := range at.ReachableAddresses {
				addrs[j] = addr.ValueString()
			}
			atMap["reachableAddresses"] = addrs
		}
		accessTypes[i] = atMap
	}
	reqBody["accessTypes"] = accessTypes

	resourceAddresses := make([]map[string]interface{}, len(data.ResourceAddresses))
	for i, ra := range data.ResourceAddresses {
		raMap := map[string]interface{}{}

		dests := make([]string, len(ra.DestinationAddr))
		for j, d := range ra.DestinationAddr {
			dests[j] = d.ValueString()
		}
		raMap["destinationAddr"] = dests

		pps := make([]map[string]interface{}, len(ra.ProtocolPorts))
		for j, pp := range ra.ProtocolPorts {
			pps[j] = map[string]interface{}{
				"protocol": pp.Protocol.ValueString(),
				"ports":    pp.Ports.ValueString(),
			}
		}
		raMap["protocolPorts"] = pps
		resourceAddresses[i] = raMap
	}
	reqBody["resourceAddresses"] = resourceAddresses

	respHTTP, err := r.client.Query("policies", "privateResources/"+data.ID.ValueString(), "PUT", reqBody)
	if err != nil {
		resp.Diagnostics.AddError("Error updating private resource", err.Error())
		return
	}
	defer respHTTP.Body.Close()

	if respHTTP.StatusCode != 200 {
		body, _ := io.ReadAll(respHTTP.Body)
		resp.Diagnostics.AddError("Error updating private resource", fmt.Sprintf("Status: %s, Body: %s", respHTTP.Status, string(body)))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PrivateResourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PrivateResourceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	respHTTP, err := r.client.Query("policies", "privateResources/"+data.ID.ValueString(), "DELETE", nil)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting private resource", err.Error())
		return
	}
	defer respHTTP.Body.Close()

	if respHTTP.StatusCode != 200 && respHTTP.StatusCode != 204 {
		body, _ := io.ReadAll(respHTTP.Body)
		resp.Diagnostics.AddError("Error deleting private resource", fmt.Sprintf("Status: %s, Body: %s", respHTTP.Status, string(body)))
		return
	}
}

func (r *PrivateResourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id := req.ID
	if _, err := strconv.ParseInt(id, 10, 64); err != nil {
		// Not a number, try to find by name
		foundID, err := apiclient.GetPrivateResourceIDByName(r.client, id)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error importing private resource",
				fmt.Sprintf("Could not find private resource with name '%s': %s", id, err.Error()),
			)
			return
		}
		req.ID = fmt.Sprintf("%d", foundID)
	}
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
