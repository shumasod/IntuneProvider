package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shumasod/IntuneProvider/internal/client"
)

var _ datasource.DataSource = &DeviceConfigurationPolicyDataSource{}

// DeviceConfigurationPolicyDataSource reads an existing Intune device configuration policy.
type DeviceConfigurationPolicyDataSource struct {
	client *client.Client
}

// graphDeviceConfigListResponse wraps the Graph API list response.
type graphDeviceConfigListResponse struct {
	Value []graphDeviceConfigurationPolicy `json:"value"`
}

func NewDeviceConfigurationPolicyDataSource() datasource.DataSource {
	return &DeviceConfigurationPolicyDataSource{}
}

func (d *DeviceConfigurationPolicyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_configuration_policy"
}

func (d *DeviceConfigurationPolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
Use this data source to look up an existing Intune **Device Configuration Policy** by display name or ID.

## Example Usage

` + "```hcl" + `
data "intune_device_configuration_policy" "existing" {
  display_name = "Windows Defender Configuration"
}

output "policy_id" {
  value = data.intune_device_configuration_policy.existing.id
}
` + "```",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The unique identifier of the device configuration policy. Either `id` or `display_name` must be specified.",
			},
			"display_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The display name of the device configuration policy. Either `id` or `display_name` must be specified.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Description of the device configuration policy.",
			},
			"odata_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The OData type of the configuration policy.",
			},
			"last_modified_date_time": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ISO 8601 timestamp when the policy was last modified.",
			},
			"created_date_time": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ISO 8601 timestamp when the policy was created.",
			},
		},
	}
}

func (d *DeviceConfigurationPolicyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *DeviceConfigurationPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config struct {
		ID          types.String `tfsdk:"id"`
		DisplayName types.String `tfsdk:"display_name"`
		Description types.String `tfsdk:"description"`
		ODataType   types.String `tfsdk:"odata_type"`
		LastModifiedDateTime types.String `tfsdk:"last_modified_date_time"`
		CreatedDateTime      types.String `tfsdk:"created_date_time"`
	}

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.ID.IsNull() && config.DisplayName.IsNull() {
		resp.Diagnostics.AddError("Missing required attribute", "Either 'id' or 'display_name' must be specified.")
		return
	}

	var policy graphDeviceConfigurationPolicy

	if !config.ID.IsNull() && config.ID.ValueString() != "" {
		// Fetch by ID directly
		err := d.client.Get(ctx, "/deviceManagement/deviceConfigurations/"+config.ID.ValueString(), &policy)
		if err != nil {
			resp.Diagnostics.AddError("Failed to read device configuration policy by ID", err.Error())
			return
		}
	} else {
		// Fetch by display name via $filter
		displayName := config.DisplayName.ValueString()
		path := fmt.Sprintf("/deviceManagement/deviceConfigurations?$filter=displayName eq '%s'", displayName)
		var listResp graphDeviceConfigListResponse
		if err := d.client.Get(ctx, path, &listResp); err != nil {
			resp.Diagnostics.AddError("Failed to list device configuration policies", err.Error())
			return
		}
		if len(listResp.Value) == 0 {
			resp.Diagnostics.AddError("Not found", fmt.Sprintf("No device configuration policy found with display name '%s'.", displayName))
			return
		}
		if len(listResp.Value) > 1 {
			resp.Diagnostics.AddError("Ambiguous result", fmt.Sprintf("Multiple device configuration policies found with display name '%s'. Use 'id' instead.", displayName))
			return
		}
		policy = listResp.Value[0]
	}

	config.ID = types.StringValue(policy.ID)
	config.DisplayName = types.StringValue(policy.DisplayName)
	config.Description = types.StringValue(policy.Description)
	config.ODataType = types.StringValue(policy.ODataType)
	config.LastModifiedDateTime = types.StringValue(policy.LastModifiedDateTime)
	config.CreatedDateTime = types.StringValue(policy.CreatedDateTime)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
