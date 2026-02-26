package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shumasod/IntuneProvider/internal/client"
)

var _ resource.Resource = &AppProtectionPolicyAndroidResource{}
var _ resource.ResourceWithImportState = &AppProtectionPolicyAndroidResource{}

// AppProtectionPolicyAndroidResource manages Intune Android App Protection Policies (MAM).
type AppProtectionPolicyAndroidResource struct {
	client *client.Client
}

// AppProtectionPolicyAndroidModel describes the resource data model.
type AppProtectionPolicyAndroidModel struct {
	ID                             types.String `tfsdk:"id"`
	DisplayName                    types.String `tfsdk:"display_name"`
	Description                    types.String `tfsdk:"description"`
	PeriodOfflineBeforeWipeHours   types.Int64  `tfsdk:"period_offline_before_wipe_hours"`
	PeriodOfflineBeforeAccessCheck types.String `tfsdk:"period_offline_before_access_check"`
	PinRequired                    types.Bool   `tfsdk:"pin_required"`
	MinimumPinLength               types.Int64  `tfsdk:"minimum_pin_length"`
	AllowedDataStorageLocations    types.List   `tfsdk:"allowed_data_storage_locations"`
	DataBackupBlocked              types.Bool   `tfsdk:"data_backup_blocked"`
	DeviceComplianceRequired       types.Bool   `tfsdk:"device_compliance_required"`
	ManagedBrowserToOpenLinksRequired types.Bool `tfsdk:"managed_browser_to_open_links_required"`
	SaveAsBlocked                  types.Bool   `tfsdk:"save_as_blocked"`
	PrintBlocked                   types.Bool   `tfsdk:"print_blocked"`
	FingerprintBlocked             types.Bool   `tfsdk:"fingerprint_blocked"`
	DisableAppEncryptionIfDeviceEncryptionIsEnabled types.Bool `tfsdk:"disable_app_encryption_if_device_encryption_is_enabled"`
	EncryptAppData                 types.Bool   `tfsdk:"encrypt_app_data"`
	MinimumRequiredOSVersion       types.String `tfsdk:"minimum_required_os_version"`
	MinimumWarningOSVersion        types.String `tfsdk:"minimum_warning_os_version"`
	MinimumRequiredAppVersion      types.String `tfsdk:"minimum_required_app_version"`
	ScreenCaptureBlocked           types.Bool   `tfsdk:"screen_capture_blocked"`
	Apps                           types.List   `tfsdk:"apps"`
	Assignments                    types.List   `tfsdk:"assignments"`
	CreatedDateTime                types.String `tfsdk:"created_date_time"`
	LastModifiedDateTime           types.String `tfsdk:"last_modified_date_time"`
}

// graphAndroidManagedAppProtection represents the Graph API payload for Android MAM policies.
type graphAndroidManagedAppProtection struct {
	ODataType                       string   `json:"@odata.type,omitempty"`
	ID                              string   `json:"id,omitempty"`
	DisplayName                     string   `json:"displayName"`
	Description                     string   `json:"description,omitempty"`
	PeriodOfflineBeforeWipeHours    int64    `json:"periodOfflineBeforeWipeInDays,omitempty"`
	PeriodOfflineBeforeAccessCheck  string   `json:"periodOfflineBeforeAccessCheck,omitempty"`
	PinRequired                     bool     `json:"pinRequired,omitempty"`
	MinimumPinLength                int64    `json:"minimumPinLength,omitempty"`
	AllowedDataStorageLocations     []string `json:"allowedDataStorageLocations,omitempty"`
	DataBackupBlocked               bool     `json:"dataBackupBlocked,omitempty"`
	DeviceComplianceRequired        bool     `json:"deviceComplianceRequired,omitempty"`
	ManagedBrowserToOpenLinksRequired bool   `json:"managedBrowserToOpenLinksRequired,omitempty"`
	SaveAsBlocked                   bool     `json:"saveAsBlocked,omitempty"`
	PrintBlocked                    bool     `json:"printBlocked,omitempty"`
	FingerprintBlocked              bool     `json:"fingerprintBlocked,omitempty"`
	DisableAppEncryptionIfDeviceEncryptionIsEnabled bool `json:"disableAppEncryptionIfDeviceEncryptionIsEnabled,omitempty"`
	EncryptAppData                  bool     `json:"encryptAppData,omitempty"`
	MinimumRequiredOSVersion        string   `json:"minimumRequiredOsVersion,omitempty"`
	MinimumWarningOSVersion         string   `json:"minimumWarningOsVersion,omitempty"`
	MinimumRequiredAppVersion       string   `json:"minimumRequiredAppVersion,omitempty"`
	ScreenCaptureBlocked            bool     `json:"screenCaptureBlocked,omitempty"`
	CreatedDateTime                 string   `json:"createdDateTime,omitempty"`
	LastModifiedDateTime            string   `json:"lastModifiedDateTime,omitempty"`
}

