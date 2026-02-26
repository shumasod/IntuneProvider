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

var _ resource.Resource = &AppProtectionPolicyIOSResource{}
var _ resource.ResourceWithImportState = &AppProtectionPolicyIOSResource{}

// AppProtectionPolicyIOSResource manages Intune iOS App Protection Policies (MAM).
type AppProtectionPolicyIOSResource struct {
	client *client.Client
}

// AppProtectionPolicyIOSModel describes the resource data model.
type AppProtectionPolicyIOSModel struct {
	ID                          types.String `tfsdk:"id"`
	DisplayName                 types.String `tfsdk:"display_name"`
	Description                 types.String `tfsdk:"description"`
	PeriodOfflineBeforeWipeHours types.Int64 `tfsdk:"period_offline_before_wipe_hours"`
	PeriodOfflineBeforeAccessCheck types.String `tfsdk:"period_offline_before_access_check"`
	PinRequired                 types.Bool   `tfsdk:"pin_required"`
	MinimumPinLength            types.Int64  `tfsdk:"minimum_pin_length"`
	AllowedDataStorageLocations types.List   `tfsdk:"allowed_data_storage_locations"`
	DataBackupBlocked           types.Bool   `tfsdk:"data_backup_blocked"`
	DeviceComplianceRequired    types.Bool   `tfsdk:"device_compliance_required"`
	ManagedBrowserToOpenLinksRequired types.Bool `tfsdk:"managed_browser_to_open_links_required"`
	SaveAsBlocked               types.Bool   `tfsdk:"save_as_blocked"`
	OrganizerSyncBlocked        types.Bool   `tfsdk:"organizer_sync_blocked"`
	PrintBlocked                types.Bool   `tfsdk:"print_blocked"`
	FingerprintBlocked          types.Bool   `tfsdk:"fingerprint_blocked"`
	DisableAppPinIfDevicePinIsSet types.Bool `tfsdk:"disable_app_pin_if_device_pin_is_set"`
	MinimumRequiredOSVersion    types.String `tfsdk:"minimum_required_os_version"`
	MinimumWarningOSVersion     types.String `tfsdk:"minimum_warning_os_version"`
	MinimumRequiredAppVersion   types.String `tfsdk:"minimum_required_app_version"`
	Apps                        types.List   `tfsdk:"apps"`
	Assignments                 types.List   `tfsdk:"assignments"`
	CreatedDateTime             types.String `tfsdk:"created_date_time"`
	LastModifiedDateTime        types.String `tfsdk:"last_modified_date_time"`
}

// graphIOSManagedAppProtection represents the Graph API payload for iOS MAM policies.
type graphIOSManagedAppProtection struct {
	ODataType                        string   `json:"@odata.type,omitempty"`
	ID                               string   `json:"id,omitempty"`
	DisplayName                      string   `json:"displayName"`
	Description                      string   `json:"description,omitempty"`
	PeriodOfflineBeforeWipeHours     int64    `json:"periodOfflineBeforeWipeInDays,omitempty"`
	PeriodOfflineBeforeAccessCheck   string   `json:"periodOfflineBeforeAccessCheck,omitempty"`
	PinRequired                      bool     `json:"pinRequired,omitempty"`
	MinimumPinLength                 int64    `json:"minimumPinLength,omitempty"`
	AllowedDataStorageLocations      []string `json:"allowedDataStorageLocations,omitempty"`
	DataBackupBlocked                bool     `json:"dataBackupBlocked,omitempty"`
	DeviceComplianceRequired         bool     `json:"deviceComplianceRequired,omitempty"`
	ManagedBrowserToOpenLinksRequired bool    `json:"managedBrowserToOpenLinksRequired,omitempty"`
	SaveAsBlocked                    bool     `json:"saveAsBlocked,omitempty"`
	OrganizerSyncBlocked             bool     `json:"organizerSyncBlocked,omitempty"`
	PrintBlocked                     bool     `json:"printBlocked,omitempty"`
	FingerprintBlocked               bool     `json:"fingerprintBlocked,omitempty"`
	DisableAppPinIfDevicePinIsSet    bool     `json:"disableAppPinIfDevicePinIsSet,omitempty"`
	MinimumRequiredOSVersion         string   `json:"minimumRequiredOsVersion,omitempty"`
	MinimumWarningOSVersion          string   `json:"minimumWarningOsVersion,omitempty"`
	MinimumRequiredAppVersion        string   `json:"minimumRequiredAppVersion,omitempty"`
	CreatedDateTime                  string   `json:"createdDateTime,omitempty"`
	LastModifiedDateTime             string   `json:"lastModifiedDateTime,omitempty"`
}

