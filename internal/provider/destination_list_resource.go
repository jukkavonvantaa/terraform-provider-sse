// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cisco/terraform-provider-sse/internal/apiclient"
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
	ID types.String `tfsdk:"id"`
	// OrganizationID       types.Int64  `tfsdk:"organization_id"`
	Access   types.String `tfsdk:"access"`
	IsGlobal types.Bool   `tfsdk:"is_global"`
	Name     types.String `tfsdk:"name"`
	// ThirdpartyCategoryID types.Int64  `tfsdk:"thirdparty_category_id"`
	// CreatedAt            types.Int64        `tfsdk:"created_at"`
	// ModifiedAt           types.Int64        `tfsdk:"modified_at"`
	// IsMspDefault      types.Bool         `tfsdk:"is_msp_default"`
	// MarkedForDeletion types.Bool         `tfsdk:"marked_for_deletion"`
	BundleTypeID types.Int64        `tfsdk:"bundle_type_id"`
	Destinations []DestinationModel `tfsdk:"destinations"`
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
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"destinations": schema.ListNestedAttribute{
				Optional: true,
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
		Access:       data.Access.ValueString(),
		IsGlobal:     data.IsGlobal.ValueBool(),
		Name:         data.Name.ValueString(),
		BundleTypeID: int(data.BundleTypeID.ValueInt64()),
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
	// data.OrganizationID = types.Int64Value(list.OrganizationID)
	// data.ThirdpartyCategoryID = types.Int64Value(int64(list.ThirdpartyCategoryID))
	// data.CreatedAt = types.Int64Value(list.CreatedAt)
	// data.ModifiedAt = types.Int64Value(list.ModifiedAt)
	// data.IsMspDefault = types.BoolValue(list.IsMspDefault)
	// data.MarkedForDeletion = types.BoolValue(list.MarkedForDeletion)

	// Handle destinations
	if len(data.Destinations) > 0 {
		var destinations []apiclient.Destination
		for _, d := range data.Destinations {
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

	// Read back to get IDs of destinations
	// We need to populate the IDs in the state
	if len(data.Destinations) > 0 {
		dests, err := apiclient.GetDestinationsDetails(r.client, list.ID)
		if err != nil {
			resp.Diagnostics.AddWarning(
				"Error reading back destinations",
				"Could not read back destinations to populate IDs: "+err.Error(),
			)
		} else {
			// Map back to model
			// Note: The order might not be preserved, so we might need to match by content
			// For simplicity, we'll just replace the list with what we got back
			var destModels []DestinationModel
			for _, d := range dests {
				id, _ := strconv.ParseInt(d.ID, 10, 64)
				destModels = append(destModels, DestinationModel{
					ID:          types.Int64Value(id),
					Destination: types.StringValue(d.Destination),
					Type:        types.StringValue(d.Type),
					Comment:     types.StringValue(d.Comment),
				})
			}
			data.Destinations = destModels
		}
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

	var destModels []DestinationModel
	for _, d := range dests {
		id, _ := strconv.ParseInt(d.ID, 10, 64)
		destModels = append(destModels, DestinationModel{
			ID:          types.Int64Value(id),
			Destination: types.StringValue(d.Destination),
			Type:        types.StringValue(d.Type),
			Comment:     types.StringValue(d.Comment),
		})
	}
	data.Destinations = destModels

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

	// Map state destinations by ID for easy lookup
	stateDestMap := make(map[int64]DestinationModel)
	for _, d := range state.Destinations {
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

	planDests := data.Destinations

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
	for _, sd := range state.Destinations {
		if !existsInPlan(sd) {
			if !sd.ID.IsNull() {
				toDeleteIDs = append(toDeleteIDs, sd.ID.ValueInt64())
			}
		}
	}

	// Helper to check if a destination exists in state
	existsInState := func(d DestinationModel) bool {
		for _, sd := range state.Destinations {
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
	_, err = apiclient.GetDestinationListDetails(r.client, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading updated destination list", err.Error())
		return
	}

	// data.ModifiedAt = types.Int64Value(list.ModifiedAt)

	dests, err := apiclient.GetDestinationsDetails(r.client, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading updated destinations", err.Error())
		return
	}

	var destModels []DestinationModel
	for _, d := range dests {
		id, _ := strconv.ParseInt(d.ID, 10, 64)
		destModels = append(destModels, DestinationModel{
			ID:          types.Int64Value(id),
			Destination: types.StringValue(d.Destination),
			Type:        types.StringValue(d.Type),
			Comment:     types.StringValue(d.Comment),
		})
	}
	data.Destinations = destModels

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
