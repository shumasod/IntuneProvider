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

var _ resource.Resource = &DeviceCompliancePolicyResource{}
var _ resource.ResourceWithImportState = &DeviceCompliancePolicyResource{}

// DeviceCompliancePolicyResource manages Intune device compliance policies.
type DeviceCompliancePolicyResource struct {
	client *client.Client
}

// DeviceCompliancePolicyModel describes the resource data model.
type DeviceCompliancePolicyModel struct {
	ID                   types.String `tfsdk:"id"`
	DisplayName          types.String `tfsdk:"display_name"`
	Description          types.String `tfsdk:"description"`
	ODataType            types.String `tfsdk:"odata_type"`
	Assignments          types.List   `tfsdk:"assignments"`
	Settings             types.Map    `tfsdk:"settings"`
	ScheduledActionConfigs types.List `tfsdk:"scheduled_action_configs"`
	LastModifiedDateTime types.String `tfsdk:"last_modified_date_time"`
	CreatedDateTime      types.String `tfsdk:"created_date_time"`
}

// graphDeviceCompliancePolicy is the Graph API payload structure for compliance policies.
type graphDeviceCompliancePolicy struct {
	ODataType            string                 `json:"@odata.type"`
	ID                   string                 `json:"id,omitempty"`
	DisplayName          string                 `json:"displayName"`
	Description          string                 `json:"description,omitempty"`
	Settings             map[string]interface{} `json:"settings,omitempty"`
	LastModifiedDateTime string                 `json:"lastModifiedDateTime,omitempty"`
	CreatedDateTime      string                 `json:"createdDateTime,omitempty"`
}

// graphScheduledAction represents a non-compliance scheduled action payload.
type graphScheduledAction struct {
	RuleName                      string                          `json:"ruleName,omitempty"`
	ScheduledActionConfigurations []graphScheduledActionConfig    `json:"scheduledActionConfigurations,omitempty"`
}

type graphScheduledActionConfig struct {
	ActionType                  string `json:"actionType"`
	GracePeriodHours            int    `json:"gracePeriodHours,omitempty"`
	NotificationTemplateID      string `json:"notificationTemplateId,omitempty"`
	NotificationMessageCCList   []string `json:"notificationMessageCCList,omitempty"`
}

func NewDeviceCompliancePolicyResource() resource.Resource {
	return &DeviceCompliancePolicyResource{}
}

func (r *DeviceCompliancePolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_compliance_policy"
}

func (r *DeviceCompliancePolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	scheduledActionConfigAttrTypes := map[string]attr.Type{
		"action_type":        types.StringType,
		"grace_period_hours": types.Int64Type,
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: `
Manages an Intune **Device Compliance Policy** (` + "`deviceManagement/deviceCompliancePolicies`" + `).

Compliance policies define the rules and settings that a device must comply with to be considered
compliant. Non-compliant devices can be blocked from accessing corporate resources via Conditional Access.

## Example Usage

` + "```hcl" + `
resource "intune_device_compliance_policy" "windows_basic" {
  display_name = "Windows 10 Basic Compliance"
  description  = "Requires BitLocker and up-to-date OS"
  odata_type   = "#microsoft.graph.windows10CompliancePolicy"

  settings = {
    bitLockerEnabled          = "true"
    osMinimumVersion          = "10.0.19041"
    passwordRequired          = "true"
    passwordMinimumLength     = "8"
  }

  assignments = ["00000000-0000-0000-0000-000000000001"]

  scheduled_action_configs = [
    {
      action_type        = "block"
      grace_period_hours = 0
    }
  ]
}
` + "```",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the compliance policy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The display name of the compliance policy.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Description of the compliance policy.",
			},
			"odata_type": schema.StringAttribute{
				Required: true,
				MarkdownDescription: `The OData type for the compliance policy. Common values:
- ` + "`#microsoft.graph.windows10CompliancePolicy`" + `
- ` + "`#microsoft.graph.iosCompliancePolicy`" + `
- ` + "`#microsoft.graph.androidCompliancePolicy`" + `
- ` + "`#microsoft.graph.macOSCompliancePolicy`" + ``,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"assignments": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "List of Azure AD group IDs to assign this compliance policy to.",
			},
			"settings": schema.MapAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Map of compliance settings. Available keys depend on `odata_type`.",
			},
			"scheduled_action_configs": schema.ListNestedAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "List of scheduled actions to take when a device becomes non-compliant.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"action_type": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The action to take. Allowed values: `block`, `retire`, `wipe`, `notification`.",
						},
						"grace_period_hours": schema.Int64Attribute{
							Optional:            true,
							Computed:            true,
							MarkdownDescription: "Hours before the action is applied after non-compliance is detected. `0` means immediate.",
						},
					},
					CustomType: types.ObjectType{AttrTypes: scheduledActionConfigAttrTypes},
				},
			},
			"last_modified_date_time": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The date and time the policy was last modified (ISO 8601).",
			},
			"created_date_time": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The date and time the policy was created (ISO 8601).",
			},
		},
	}
}

