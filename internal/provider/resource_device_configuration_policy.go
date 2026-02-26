package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shumasod/IntuneProvider/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var _ resource.Resource = &DeviceConfigurationPolicyResource{}
var _ resource.ResourceWithImportState = &DeviceConfigurationPolicyResource{}

// DeviceConfigurationPolicyResource manages Intune device configuration policies.
type DeviceConfigurationPolicyResource struct {
	client *client.Client
}

// DeviceConfigurationPolicyModel describes the resource data model.
type DeviceConfigurationPolicyModel struct {
	ID                  types.String `tfsdk:"id"`
	DisplayName         types.String `tfsdk:"display_name"`
	Description         types.String `tfsdk:"description"`
	ODataType           types.String `tfsdk:"odata_type"`
	Assignments         types.List   `tfsdk:"assignments"`
	Settings            types.Map    `tfsdk:"settings"`
	LastModifiedDateTime types.String `tfsdk:"last_modified_date_time"`
	CreatedDateTime     types.String `tfsdk:"created_date_time"`
}

// graphDeviceConfigurationPolicy represents the Graph API payload for device configuration.
type graphDeviceConfigurationPolicy struct {
	ODataType               string                 `json:"@odata.type"`
	ID                      string                 `json:"id,omitempty"`
	DisplayName             string                 `json:"displayName"`
	Description             string                 `json:"description,omitempty"`
	Settings                map[string]interface{} `json:"settings,omitempty"`
	LastModifiedDateTime    string                 `json:"lastModifiedDateTime,omitempty"`
	CreatedDateTime         string                 `json:"createdDateTime,omitempty"`
}

// graphAssignment represents the Graph API assignment payload.
type graphAssignment struct {
	Target graphAssignmentTarget `json:"target"`
}

type graphAssignmentTarget struct {
	ODataType string `json:"@odata.type"`
	GroupID   string `json:"groupId,omitempty"`
}

// graphAssignmentsResponse wraps the assignments list response.
type graphAssignmentsResponse struct {
	Value []struct {
		ID     string              `json:"id"`
		Target graphAssignmentTarget `json:"target"`
	} `json:"value"`
}

func NewDeviceConfigurationPolicyResource() resource.Resource {
	return &DeviceConfigurationPolicyResource{}
}

func (r *DeviceConfigurationPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_configuration_policy"
}

func (r *DeviceConfigurationPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
Manages an Intune **Device Configuration Policy** (` + "`deviceManagement/deviceConfigurations`" + `).

Use this resource to create and manage device configuration profiles for Windows, iOS, and Android
devices enrolled in Microsoft Intune.

## Example Usage

` + "```hcl" + `
resource "intune_device_configuration_policy" "windows_defender" {
  display_name = "Windows Defender Configuration"
  description  = "Enables Windows Defender antivirus settings"
  odata_type   = "#microsoft.graph.windows10EndpointProtectionConfiguration"

  settings = {
    defenderEnabled = "true"
    defenderMonitorFileActivity = "monitorAllFiles"
  }

  assignments = ["00000000-0000-0000-0000-000000000001"]
}
` + "```",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the device configuration policy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The display name of the device configuration policy.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Description of the device configuration policy.",
			},
			"odata_type": schema.StringAttribute{
				Required: true,
				MarkdownDescription: `The OData type that determines the configuration profile type. Common values:
- ` + "`#microsoft.graph.windows10GeneralConfiguration`" + `
- ` + "`#microsoft.graph.windows10EndpointProtectionConfiguration`" + `
- ` + "`#microsoft.graph.iosGeneralDeviceConfiguration`" + `
- ` + "`#microsoft.graph.androidGeneralDeviceConfiguration`" + ``,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"assignments": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "List of Azure AD group IDs to assign this policy to.",
			},
			"settings": schema.MapAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Map of configuration settings specific to the `odata_type`. Keys and values vary by policy type.",
			},
			"last_modified_date_time": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The date and time the configuration policy was last modified (ISO 8601).",
			},
			"created_date_time": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The date and time the configuration policy was created (ISO 8601).",
			},
		},
	}
}

