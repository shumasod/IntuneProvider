package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shumasod/IntuneProvider/internal/client"
)

var _ datasource.DataSource = &DeviceCompliancePolicyDataSource{}

// DeviceCompliancePolicyDataSource reads an existing Intune device compliance policy.
type DeviceCompliancePolicyDataSource struct {
	client *client.Client
}

// graphCompliancePolicyListResponse wraps the Graph API list response for compliance policies.
type graphCompliancePolicyListResponse struct {
	Value []graphDeviceCompliancePolicy `json:"value"`
}

func NewDeviceCompliancePolicyDataSource() datasource.DataSource {
	return &DeviceCompliancePolicyDataSource{}
}

func (d *DeviceCompliancePolicyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_compliance_policy"
}

func (d *DeviceCompliancePolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
Use this data source to look up an existing Intune **Device Compliance Policy** by display name or ID.

## Example Usage

` + "```hcl" + `
data "intune_device_compliance_policy" "existing" {
  display_name = "Windows 10 Basic Compliance"
}

output "compliance_policy_id" {
  value = data.intune_device_compliance_policy.existing.id
}
` + "```",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The unique identifier of the compliance policy. Either `id` or `display_name` must be specified.",
			},
			"display_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The display name of the compliance policy. Either `id` or `display_name` must be specified.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Description of the compliance policy.",
			},
			"odata_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The OData type of the compliance policy.",
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

func (d *DeviceCompliancePolicyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DeviceCompliancePolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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

	var policy graphDeviceCompliancePolicy

	if !config.ID.IsNull() && config.ID.ValueString() != "" {
		// Fetch by ID directly
		err := d.client.Get(ctx, "/deviceManagement/deviceCompliancePolicies/"+config.ID.ValueString(), &policy)
		if err != nil {
			resp.Diagnostics.AddError("Failed to read device compliance policy by ID", err.Error())
			return
		}
	} else {
		// Fetch by display name via $filter
		displayName := config.DisplayName.ValueString()
		path := fmt.Sprintf("/deviceManagement/deviceCompliancePolicies?$filter=displayName eq '%s'", displayName)
		var listResp graphCompliancePolicyListResponse
		if err := d.client.Get(ctx, path, &listResp); err != nil {
			resp.Diagnostics.AddError("Failed to list device compliance policies", err.Error())
			return
		}
		if len(listResp.Value) == 0 {
			resp.Diagnostics.AddError("Not found", fmt.Sprintf("No device compliance policy found with display name '%s'.", displayName))
			return
		}
		if len(listResp.Value) > 1 {
			resp.Diagnostics.AddError("Ambiguous result", fmt.Sprintf("Multiple device compliance policies found with display name '%s'. Use 'id' instead.", displayName))
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