func (r *DeviceCompliancePolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DeviceCompliancePolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DeviceCompliancePolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := r.buildPayload(plan)

	var result graphDeviceCompliancePolicy
	err := r.client.Post(ctx, "/deviceManagement/deviceCompliancePolicies", payload, &result)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create device compliance policy", err.Error())
		return
	}

	plan.ID = types.StringValue(result.ID)
	plan.LastModifiedDateTime = types.StringValue(result.LastModifiedDateTime)
	plan.CreatedDateTime = types.StringValue(result.CreatedDateTime)

	// Apply scheduled actions (non-compliance actions)
	if err := r.applyScheduledActions(ctx, result.ID, plan.ScheduledActionConfigs); err != nil {
		resp.Diagnostics.AddError("Failed to set scheduled actions", err.Error())
		return
	}

	if !plan.Assignments.IsNull() && !plan.Assignments.IsUnknown() {
		if err := r.applyComplianceAssignments(ctx, result.ID, plan.Assignments); err != nil {
			resp.Diagnostics.AddError("Failed to assign compliance policy", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DeviceCompliancePolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DeviceCompliancePolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result graphDeviceCompliancePolicy
	err := r.client.Get(ctx, "/deviceManagement/deviceCompliancePolicies/"+state.ID.ValueString(), &result)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read device compliance policy", err.Error())
		return
	}

	state.DisplayName = types.StringValue(result.DisplayName)
	state.Description = types.StringValue(result.Description)
	state.ODataType = types.StringValue(result.ODataType)
	state.LastModifiedDateTime = types.StringValue(result.LastModifiedDateTime)
	state.CreatedDateTime = types.StringValue(result.CreatedDateTime)

	assignments, err := r.readComplianceAssignments(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read compliance assignments", err.Error())
		return
	}
	state.Assignments = assignments

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DeviceCompliancePolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DeviceCompliancePolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := r.buildPayload(plan)

	err := r.client.Patch(ctx, "/deviceManagement/deviceCompliancePolicies/"+plan.ID.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update device compliance policy", err.Error())
		return
	}

	if !plan.Assignments.IsNull() && !plan.Assignments.IsUnknown() {
		if err := r.applyComplianceAssignments(ctx, plan.ID.ValueString(), plan.Assignments); err != nil {
			resp.Diagnostics.AddError("Failed to update compliance assignments", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DeviceCompliancePolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DeviceCompliancePolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, "/deviceManagement/deviceCompliancePolicies/"+state.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete device compliance policy", err.Error())
	}
}

func (r *DeviceCompliancePolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var state DeviceCompliancePolicyModel
	state.ID = types.StringValue(req.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DeviceCompliancePolicyResource) buildPayload(plan DeviceCompliancePolicyModel) graphDeviceCompliancePolicy {
	payload := graphDeviceCompliancePolicy{
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
	return payload
}

func (r *DeviceCompliancePolicyResource) applyScheduledActions(ctx context.Context, policyID string, configs types.List) error {
	if configs.IsNull() || configs.IsUnknown() {
		// Default: block immediately
		action := graphScheduledAction{
			RuleName: "NonCompliance",
			ScheduledActionConfigurations: []graphScheduledActionConfig{
				{ActionType: "block", GracePeriodHours: 0},
			},
		}
		payload := map[string]interface{}{"deviceComplianceScheduledActionForRules": []graphScheduledAction{action}}
		return r.client.Post(ctx, fmt.Sprintf("/deviceManagement/deviceCompliancePolicies/%s/scheduleActionsForRules", policyID), payload, nil)
	}

	var actionConfigs []graphScheduledActionConfig
	elements := configs.Elements()
	for _, elem := range elements {
		if obj, ok := elem.(types.Object); ok {
			attrs := obj.Attributes()
			config := graphScheduledActionConfig{}
			if v, ok := attrs["action_type"].(types.String); ok {
				config.ActionType = v.ValueString()
			}
			if v, ok := attrs["grace_period_hours"].(types.Int64); ok {
				config.GracePeriodHours = int(v.ValueInt64())
			}
			actionConfigs = append(actionConfigs, config)
		}
	}

	action := graphScheduledAction{
		RuleName:                      "NonCompliance",
		ScheduledActionConfigurations: actionConfigs,
	}
	payload := map[string]interface{}{"deviceComplianceScheduledActionForRules": []graphScheduledAction{action}}
	return r.client.Post(ctx, fmt.Sprintf("/deviceManagement/deviceCompliancePolicies/%s/scheduleActionsForRules", policyID), payload, nil)
}

func (r *DeviceCompliancePolicyResource) applyComplianceAssignments(ctx context.Context, policyID string, assignments types.List) error {
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
	return r.client.Post(ctx, fmt.Sprintf("/deviceManagement/deviceCompliancePolicies/%s/assign", policyID), payload, nil)
}

func (r *DeviceCompliancePolicyResource) readComplianceAssignments(ctx context.Context, policyID string) (types.List, error) {
	var result graphAssignmentsResponse
	err := r.client.Get(ctx, fmt.Sprintf("/deviceManagement/deviceCompliancePolicies/%s/assignments", policyID), &result)
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
