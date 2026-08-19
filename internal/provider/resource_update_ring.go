package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shumasod/IntuneProvider/internal/client"
)

var _ resource.Resource = &WindowsUpdateRingResource{}

// WindowsUpdateRingResource manages Windows Update for Business rings in Intune.
type WindowsUpdateRingResource struct {
	client *client.Client
}

// WindowsUpdateRingModel describes the resource data model.
type WindowsUpdateRingModel struct {
	ID                            types.String `tfsdk:"id"`
	DisplayName                   types.String `tfsdk:"display_name"`
	Description                   types.String `tfsdk:"description"`
	QualityUpdatesDeferralPeriod  types.Int64  `tfsdk:"quality_updates_deferral_period_days"`
	FeatureUpdatesDeferralPeriod  types.Int64  `tfsdk:"feature_updates_deferral_period_days"`
	AutomaticUpdateMode           types.String `tfsdk:"automatic_update_mode"`
	UpdateNotificationLevel       types.String `tfsdk:"update_notification_level"`
	BusinessReadyUpdatesOnly      types.String `tfsdk:"business_ready_updates_only"`
	MicrosoftUpdateServiceAllowed types.Bool   `tfsdk:"microsoft_update_service_allowed"`
	DriverUpdatesDisabled         types.Bool   `tfsdk:"driver_updates_disabled"`
}

type graphWindowsUpdateRing struct {
	ODataType                     string `json:"@odata.type,omitempty"`
	ID                            string `json:"id,omitempty"`
	DisplayName                   string `json:"displayName"`
	Description                   string `json:"description,omitempty"`
	QualityUpdatesDeferralPeriod  int    `json:"qualityUpdatesDeferralPeriodInDays,omitempty"`
	FeatureUpdatesDeferralPeriod  int    `json:"featureUpdatesDeferralPeriodInDays,omitempty"`
	AutomaticUpdateMode           string `json:"automaticUpdateMode,omitempty"`
	UpdateNotificationLevel       string `json:"updateNotificationLevel,omitempty"`
	BusinessReadyUpdatesOnly      string `json:"businessReadyUpdatesOnly,omitempty"`
	MicrosoftUpdateServiceAllowed bool   `json:"microsoftUpdateServiceAllowed,omitempty"`
	DriverUpdatesDisabled         bool   `json:"driversExcluded,omitempty"`
}

func NewWindowsUpdateRingResource() resource.Resource {
	return &WindowsUpdateRingResource{}
}

func (r *WindowsUpdateRingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_windows_update_ring"
}

func (r *WindowsUpdateRingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Windows Update for Business ring in Microsoft Intune.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the update ring.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"display_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The display name of the update ring.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Description of the update ring.",
			},
			"quality_updates_deferral_period_days": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
				MarkdownDescription: "Number of days to defer quality updates (0-30).",
			},
			"feature_updates_deferral_period_days": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
				MarkdownDescription: "Number of days to defer feature updates (0-365).",
			},
			"automatic_update_mode": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Automatic update mode: 'userDefined', 'notifyDownload', 'autoInstallAtMaintenanceTime', 'autoInstallAndRebootAtMaintenanceTime', 'autoInstallAndRebootAtScheduledTime', 'autoInstallAndRebootWithoutEndUserControl', 'windowsDefault'.",
			},
			"update_notification_level": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Update notification level: 'notConfigured', 'defaultNotifications', 'restartWarningsOnly', 'disableAllNotifications'.",
			},
			"business_ready_updates_only": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Update channel: 'userDefined', 'all', 'businessReadyOnly', 'windowsInsiderBuildFast', 'windowsInsiderBuildSlow', 'windowsInsiderBuildRelease'.",
			},
			"microsoft_update_service_allowed": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Allow Microsoft Update Service.",
			},
			"driver_updates_disabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Exclude drivers from Windows Update.",
			},
		},
	}
}

func (r *WindowsUpdateRingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *WindowsUpdateRingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data WindowsUpdateRingModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := r.toGraph(&data)
	var result graphWindowsUpdateRing
	if err := r.client.Post(ctx, "/deviceManagement/deviceConfigurations", payload, &result); err != nil {
		resp.Diagnostics.AddError("Create Update Ring Error", err.Error())
		return
	}

	r.fromGraph(&result, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WindowsUpdateRingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data WindowsUpdateRingModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result graphWindowsUpdateRing
	if err := r.client.Get(ctx, "/deviceManagement/deviceConfigurations/"+data.ID.ValueString(), &result); err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Update Ring Error", err.Error())
		return
	}

	r.fromGraph(&result, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WindowsUpdateRingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data WindowsUpdateRingModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := r.toGraph(&data)
	if err := r.client.Patch(ctx, "/deviceManagement/deviceConfigurations/"+data.ID.ValueString(), payload); err != nil {
		resp.Diagnostics.AddError("Update Ring Update Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WindowsUpdateRingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data WindowsUpdateRingModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Delete(ctx, "/deviceManagement/deviceConfigurations/"+data.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Delete Update Ring Error", err.Error())
	}
}

func (r *WindowsUpdateRingResource) toGraph(data *WindowsUpdateRingModel) *graphWindowsUpdateRing {
	return &graphWindowsUpdateRing{
		ODataType:                    "#microsoft.graph.windowsUpdateForBusinessConfiguration",
		DisplayName:                  data.DisplayName.ValueString(),
		Description:                  data.Description.ValueString(),
		QualityUpdatesDeferralPeriod: int(data.QualityUpdatesDeferralPeriod.ValueInt64()),
		FeatureUpdatesDeferralPeriod: int(data.FeatureUpdatesDeferralPeriod.ValueInt64()),
		AutomaticUpdateMode:          data.AutomaticUpdateMode.ValueString(),
		UpdateNotificationLevel:      data.UpdateNotificationLevel.ValueString(),
		BusinessReadyUpdatesOnly:     data.BusinessReadyUpdatesOnly.ValueString(),
		MicrosoftUpdateServiceAllowed: data.MicrosoftUpdateServiceAllowed.ValueBool(),
		DriverUpdatesDisabled:        data.DriverUpdatesDisabled.ValueBool(),
	}
}

func (r *WindowsUpdateRingResource) fromGraph(result *graphWindowsUpdateRing, data *WindowsUpdateRingModel) {
	data.ID = types.StringValue(result.ID)
	data.DisplayName = types.StringValue(result.DisplayName)
	data.Description = types.StringValue(result.Description)
	data.QualityUpdatesDeferralPeriod = types.Int64Value(int64(result.QualityUpdatesDeferralPeriod))
	data.FeatureUpdatesDeferralPeriod = types.Int64Value(int64(result.FeatureUpdatesDeferralPeriod))
	data.AutomaticUpdateMode = types.StringValue(result.AutomaticUpdateMode)
	data.UpdateNotificationLevel = types.StringValue(result.UpdateNotificationLevel)
	data.BusinessReadyUpdatesOnly = types.StringValue(result.BusinessReadyUpdatesOnly)
	data.MicrosoftUpdateServiceAllowed = types.BoolValue(result.MicrosoftUpdateServiceAllowed)
	data.DriverUpdatesDisabled = types.BoolValue(result.DriverUpdatesDisabled)
}
