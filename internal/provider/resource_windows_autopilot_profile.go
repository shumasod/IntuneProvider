package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shumasod/IntuneProvider/internal/client"
)

var _ resource.Resource = &WindowsAutopilotProfileResource{}

// WindowsAutopilotProfileResource manages Windows Autopilot deployment profiles in Intune.
type WindowsAutopilotProfileResource struct {
	client *client.Client
}

// WindowsAutopilotProfileModel describes the resource data model.
type WindowsAutopilotProfileModel struct {
	ID                    types.String `tfsdk:"id"`
	DisplayName           types.String `tfsdk:"display_name"`
	Description           types.String `tfsdk:"description"`
	Language              types.String `tfsdk:"language"`
	ExtractHardwareHash   types.Bool   `tfsdk:"extract_hardware_hash"`
	DeviceNameTemplate    types.String `tfsdk:"device_name_template"`
	DeviceType            types.String `tfsdk:"device_type"`
	EnableWhiteGlove      types.Bool   `tfsdk:"enable_white_glove"`
	HidePrivacySettings   types.Bool   `tfsdk:"hide_privacy_settings"`
	HideEULA              types.Bool   `tfsdk:"hide_eula"`
	HideChangeAccountOpts types.Bool   `tfsdk:"hide_change_account_options"`
}

// graphWindowsAutopilotProfile is the Graph API payload for Autopilot profiles.
type graphWindowsAutopilotProfile struct {
	ODataType              string                 `json:"@odata.type"`
	ID                     string                 `json:"id,omitempty"`
	DisplayName            string                 `json:"displayName"`
	Description            string                 `json:"description,omitempty"`
	Language               string                 `json:"language,omitempty"`
	ExtractHardwareHash    bool                   `json:"extractHardwareHash,omitempty"`
	DeviceNameTemplate     string                 `json:"deviceNameTemplate,omitempty"`
	DeviceType             string                 `json:"deviceType,omitempty"`
	EnableWhiteGlove       bool                   `json:"enableWhiteGlove,omitempty"`
	OutOfBoxExperience     graphOOBESettings      `json:"outOfBoxExperienceSettings,omitempty"`
}

type graphOOBESettings struct {
	HidePrivacySettings      bool `json:"hidePrivacySettings,omitempty"`
	HideEULA                 bool `json:"hideEULA,omitempty"`
	HideChangeAccountOptions bool `json:"hideChangeAccountOptions,omitempty"`
}

func NewWindowsAutopilotProfileResource() resource.Resource {
	return &WindowsAutopilotProfileResource{}
}

func (r *WindowsAutopilotProfileResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_windows_autopilot_profile"
}

func (r *WindowsAutopilotProfileResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Windows Autopilot deployment profile in Microsoft Intune.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the Autopilot profile.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"display_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The display name of the Autopilot profile.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Description of the Autopilot profile.",
			},
			"language": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Language/locale for the OOBE experience (e.g. 'os-default').",
			},
			"extract_hardware_hash": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether to extract hardware hash during OOBE.",
			},
			"device_name_template": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Device name template (e.g. 'CORP-%SERIAL%'). Max 15 chars.",
			},
			"device_type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Device type: 'windowsPc' or 'holoLens'.",
			},
			"enable_white_glove": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Enable pre-provisioned deployment (White Glove).",
			},
			"hide_privacy_settings": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Hide privacy settings in OOBE.",
			},
			"hide_eula": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Hide EULA in OOBE.",
			},
			"hide_change_account_options": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Hide change account options in OOBE sign-in.",
			},
		},
	}
}

func (r *WindowsAutopilotProfileResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *WindowsAutopilotProfileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data WindowsAutopilotProfileModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := r.toGraph(&data)
	var result graphWindowsAutopilotProfile
	if err := r.client.Post(ctx, "/deviceManagement/windowsAutopilotDeploymentProfiles", payload, &result); err != nil {
		resp.Diagnostics.AddError("Create Autopilot Profile Error", err.Error())
		return
	}

	r.fromGraph(&result, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WindowsAutopilotProfileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data WindowsAutopilotProfileModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result graphWindowsAutopilotProfile
	err := r.client.Get(ctx, "/deviceManagement/windowsAutopilotDeploymentProfiles/"+data.ID.ValueString(), &result)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Autopilot Profile Error", err.Error())
		return
	}

	r.fromGraph(&result, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WindowsAutopilotProfileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data WindowsAutopilotProfileModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := r.toGraph(&data)
	if err := r.client.Patch(ctx, "/deviceManagement/windowsAutopilotDeploymentProfiles/"+data.ID.ValueString(), payload); err != nil {
		resp.Diagnostics.AddError("Update Autopilot Profile Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WindowsAutopilotProfileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data WindowsAutopilotProfileModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, "/deviceManagement/windowsAutopilotDeploymentProfiles/"+data.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Delete Autopilot Profile Error", err.Error())
	}
}

func (r *WindowsAutopilotProfileResource) toGraph(data *WindowsAutopilotProfileModel) *graphWindowsAutopilotProfile {
	return &graphWindowsAutopilotProfile{
		ODataType:           "#microsoft.graph.azureADWindowsAutopilotDeploymentProfile",
		DisplayName:         data.DisplayName.ValueString(),
		Description:         data.Description.ValueString(),
		Language:            data.Language.ValueString(),
		ExtractHardwareHash: data.ExtractHardwareHash.ValueBool(),
		DeviceNameTemplate:  data.DeviceNameTemplate.ValueString(),
		DeviceType:          data.DeviceType.ValueString(),
		EnableWhiteGlove:    data.EnableWhiteGlove.ValueBool(),
		OutOfBoxExperience: graphOOBESettings{
			HidePrivacySettings:      data.HidePrivacySettings.ValueBool(),
			HideEULA:                 data.HideEULA.ValueBool(),
			HideChangeAccountOptions: data.HideChangeAccountOpts.ValueBool(),
		},
	}
}

func (r *WindowsAutopilotProfileResource) fromGraph(result *graphWindowsAutopilotProfile, data *WindowsAutopilotProfileModel) {
	data.ID = types.StringValue(result.ID)
	data.DisplayName = types.StringValue(result.DisplayName)
	data.Description = types.StringValue(result.Description)
	data.Language = types.StringValue(result.Language)
	data.ExtractHardwareHash = types.BoolValue(result.ExtractHardwareHash)
	data.DeviceNameTemplate = types.StringValue(result.DeviceNameTemplate)
	data.DeviceType = types.StringValue(result.DeviceType)
	data.EnableWhiteGlove = types.BoolValue(result.EnableWhiteGlove)
	data.HidePrivacySettings = types.BoolValue(result.OutOfBoxExperience.HidePrivacySettings)
	data.HideEULA = types.BoolValue(result.OutOfBoxExperience.HideEULA)
	data.HideChangeAccountOpts = types.BoolValue(result.OutOfBoxExperience.HideChangeAccountOptions)
}