func NewAppProtectionPolicyAndroidResource() resource.Resource {
	return &AppProtectionPolicyAndroidResource{}
}

func (r *AppProtectionPolicyAndroidResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app_protection_policy_android"
}

func (r *AppProtectionPolicyAndroidResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
Manages an Intune **Android App Protection Policy** (MAM) via ` + "`deviceAppManagement/androidManagedAppProtections`" + `.

App Protection Policies (APP) protect corporate data within managed apps on Android devices without
requiring full device enrollment. This is the recommended approach for BYOD (Bring Your Own Device) scenarios.

## Example Usage

` + "```hcl" + `
resource "intune_app_protection_policy_android" "teams" {
  display_name        = "Android Teams MAM Policy"
  description         = "Protects Teams data on Android BYOD devices"
  pin_required        = true
  minimum_pin_length  = 6
  data_backup_blocked = true
  save_as_blocked     = true
  screen_capture_blocked = true
  encrypt_app_data    = true
  period_offline_before_wipe_hours = 720
  minimum_required_os_version = "10.0"

  apps = ["com.microsoft.teams"]
  assignments = ["00000000-0000-0000-0000-000000000001"]
}
` + "```",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the Android app protection policy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The display name of the Android app protection policy.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Description of the policy.",
			},
			"period_offline_before_wipe_hours": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(720),
				MarkdownDescription: "Hours the device can be offline before managed app data is wiped. Default: `720` (30 days).",
			},
			"period_offline_before_access_check": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "ISO 8601 duration after which access is rechecked when offline.",
			},
			"pin_required": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Require a PIN to access managed apps. Default: `true`.",
			},
			"minimum_pin_length": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(4),
				MarkdownDescription: "Minimum PIN length for app access. Default: `4`.",
			},
			"allowed_data_storage_locations": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Data storage locations allowed for saving managed app data (e.g., `oneDriveForBusiness`, `sharePoint`).",
			},
			"data_backup_blocked": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Block Android backup of managed app data. Default: `false`.",
			},
			"device_compliance_required": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Require device compliance for app access. Default: `false`.",
			},
			"managed_browser_to_open_links_required": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Require managed browser for web links from managed apps. Default: `false`.",
			},
			"save_as_blocked": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Block users from saving copies of managed data to non-managed locations. Default: `false`.",
			},
			"print_blocked": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Block printing managed data. Default: `false`.",
			},
			"fingerprint_blocked": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Block fingerprint unlock in place of the managed-app PIN. Default: `false`.",
			},
			"disable_app_encryption_if_device_encryption_is_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Skip app-level encryption when device-level encryption is already enabled. Default: `false`.",
			},
			"encrypt_app_data": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Encrypt managed app data at rest. Default: `true`.",
			},
			"minimum_required_os_version": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Minimum Android OS version required (e.g., `10.0`).",
			},
			"minimum_warning_os_version": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Android OS version below which the user receives a warning.",
			},
			"minimum_required_app_version": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Minimum app version required. Apps below this version are blocked.",
			},
			"screen_capture_blocked": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Block screen captures within managed apps. Default: `false`.",
			},
			"apps": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "List of Android app package IDs to target (e.g., `com.microsoft.teams`).",
			},
			"assignments": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "List of Azure AD group IDs to assign this policy to.",
			},
			"created_date_time": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ISO 8601 timestamp when the policy was created.",
			},
			"last_modified_date_time": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ISO 8601 timestamp when the policy was last modified.",
			},
		},
	}
}

func (r *AppProtectionPolicyAndroidResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AppProtectionPolicyAndroidResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AppProtectionPolicyAndroidModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := r.buildPayload(plan)

	var result graphAndroidManagedAppProtection
	err := r.client.Post(ctx, "/deviceAppManagement/androidManagedAppProtections", payload, &result)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Android app protection policy", err.Error())
		return
	}

	plan.ID = types.StringValue(result.ID)
	plan.CreatedDateTime = types.StringValue(result.CreatedDateTime)
	plan.LastModifiedDateTime = types.StringValue(result.LastModifiedDateTime)

	if err := r.applyAndroidApps(ctx, result.ID, plan.Apps); err != nil {
		resp.Diagnostics.AddError("Failed to assign apps to Android MAM policy", err.Error())
		return
	}

	if err := r.applyAndroidAssignments(ctx, result.ID, plan.Assignments); err != nil {
		resp.Diagnostics.AddError("Failed to assign groups to Android MAM policy", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AppProtectionPolicyAndroidResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AppProtectionPolicyAndroidModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result graphAndroidManagedAppProtection
	err := r.client.Get(ctx, "/deviceAppManagement/androidManagedAppProtections/"+state.ID.ValueString(), &result)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read Android app protection policy", err.Error())
		return
	}

	state.DisplayName = types.StringValue(result.DisplayName)
	state.Description = types.StringValue(result.Description)
	state.PinRequired = types.BoolValue(result.PinRequired)
	state.MinimumPinLength = types.Int64Value(result.MinimumPinLength)
	state.DataBackupBlocked = types.BoolValue(result.DataBackupBlocked)
	state.DeviceComplianceRequired = types.BoolValue(result.DeviceComplianceRequired)
	state.ManagedBrowserToOpenLinksRequired = types.BoolValue(result.ManagedBrowserToOpenLinksRequired)
	state.SaveAsBlocked = types.BoolValue(result.SaveAsBlocked)
	state.PrintBlocked = types.BoolValue(result.PrintBlocked)
	state.FingerprintBlocked = types.BoolValue(result.FingerprintBlocked)
	state.DisableAppEncryptionIfDeviceEncryptionIsEnabled = types.BoolValue(result.DisableAppEncryptionIfDeviceEncryptionIsEnabled)
	state.EncryptAppData = types.BoolValue(result.EncryptAppData)
	state.MinimumRequiredOSVersion = types.StringValue(result.MinimumRequiredOSVersion)
	state.MinimumWarningOSVersion = types.StringValue(result.MinimumWarningOSVersion)
	state.MinimumRequiredAppVersion = types.StringValue(result.MinimumRequiredAppVersion)
	state.ScreenCaptureBlocked = types.BoolValue(result.ScreenCaptureBlocked)
	state.CreatedDateTime = types.StringValue(result.CreatedDateTime)
	state.LastModifiedDateTime = types.StringValue(result.LastModifiedDateTime)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *AppProtectionPolicyAndroidResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AppProtectionPolicyAndroidModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := r.buildPayload(plan)

	err := r.client.Patch(ctx, "/deviceAppManagement/androidManagedAppProtections/"+plan.ID.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update Android app protection policy", err.Error())
		return
	}

	if err := r.applyAndroidApps(ctx, plan.ID.ValueString(), plan.Apps); err != nil {
		resp.Diagnostics.AddError("Failed to update apps for Android MAM policy", err.Error())
		return
	}

	if err := r.applyAndroidAssignments(ctx, plan.ID.ValueString(), plan.Assignments); err != nil {
		resp.Diagnostics.AddError("Failed to update group assignments for Android MAM policy", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AppProtectionPolicyAndroidResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AppProtectionPolicyAndroidModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, "/deviceAppManagement/androidManagedAppProtections/"+state.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete Android app protection policy", err.Error())
	}
}

func (r *AppProtectionPolicyAndroidResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var state AppProtectionPolicyAndroidModel
	state.ID = types.StringValue(req.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *AppProtectionPolicyAndroidResource) buildPayload(plan AppProtectionPolicyAndroidModel) graphAndroidManagedAppProtection {
	payload := graphAndroidManagedAppProtection{
		ODataType:                       "#microsoft.graph.androidManagedAppProtection",
		DisplayName:                     plan.DisplayName.ValueString(),
		Description:                     plan.Description.ValueString(),
		PeriodOfflineBeforeWipeHours:    plan.PeriodOfflineBeforeWipeHours.ValueInt64(),
		PeriodOfflineBeforeAccessCheck:  plan.PeriodOfflineBeforeAccessCheck.ValueString(),
		PinRequired:                     plan.PinRequired.ValueBool(),
		MinimumPinLength:                plan.MinimumPinLength.ValueInt64(),
		DataBackupBlocked:               plan.DataBackupBlocked.ValueBool(),
		DeviceComplianceRequired:        plan.DeviceComplianceRequired.ValueBool(),
		ManagedBrowserToOpenLinksRequired: plan.ManagedBrowserToOpenLinksRequired.ValueBool(),
		SaveAsBlocked:                   plan.SaveAsBlocked.ValueBool(),
		PrintBlocked:                    plan.PrintBlocked.ValueBool(),
		FingerprintBlocked:              plan.FingerprintBlocked.ValueBool(),
		DisableAppEncryptionIfDeviceEncryptionIsEnabled: plan.DisableAppEncryptionIfDeviceEncryptionIsEnabled.ValueBool(),
		EncryptAppData:                  plan.EncryptAppData.ValueBool(),
		MinimumRequiredOSVersion:        plan.MinimumRequiredOSVersion.ValueString(),
		MinimumWarningOSVersion:         plan.MinimumWarningOSVersion.ValueString(),
		MinimumRequiredAppVersion:       plan.MinimumRequiredAppVersion.ValueString(),
		ScreenCaptureBlocked:            plan.ScreenCaptureBlocked.ValueBool(),
	}

	if !plan.AllowedDataStorageLocations.IsNull() && !plan.AllowedDataStorageLocations.IsUnknown() {
		for _, v := range plan.AllowedDataStorageLocations.Elements() {
			if sv, ok := v.(types.String); ok {
				payload.AllowedDataStorageLocations = append(payload.AllowedDataStorageLocations, sv.ValueString())
			}
		}
	}

	return payload
}

// applyAndroidApps targets specific Android app package IDs to the MAM policy.
func (r *AppProtectionPolicyAndroidResource) applyAndroidApps(ctx context.Context, policyID string, apps types.List) error {
	if apps.IsNull() || apps.IsUnknown() {
		return nil
	}

	var managedApps []graphManagedMobileApp
	for _, v := range apps.Elements() {
		if sv, ok := v.(types.String); ok {
			managedApps = append(managedApps, graphManagedMobileApp{
				MobileAppIdentifier: graphMobileAppIdentifier{
					ODataType: "#microsoft.graph.androidMobileAppIdentifier",
					PackageID: sv.ValueString(),
				},
			})
		}
	}

	payload := map[string]interface{}{"apps": managedApps}
	return r.client.Post(ctx, fmt.Sprintf("/deviceAppManagement/androidManagedAppProtections/%s/targetApps", policyID), payload, nil)
}

// applyAndroidAssignments assigns Azure AD groups to an Android MAM policy.
func (r *AppProtectionPolicyAndroidResource) applyAndroidAssignments(ctx context.Context, policyID string, assignments types.List) error {
	if assignments.IsNull() || assignments.IsUnknown() {
		return nil
	}

	var mamAssignments []map[string]interface{}
	for _, v := range assignments.Elements() {
		if sv, ok := v.(types.String); ok {
			mamAssignments = append(mamAssignments, map[string]interface{}{
				"target": map[string]interface{}{
					"@odata.type": "#microsoft.graph.groupAssignmentTarget",
					"groupId":     sv.ValueString(),
				},
			})
		}
	}

	payload := map[string]interface{}{"assignments": mamAssignments}
	return r.client.Post(ctx, fmt.Sprintf("/deviceAppManagement/androidManagedAppProtections/%s/assign", policyID), payload, nil)
}
