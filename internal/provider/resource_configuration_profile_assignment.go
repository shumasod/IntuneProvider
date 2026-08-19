package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shumasod/IntuneProvider/internal/client"
)

var _ resource.Resource = &ConfigProfileAssignmentResource{}

// ConfigProfileAssignmentResource manages assignments between Intune config profiles and groups.
type ConfigProfileAssignmentResource struct {
	client *client.Client
}

// ConfigProfileAssignmentModel describes the resource data model.
type ConfigProfileAssignmentModel struct {
	ID        types.String `tfsdk:"id"`
	ProfileID types.String `tfsdk:"profile_id"`
	GroupID   types.String `tfsdk:"group_id"`
	Intent    types.String `tfsdk:"intent"`
}

type graphProfileAssignment struct {
	ID     string                `json:"id,omitempty"`
	Target graphAssignmentTarget `json:"target"`
	Intent string                `json:"intent,omitempty"`
}

func NewConfigProfileAssignmentResource() resource.Resource {
	return &ConfigProfileAssignmentResource{}
}

func (r *ConfigProfileAssignmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_configuration_profile_assignment"
}

func (r *ConfigProfileAssignmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Assigns an Intune device configuration profile to a group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Composite ID: `{profile_id}/{assignment_id}`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"profile_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the device configuration profile to assign.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"group_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the Azure AD group to assign the profile to.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"intent": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Assignment intent: 'apply', 'remove', or 'available'.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *ConfigProfileAssignmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("expected *client.Client, got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *ConfigProfileAssignmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConfigProfileAssignmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := &graphProfileAssignment{
		Target: graphAssignmentTarget{
			ODataType: "#microsoft.graph.groupAssignmentTarget",
			GroupID:   data.GroupID.ValueString(),
		},
		Intent: data.Intent.ValueString(),
	}

	var result graphProfileAssignment
	path := fmt.Sprintf("/deviceManagement/deviceConfigurations/%s/assignments", data.ProfileID.ValueString())
	if err := r.client.Post(ctx, path, payload, &result); err != nil {
		resp.Diagnostics.AddError("Create Assignment Error", err.Error())
		return
	}

	data.ID = types.StringValue(data.ProfileID.ValueString() + "/" + result.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigProfileAssignmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConfigProfileAssignmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parts := strings.SplitN(data.ID.ValueString(), "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid ID", "Expected format: {profile_id}/{assignment_id}")
		return
	}
	profileID, assignmentID := parts[0], parts[1]

	var result graphProfileAssignment
	path := fmt.Sprintf("/deviceManagement/deviceConfigurations/%s/assignments/%s", profileID, assignmentID)
	if err := r.client.Get(ctx, path, &result); err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Assignment Error", err.Error())
		return
	}

	data.GroupID = types.StringValue(result.Target.GroupID)
	data.Intent = types.StringValue(result.Intent)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigProfileAssignmentResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update Not Supported", "Assignments are immutable; use requires_replace.")
}

func (r *ConfigProfileAssignmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ConfigProfileAssignmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parts := strings.SplitN(data.ID.ValueString(), "/", 2)
	if len(parts) != 2 {
		return
	}
	profileID, assignmentID := parts[0], parts[1]

	path := fmt.Sprintf("/deviceManagement/deviceConfigurations/%s/assignments/%s", profileID, assignmentID)
	if err := r.client.Delete(ctx, path); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Delete Assignment Error", err.Error())
	}
}