func (r *DeviceConfigurationPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *DeviceConfigurationPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DeviceConfigurationPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := graphDeviceConfigurationPolicy{
		ODataType:   plan.ODataType.ValueString(),
		DisplayName: plan.DisplayName.ValueString(),
		Description: plan.Description.ValueString(),
	}

	if !plan.Settings.IsNull() && !plan.Settings.IsUnknown() {
		settings := make(map[string]interface{})
		for k, v := range plan.Settings.Elements() {
			if sv, ok := v.(types.String); ok {
				settings[k] = sv.ValueString()
			}
		}
		payload.Settings = settings
	}

	var result graphDeviceConfigurationPolicy
	err := r.client.Post(ctx, "/deviceManagement/deviceConfigurations", payload, &result)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create device configuration policy", err.Error())
		return
	}

	plan.ID = types.StringValue(result.ID)
	plan.LastModifiedDateTime = types.StringValue(result.LastModifiedDateTime)
	plan.CreatedDateTime = types.StringValue(result.CreatedDateTime)

	// Apply group assignments if provided
	if !plan.Assignments.IsNull() && !plan.Assignments.IsUnknown() {
		if diags := r.applyAssignments(ctx, result.ID, plan.Assignments); diags != nil {
			resp.Diagnostics.AddError("Failed to assign device configuration policy", diags.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DeviceConfigurationPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DeviceConfigurationPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result graphDeviceConfigurationPolicy
	err := r.client.Get(ctx, "/deviceManagement/deviceConfigurations/"+state.ID.ValueString(), &result)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read device configuration policy", err.Error())
		return
	}

	state.DisplayName = types.StringValue(result.DisplayName)
	state.Description = types.StringValue(result.Description)
	state.ODataType = types.StringValue(result.ODataType)
	state.LastModifiedDateTime = types.StringValue(result.LastModifiedDateTime)
	state.CreatedDateTime = types.StringValue(result.CreatedDateTime)

	// Read assignments
	assignments, err := r.readAssignments(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read assignments", err.Error())
		return
	}
	state.Assignments = assignments

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DeviceConfigurationPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DeviceConfigurationPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := graphDeviceConfigurationPolicy{
		ODataType:   plan.ODataType.ValueString(),
		DisplayName: plan.DisplayName.ValueString(),
		Description: plan.Description.ValueString(),
	}

	if !plan.Settings.IsNull() && !plan.Settings.IsUnknown() {
		settings := make(map[string]interface{})
		for k, v := range plan.Settings.Elements() {
			if sv, ok := v.(types.String); ok {
				settings[k] = sv.ValueString()
			}
		}
		payload.Settings = settings
	}

	err := r.client.Patch(ctx, "/deviceManagement/deviceConfigurations/"+plan.ID.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update device configuration policy", err.Error())
		return
	}

	if !plan.Assignments.IsNull() && !plan.Assignments.IsUnknown() {
		if diags := r.applyAssignments(ctx, plan.ID.ValueString(), plan.Assignments); diags != nil {
			resp.Diagnostics.AddError("Failed to update assignments", diags.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DeviceConfigurationPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DeviceConfigurationPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, "/deviceManagement/deviceConfigurations/"+state.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete device configuration policy", err.Error())
	}
}

func (r *DeviceConfigurationPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var state DeviceConfigurationPolicyModel
	state.ID = types.StringValue(req.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// applyAssignments sends a batch assign request for all specified group IDs.
func (r *DeviceConfigurationPolicyResource) applyAssignments(ctx context.Context, policyID string, assignments types.List) error {
	var groupIDs []string
	for _, v := range assignments.Elements() {
		if sv, ok := v.(types.String); ok {
			groupIDs = append(groupIDs, sv.ValueString())
		}
	}

	var payloadAssignments []graphAssignment
	for _, gid := range groupIDs {
		payloadAssignments = append(payloadAssignments, graphAssignment{
			Target: graphAssignmentTarget{
				ODataType: "#microsoft.graph.groupAssignmentTarget",
				GroupID:   gid,
			},
		})
	}

	payload := map[string]interface{}{"assignments": payloadAssignments}
	return r.client.Post(ctx, fmt.Sprintf("/deviceManagement/deviceConfigurations/%s/assign", policyID), payload, nil)
}

// readAssignments retrieves the current group assignments for a policy.
func (r *DeviceConfigurationPolicyResource) readAssignments(ctx context.Context, policyID string) (types.List, error) {
	var result graphAssignmentsResponse
	err := r.client.Get(ctx, fmt.Sprintf("/deviceManagement/deviceConfigurations/%s/assignments", policyID), &result)
	if err != nil {
		return types.ListNull(types.StringType), err
	}

	var groupIDs []attr.Value
	for _, a := range result.Value {
		if a.Target.GroupID != "" {
			groupIDs = append(groupIDs, types.StringValue(a.Target.GroupID))
		}
	}

	list, diags := types.ListValue(types.StringType, groupIDs)
	if diags.HasError() {
		return types.ListNull(types.StringType), fmt.Errorf("failed to build assignments list")
	}
	return list, nil
}
