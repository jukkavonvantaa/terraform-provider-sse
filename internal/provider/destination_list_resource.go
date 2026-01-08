// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cisco/terraform-provider-sse/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DestinationListResource{}
var _ resource.ResourceWithImportState = &DestinationListResource{}

func NewDestinationListResource() resource.Resource {
	return &DestinationListResource{}
}

// DestinationListResource defines the resource implementation.
type DestinationListResource struct {
	client *apiclient.APIClient
}

// DestinationListResourceModel describes the resource data model.
type DestinationListResourceModel struct {
	ID     types.String `tfsdk:"id"`
	ListID types.Int64  `tfsdk:"list_id"`
	// OrganizationID       types.Int64  `tfsdk:"organization_id"`
	Access   types.String `tfsdk:"access"`
	IsGlobal types.Bool   `tfsdk:"is_global"`
	Name     types.String `tfsdk:"name"`
	// ThirdpartyCategoryID types.Int64  `tfsdk:"thirdparty_category_id"`
	// CreatedAt            types.Int64        `tfsdk:"created_at"`
	// ModifiedAt           types.Int64        `tfsdk:"modified_at"`
	// IsMspDefault      types.Bool         `tfsdk:"is_msp_default"`
	// MarkedForDeletion types.Bool         `tfsdk:"marked_for_deletion"`
	BundleTypeID types.Int64 `tfsdk:"bundle_type_id"`
	Destinations types.List  `tfsdk:"destinations"`
}

type DestinationModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Destination types.String `tfsdk:"destination"`
	Type        types.String `tfsdk:"type"`
	Comment     types.String `tfsdk:"comment"`
}

func (r *DestinationListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_destination_list"
}

