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

var _ resource.Resource = &AppAssignmentResource{}

// AppAssignmentResource manages app assignments in Intune (assigning apps to groups).
type AppAssignmentResource struct {
	client *client.Client
}

// AppAssignmentModel describes the resource data model.
type AppAssignmentModel struct {
	ID      types.String `tfsdk:"id"`
	AppID   types.String `tfsdk:"app_id"`
	GroupID types.String `tfsdk:"group_id"`
	Intent  types.String `tfsdk:"intent"`
}

type graphAppAssignment struct {
	ID     string                `json:"id,omitempty"`
	Intent string                `json:"intent"`
	Target graphAssignmentTarget `json:"target"`
}

func NewAppAssignmentResource() resource.Resource {
	return &AppAssignmentResource{}
}

func (r *AppAssignmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app_assignment"
}

func (r *AppAssignmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Assigns a mobile app to an Azure AD group in Intune.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Composite ID: `{app_id}/{assignment_id}`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"app_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the Intune mobile app to assign.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"group_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the Azure AD group to assign the app to.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"intent": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Assignment intent: 'available', 'required', 'uninstall', or 'availableWithoutEnrollment'.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *AppAssignmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AppAssignmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AppAssignmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := &graphAppAssignment{
		Intent: data.Intent.ValueString(),
		Target: graphAssignmentTarget{
			ODataType: "#microsoft.graph.groupAssignmentTarget",
			GroupID:   data.GroupID.ValueString(),
		},
	}

	var result graphAppAssignment
	path := fmt.Sprintf("/deviceAppManagement/mobileApps/%s/assignments", data.AppID.ValueString())
	if err := r.client.Post(ctx, path, payload, &result); err != nil {
		resp.Diagnostics.AddError("Create App Assignment Error", err.Error())
		return
	}

	data.ID = types.StringValue(data.AppID.ValueString() + "/" + result.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AppAssignmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AppAssignmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parts := strings.SplitN(data.ID.ValueString(), "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid ID", "Expected format: {app_id}/{assignment_id}")
		return
	}
	appID, assignmentID := parts[0], parts[1]

	var result graphAppAssignment
	path := fmt.Sprintf("/deviceAppManagement/mobileApps/%s/assignments/%s", appID, assignmentID)
	if err := r.client.Get(ctx, path, &result); err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read App Assignment Error", err.Error())
		return
	}

	data.Intent = types.StringValue(result.Intent)
	data.GroupID = types.StringValue(result.Target.GroupID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AppAssignmentResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update Not Supported", "App assignments are immutable; use requires_replace.")
}

func (r *AppAssignmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AppAssignmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parts := strings.SplitN(data.ID.ValueString(), "/", 2)
	if len(parts) != 2 {
		return
	}
	appID, assignmentID := parts[0], parts[1]

	path := fmt.Sprintf("/deviceAppManagement/mobileApps/%s/assignments/%s", appID, assignmentID)
	if err := r.client.Delete(ctx, path); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Delete App Assignment Error", err.Error())
	}
}
