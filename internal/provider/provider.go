// Package provider implements the Terraform provider for Microsoft Intune.
// It authenticates via Azure AD service principal (client credentials) and
// exposes resources and data sources for managing Intune policies.
package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shumasod/IntuneProvider/internal/client"
)

// Ensure IntuneProvider satisfies various provider interfaces.
var _ provider.Provider = &IntuneProvider{}
var _ provider.ProviderWithFunctions = &IntuneProvider{}

// IntuneProvider defines the provider implementation.
type IntuneProvider struct {
	version string
}

// IntuneProviderModel describes the provider data model.
type IntuneProviderModel struct {
	TenantID     types.String `tfsdk:"tenant_id"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
}

// New returns a function that creates a new provider instance.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &IntuneProvider{version: version}
	}
}

// Metadata returns the provider type name and version.
func (p *IntuneProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "intune"
	resp.Version = p.version
}

// Schema defines the provider-level configuration attributes.
// All attributes can also be set via environment variables.
func (p *IntuneProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
The **Intune** provider enables Terraform to manage Microsoft Intune resources via
the Microsoft Graph API. It authenticates using an Azure AD service principal
(client credentials flow).

## Authentication

Set the following provider attributes or the corresponding environment variables:

| Attribute       | Environment Variable     |
|-----------------|--------------------------|
| ` + "`tenant_id`" + `     | ` + "`INTUNE_TENANT_ID`" + `     |
| ` + "`client_id`" + `     | ` + "`INTUNE_CLIENT_ID`" + `     |
| ` + "`client_secret`" + ` | ` + "`INTUNE_CLIENT_SECRET`" + ` |

## Required Permissions (Microsoft Graph)

The service principal needs the following **Application** permissions:

- ` + "`DeviceManagementConfiguration.ReadWrite.All`" + `
- ` + "`DeviceManagementApps.ReadWrite.All`" + `
`,
		Attributes: map[string]schema.Attribute{
			"tenant_id": schema.StringAttribute{
				MarkdownDescription: "Azure AD tenant ID. Can also be set via `INTUNE_TENANT_ID`.",
				Optional:            true,
			},
			"client_id": schema.StringAttribute{
				MarkdownDescription: "Service principal (application) client ID. Can also be set via `INTUNE_CLIENT_ID`.",
				Optional:            true,
			},
			"client_secret": schema.StringAttribute{
				MarkdownDescription: "Service principal client secret. Can also be set via `INTUNE_CLIENT_SECRET`.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

// Configure initialises the shared Graph API client and stores it in the provider data.
func (p *IntuneProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config IntuneProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID := resolveValue(config.TenantID, "INTUNE_TENANT_ID")
	clientID := resolveValue(config.ClientID, "INTUNE_CLIENT_ID")
	clientSecret := resolveValue(config.ClientSecret, "INTUNE_CLIENT_SECRET")

	if tenantID == "" {
		resp.Diagnostics.AddError("Missing tenant_id", "tenant_id must be set in the provider block or via INTUNE_TENANT_ID.")
	}
	if clientID == "" {
		resp.Diagnostics.AddError("Missing client_id", "client_id must be set in the provider block or via INTUNE_CLIENT_ID.")
	}
	if clientSecret == "" {
		resp.Diagnostics.AddError("Missing client_secret", "client_secret must be set in the provider block or via INTUNE_CLIENT_SECRET.")
	}
	if resp.Diagnostics.HasError() {
		return
	}

	c, err := client.NewClient(tenantID, clientID, clientSecret)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Intune client", err.Error())
		return
	}

	resp.DataSourceData = c
	resp.ResourceData = c
}

// Resources returns the list of resources the provider supports.
func (p *IntuneProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDeviceConfigurationPolicyResource,
		NewDeviceCompliancePolicyResource,
		NewAppProtectionPolicyIOSResource,
		NewAppProtectionPolicyAndroidResource,
		NewWindowsAutopilotProfileResource,
		NewDeviceGroupResource,
	}
}

// DataSources returns the list of data sources the provider supports.
func (p *IntuneProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDeviceConfigurationPolicyDataSource,
		NewDeviceCompliancePolicyDataSource,
		NewManagedDeviceDataSource,
	}
}

// Functions returns provider-defined functions (none for now).
func (p *IntuneProvider) Functions(_ context.Context) []func() function.Function {
	return []func() function.Function{}
}

// resolveValue returns the typed string value if set, otherwise falls back to the environment variable.
func resolveValue(v types.String, envKey string) string {
	if !v.IsNull() && !v.IsUnknown() && v.ValueString() != "" {
		return v.ValueString()
	}
	return os.Getenv(envKey)
}