func (r *DestinationListResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Destination List resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Destination List ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"list_id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Destination List ID (Integer), useful for JSON encoding in rules.",
			},
			// "organization_id": schema.Int64Attribute{
			// 	Computed: true,
			// },
			"access": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"is_global": schema.BoolAttribute{
				Required: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			// "thirdparty_category_id": schema.Int64Attribute{
			// 	Computed: true,
			// },
			// "created_at": schema.Int64Attribute{
			// 	Computed: true,
			// },
			// "modified_at": schema.Int64Attribute{
			// 	Computed: true,
			// },
			// "is_msp_default": schema.BoolAttribute{
			// 	Computed: true,
			// },
			// "marked_for_deletion": schema.BoolAttribute{
			// 	Computed: true,
			// },
			"bundle_type_id": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Bundle Type ID. Available values: `1` (DNS - Domains), `2` (Web - Domains, URLs, IPs), `4` (SAML Bypass). Defaults to `1`. **Note:** When you create a destination list for Web policies, set the `bundle_type_id` to `2`.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"destinations": schema.ListNestedAttribute{
				Optional: true,
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed: true,
							Optional: true,
						},
						"destination": schema.StringAttribute{
							Required: true,
						},
						"type": schema.StringAttribute{
							Required: true,
						},
						"comment": schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func (r *DestinationListResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DestinationListResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DestinationListResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	payload := apiclient.CreateDestinationListPayload{
		Access:   data.Access.ValueString(),
		IsGlobal: data.IsGlobal.ValueBool(),
		Name:     data.Name.ValueString(),
	}

	if !data.BundleTypeID.IsNull() {
		payload.BundleTypeID = int(data.BundleTypeID.ValueInt64())
	} else {
		// Default to 1 (DNS) if not specified
		payload.BundleTypeID = 1
	}

	list, err := apiclient.PostDestinationList(r.client, payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating destination list",
			"Could not create destination list, unexpected error: "+err.Error(),
		)
		return
	}

	data.ID = types.StringValue(strconv.FormatInt(list.ID, 10))
	data.ListID = types.Int64Value(list.ID)
	data.BundleTypeID = types.Int64Value(int64(list.BundleTypeID))
	// data.OrganizationID = types.Int64Value(list.OrganizationID)
	// data.ThirdpartyCategoryID = types.Int64Value(int64(list.ThirdpartyCategoryID))
	// data.CreatedAt = types.Int64Value(list.CreatedAt)
	// data.ModifiedAt = types.Int64Value(list.ModifiedAt)
	// data.IsMspDefault = types.BoolValue(list.IsMspDefault)
	// data.MarkedForDeletion = types.BoolValue(list.MarkedForDeletion)

	// Handle destinations
	var planDestinations []DestinationModel
	if !data.Destinations.IsNull() && !data.Destinations.IsUnknown() {
		resp.Diagnostics.Append(data.Destinations.ElementsAs(ctx, &planDestinations, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		var destinations []apiclient.Destination
		for _, d := range planDestinations {
			destinations = append(destinations, apiclient.Destination{
				Destination: d.Destination.ValueString(),
				Type:        d.Type.ValueString(),
				Comment:     d.Comment.ValueString(),
			})
		}

		err := apiclient.PostDestinations(r.client, list.ID, destinations)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error adding destinations",
				"Could not add destinations to list, unexpected error: "+err.Error(),
			)
			// Try to cleanup
			_ = apiclient.DeleteDestinationList(r.client, list.ID)
			return
		}
	}

	if len(planDestinations) > 0 {
		var dests []apiclient.Destination
		var err error

		// Retry loop to handle eventual consistency
		// Increased to 30 seconds as API can be slow to index
		for i := 0; i < 30; i++ {
			dests, err = apiclient.GetDestinationsDetails(r.client, list.ID)
			if err == nil && len(dests) >= len(planDestinations) {
				break
			}
			time.Sleep(1 * time.Second)
		}

		if err != nil || len(dests) < len(planDestinations) {
			resp.Diagnostics.AddWarning(
				"Consistency Warning",
				"Could not read back all created destinations from API immediately. State may be incomplete. Please run 'tofu apply' again later to refresh state.",
			)
			// Fallback: use plan data but set IDs to null where unknown
			// This prevents "inconsistent result" error by providing the expected list structure
			var fallbackDests []DestinationModel
			for _, d := range planDestinations {
				newD := d
				if newD.ID.IsUnknown() {
					newD.ID = types.Int64Null()
				}
				fallbackDests = append(fallbackDests, newD)
			}
			data.Destinations, _ = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: map[string]attr.Type{
				"id":          types.Int64Type,
				"destination": types.StringType,
				"type":        types.StringType,
				"comment":     types.StringType,
			}}, fallbackDests)
		} else {
			// Map back to model using helper to ensure consistency
			data.Destinations = mapDestinationsToModel(ctx, dests, planDestinations)
		}
	} else {
		data.Destinations = types.ListNull(types.ObjectType{AttrTypes: map[string]attr.Type{
			"id":          types.Int64Type,
			"destination": types.StringType,
			"type":        types.StringType,
			"comment":     types.StringType,
		}})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DestinationListResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DestinationListResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(data.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", "Could not parse ID")
		return
	}

	list, err := apiclient.GetDestinationListDetails(r.client, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading destination list",
			"Could not read destination list ID "+data.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	data.Name = types.StringValue(list.Name)
	data.ListID = types.Int64Value(list.ID)
	data.Access = types.StringValue(list.Access)
	data.IsGlobal = types.BoolValue(list.IsGlobal)
	data.BundleTypeID = types.Int64Value(int64(list.BundleTypeID))
	// data.OrganizationID = types.Int64Value(list.OrganizationID)
	// data.ThirdpartyCategoryID = types.Int64Value(int64(list.ThirdpartyCategoryID))
	// data.CreatedAt = types.Int64Value(list.CreatedAt)
	// data.ModifiedAt = types.Int64Value(list.ModifiedAt)
	// data.IsMspDefault = types.BoolValue(list.IsMspDefault)
	// data.MarkedForDeletion = types.BoolValue(list.MarkedForDeletion)

	// Read destinations
	dests, err := apiclient.GetDestinationsDetails(r.client, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading destinations",
			"Could not read destinations for list ID "+data.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	var stateDestinations []DestinationModel
	if !data.Destinations.IsNull() && !data.Destinations.IsUnknown() {
		data.Destinations.ElementsAs(ctx, &stateDestinations, false)
	}

	// Map back to model using helper to ensure consistency with state
	data.Destinations = mapDestinationsToModel(ctx, dests, stateDestinations)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DestinationListResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DestinationListResourceModel
	var state DestinationListResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(data.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", "Could not parse ID")
		return
	}

	// Update Name if changed
	if !data.Name.Equal(state.Name) {
		payload := apiclient.UpdateDestinationListPayload{
			Name: data.Name.ValueString(),
		}
		_, err := apiclient.PatchDestinationList(r.client, id, payload)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating destination list",
				"Could not update destination list: "+err.Error(),
			)
			return
		}
	}

	// Update Destinations
	// Simple strategy: Calculate diff
	// To delete: IDs in state but not in plan (or matching content not in plan)
	// To add: Content in plan not in state

	var stateDestinations []DestinationModel
	if !state.Destinations.IsNull() && !state.Destinations.IsUnknown() {
		state.Destinations.ElementsAs(ctx, &stateDestinations, false)
	}

	var planDests []DestinationModel
	if !data.Destinations.IsNull() && !data.Destinations.IsUnknown() {
		data.Destinations.ElementsAs(ctx, &planDests, false)
	}

	// Map state destinations by ID for easy lookup
	stateDestMap := make(map[int64]DestinationModel)
	for _, d := range stateDestinations {
		if !d.ID.IsNull() {
			stateDestMap[d.ID.ValueInt64()] = d
		}
	}

	// Identify destinations to keep (and thus which to delete)
	// Since plan doesn't have IDs for new items, we match by content?
	// But existing items in plan might have IDs if they were preserved?
	// Terraform plan should preserve IDs for existing items if they match.

	// Actually, for a ListNestedAttribute, Terraform might not preserve IDs if the list order changes significantly or if it's treated as a set.
	// But here it's a list.

	// Let's look at what we have.
	// We can just delete all and re-add all? That's inefficient and changes IDs.
	// Better:
	// 1. Find IDs to delete.
	// 2. Find items to add.

	// If the user provides IDs in the config (they shouldn't, it's computed), we could use them.
	// But usually user config doesn't have IDs.

	// We need to match plan items to state items to see what's preserved.
	// However, since we don't have a unique key other than ID (which is computed), it's hard to map exactly if duplicates are allowed.
	// Assuming no duplicates for (Destination, Type).

	// Let's try to match by (Destination, Type).

	toAdd := []apiclient.Destination{}
	toDeleteIDs := []int64{}

	// Helper to check if a destination exists in a list
	existsInPlan := func(d DestinationModel) bool {
		for _, pd := range planDests {
			if pd.Destination.Equal(d.Destination) && pd.Type.Equal(d.Type) {
				// Check comment too?
				if pd.Comment.Equal(d.Comment) {
					return true
				}
			}
		}
		return false
	}

	// Find items in state that are NOT in plan -> Delete
	for _, sd := range stateDestinations {
		if !existsInPlan(sd) {
			if !sd.ID.IsNull() {
				toDeleteIDs = append(toDeleteIDs, sd.ID.ValueInt64())
			}
		}
	}

	// Helper to check if a destination exists in state
	existsInState := func(d DestinationModel) bool {
		for _, sd := range stateDestinations {
			if sd.Destination.Equal(d.Destination) && sd.Type.Equal(d.Type) {
				if sd.Comment.Equal(d.Comment) {
					return true
				}
			}
		}
		return false
	}

	// Find items in plan that are NOT in state -> Add
	for _, pd := range planDests {
		if !existsInState(pd) {
			toAdd = append(toAdd, apiclient.Destination{
				Destination: pd.Destination.ValueString(),
				Type:        pd.Type.ValueString(),
				Comment:     pd.Comment.ValueString(),
			})
		}
	}

	if len(toDeleteIDs) > 0 {
		err := apiclient.DeleteDestinations(r.client, id, toDeleteIDs)
		if err != nil {
			resp.Diagnostics.AddError("Error deleting destinations", err.Error())
			return
		}
	}

	if len(toAdd) > 0 {
		err := apiclient.PostDestinations(r.client, id, toAdd)
		if err != nil {
			resp.Diagnostics.AddError("Error adding destinations", err.Error())
			return
		}
	}

	// Refresh state
	// Read back everything
	list, err := apiclient.GetDestinationListDetails(r.client, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading updated destination list", err.Error())
		return
	}
	data.ListID = types.Int64Value(list.ID)

	// data.ModifiedAt = types.Int64Value(list.ModifiedAt)

	dests, err := apiclient.GetDestinationsDetails(r.client, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading updated destinations", err.Error())
		return
	}

	// Map back to model using helper to ensure consistency with plan
	data.Destinations = mapDestinationsToModel(ctx, dests, planDests)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DestinationListResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DestinationListResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(data.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", "Could not parse ID")
		return
	}

	err = apiclient.DeleteDestinationList(r.client, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting destination list",
			"Could not delete destination list ID "+data.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}

func (r *DestinationListResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id := req.ID
	if _, err := strconv.ParseInt(id, 10, 64); err != nil {
		// Not a number, try to find by name
		foundID, err := apiclient.GetDestinationListIDByName(r.client, id)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error importing destination list",
				fmt.Sprintf("Could not find destination list with name '%s': %s", id, err.Error()),
			)
			return
		}
		req.ID = fmt.Sprintf("%d", foundID)
	}
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func mapDestinationsToModel(ctx context.Context, apiDests []apiclient.Destination, refDests []DestinationModel) types.List {
	var destModels []DestinationModel
	for _, d := range apiDests {
		id, _ := strconv.ParseInt(d.ID, 10, 64)

		// Handle comment being empty string vs null
		comment := types.StringValue(d.Comment)
		if d.Comment == "" {
			comment = types.StringNull()
		}

		destType := types.StringValue(d.Type)

		// Find matching destination in reference to check what user provided
		for _, refDest := range refDests {
			if refDest.Destination.ValueString() == d.Destination {
				// If types match case-insensitively, use the reference's value
				if strings.EqualFold(refDest.Type.ValueString(), d.Type) {
					destType = refDest.Type
				}
				// Also try to preserve comment null-ness if it matches empty string
				if d.Comment == "" {
					if refDest.Comment.IsNull() {
						comment = types.StringNull()
					} else {
						comment = types.StringValue("")
					}
				}
				break
			}
		}

		destModels = append(destModels, DestinationModel{
			ID:          types.Int64Value(id),
			Destination: types.StringValue(d.Destination),
			Type:        destType,
			Comment:     comment,
		})
	}

	result, _ := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: map[string]attr.Type{
		"id":          types.Int64Type,
		"destination": types.StringType,
		"type":        types.StringType,
		"comment":     types.StringType,
	}}, destModels)

	return result
}