// graphManagedMobileApp represents an app to target with the MAM policy.
type graphManagedMobileApp struct {
	MobileAppIdentifier graphMobileAppIdentifier `json:"mobileAppIdentifier"`
}

type graphMobileAppIdentifier struct {
	ODataType  string `json:"@odata.type"`
	BundleID   string `json:"bundleId,omitempty"`
	PackageID  string `json:"packageId,omitempty"`
}

func NewAppProtectionPolicyIOSResource() resource.Resource {
	return &AppProtectionPolicyIOSResource{}
}

func (r *AppProtectionPolicyIOSResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app_protection_policy_ios"
}

func (r *AppProtectionPolicyIOSResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
Manages an Intune **iOS App Protection Policy** (MAM) via ` + "`deviceAppManagement/iosManagedAppProtections`" + `.

App Protection Policies (APP) protect corporate data within managed apps without requiring full
device enrollment (MAM-WE). They are commonly used in BYOD scenarios.

## Example Usage

` + "```hcl" + `
resource "intune_app_protection_policy_ios" "outlook" {
  display_name  = "iOS Outlook MAM Policy"
  description   = "Protects Outlook data on iOS devices"
  pin_required  = true
  minimum_pin_length = 6
  data_backup_blocked = true
  save_as_blocked     = true
  print_blocked       = true
  device_compliance_required = false
  period_offline_before_wipe_hours = 720
  minimum_required_os_version = "15.0"

  apps = ["com.microsoft.Outlook"]
  assignments = ["00000000-0000-0000-0000-000000000001"]
}
` + "```",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the iOS app protection policy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The display name of the iOS app protection policy.",
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
				MarkdownDescription: "ISO 8601 duration after which access is rechecked when offline (e.g., `PT720H`).",
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
				MarkdownDescription: "Block iTunes/iCloud backup of managed app data. Default: `false`.",
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
			"organizer_sync_blocked": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Block syncing contacts, calendars, etc. to native apps. Default: `false`.",
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
				MarkdownDescription: "Block Touch ID in place of the managed-app PIN. Default: `false`.",
			},
			"disable_app_pin_if_device_pin_is_set": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Disable the app PIN when a device-level PIN is configured. Default: `false`.",
			},
			"minimum_required_os_version": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Minimum iOS version required to run the managed app (e.g., `15.0`).",
			},
			"minimum_warning_os_version": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "iOS version below which the user receives a warning (e.g., `14.0`).",
			},
			"minimum_required_app_version": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Minimum app version required. Apps below this version are blocked.",
			},
			"apps": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "List of iOS app bundle IDs to target (e.g., `com.microsoft.Outlook`).",
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

func (r *AppProtectionPolicyIOSResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AppProtectionPolicyIOSResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AppProtectionPolicyIOSModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := r.buildPayload(plan)

	var result graphIOSManagedAppProtection
	err := r.client.Post(ctx, "/deviceAppManagement/iosManagedAppProtections", payload, &result)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create iOS app protection policy", err.Error())
		return
	}

	plan.ID = types.StringValue(result.ID)
	plan.CreatedDateTime = types.StringValue(result.CreatedDateTime)
	plan.LastModifiedDateTime = types.StringValue(result.LastModifiedDateTime)

	if err := r.applyApps(ctx, result.ID, plan.Apps); err != nil {
		resp.Diagnostics.AddError("Failed to assign apps to iOS MAM policy", err.Error())
		return
	}

	if err := r.applyMAMAssignments(ctx, "iosManagedAppProtections", result.ID, plan.Assignments); err != nil {
		resp.Diagnostics.AddError("Failed to assign groups to iOS MAM policy", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AppProtectionPolicyIOSResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AppProtectionPolicyIOSModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result graphIOSManagedAppProtection
	err := r.client.Get(ctx, "/deviceAppManagement/iosManagedAppProtections/"+state.ID.ValueString(), &result)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read iOS app protection policy", err.Error())
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
	state.OrganizerSyncBlocked = types.BoolValue(result.OrganizerSyncBlocked)
	state.PrintBlocked = types.BoolValue(result.PrintBlocked)
	state.FingerprintBlocked = types.BoolValue(result.FingerprintBlocked)
	state.DisableAppPinIfDevicePinIsSet = types.BoolValue(result.DisableAppPinIfDevicePinIsSet)
	state.MinimumRequiredOSVersion = types.StringValue(result.MinimumRequiredOSVersion)
	state.MinimumWarningOSVersion = types.StringValue(result.MinimumWarningOSVersion)
	state.MinimumRequiredAppVersion = types.StringValue(result.MinimumRequiredAppVersion)
	state.CreatedDateTime = types.StringValue(result.CreatedDateTime)
	state.LastModifiedDateTime = types.StringValue(result.LastModifiedDateTime)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *AppProtectionPolicyIOSResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AppProtectionPolicyIOSModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := r.buildPayload(plan)

	err := r.client.Patch(ctx, "/deviceAppManagement/iosManagedAppProtections/"+plan.ID.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update iOS app protection policy", err.Error())
		return
	}

	if err := r.applyApps(ctx, plan.ID.ValueString(), plan.Apps); err != nil {
		resp.Diagnostics.AddError("Failed to update apps for iOS MAM policy", err.Error())
		return
	}

	if err := r.applyMAMAssignments(ctx, "iosManagedAppProtections", plan.ID.ValueString(), plan.Assignments); err != nil {
		resp.Diagnostics.AddError("Failed to update group assignments for iOS MAM policy", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AppProtectionPolicyIOSResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AppProtectionPolicyIOSModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, "/deviceAppManagement/iosManagedAppProtections/"+state.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete iOS app protection policy", err.Error())
	}
}

func (r *AppProtectionPolicyIOSResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var state AppProtectionPolicyIOSModel
	state.ID = types.StringValue(req.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *AppProtectionPolicyIOSResource) buildPayload(plan AppProtectionPolicyIOSModel) graphIOSManagedAppProtection {
	payload := graphIOSManagedAppProtection{
		ODataType:                        "#microsoft.graph.iosManagedAppProtection",
		DisplayName:                      plan.DisplayName.ValueString(),
		Description:                      plan.Description.ValueString(),
		PeriodOfflineBeforeWipeHours:     plan.PeriodOfflineBeforeWipeHours.ValueInt64(),
		PeriodOfflineBeforeAccessCheck:   plan.PeriodOfflineBeforeAccessCheck.ValueString(),
		PinRequired:                      plan.PinRequired.ValueBool(),
		MinimumPinLength:                 plan.MinimumPinLength.ValueInt64(),
		DataBackupBlocked:                plan.DataBackupBlocked.ValueBool(),
		DeviceComplianceRequired:         plan.DeviceComplianceRequired.ValueBool(),
		ManagedBrowserToOpenLinksRequired: plan.ManagedBrowserToOpenLinksRequired.ValueBool(),
		SaveAsBlocked:                    plan.SaveAsBlocked.ValueBool(),
		OrganizerSyncBlocked:             plan.OrganizerSyncBlocked.ValueBool(),
		PrintBlocked:                     plan.PrintBlocked.ValueBool(),
		FingerprintBlocked:               plan.FingerprintBlocked.ValueBool(),
		DisableAppPinIfDevicePinIsSet:    plan.DisableAppPinIfDevicePinIsSet.ValueBool(),
		MinimumRequiredOSVersion:         plan.MinimumRequiredOSVersion.ValueString(),
		MinimumWarningOSVersion:          plan.MinimumWarningOSVersion.ValueString(),
		MinimumRequiredAppVersion:        plan.MinimumRequiredAppVersion.ValueString(),
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

// applyApps targets specific iOS app bundle IDs to the MAM policy.
func (r *AppProtectionPolicyIOSResource) applyApps(ctx context.Context, policyID string, apps types.List) error {
	if apps.IsNull() || apps.IsUnknown() {
		return nil
	}

	var managedApps []graphManagedMobileApp
	for _, v := range apps.Elements() {
		if sv, ok := v.(types.String); ok {
			managedApps = append(managedApps, graphManagedMobileApp{
				MobileAppIdentifier: graphMobileAppIdentifier{
					ODataType: "#microsoft.graph.iosMobileAppIdentifier",
					BundleID:  sv.ValueString(),
				},
			})
		}
	}

	payload := map[string]interface{}{"apps": managedApps}
	return r.client.Post(ctx, fmt.Sprintf("/deviceAppManagement/iosManagedAppProtections/%s/targetApps", policyID), payload, nil)
}

// applyMAMAssignments assigns Azure AD groups to a MAM policy.
func (r *AppProtectionPolicyIOSResource) applyMAMAssignments(ctx context.Context, policyPath, policyID string, assignments types.List) error {
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
	return r.client.Post(ctx, fmt.Sprintf("/deviceAppManagement/%s/%s/assign", policyPath, policyID), payload, nil)
}
