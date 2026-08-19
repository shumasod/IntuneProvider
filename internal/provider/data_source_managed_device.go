package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shumasod/IntuneProvider/internal/client"
)

var _ datasource.DataSource = &ManagedDeviceDataSource{}

// ManagedDeviceDataSource reads an Intune managed device by its ID or serial number.
type ManagedDeviceDataSource struct {
	client *client.Client
}

// ManagedDeviceDataModel describes the data source data model.
type ManagedDeviceDataModel struct {
	ID               types.String `tfsdk:"id"`
	DeviceName       types.String `tfsdk:"device_name"`
	SerialNumber     types.String `tfsdk:"serial_number"`
	OperatingSystem  types.String `tfsdk:"operating_system"`
	OSVersion        types.String `tfsdk:"os_version"`
	Manufacturer     types.String `tfsdk:"manufacturer"`
	Model            types.String `tfsdk:"model"`
	UserID           types.String `tfsdk:"user_id"`
	UserPrincipalName types.String `tfsdk:"user_principal_name"`
	ComplianceState  types.String `tfsdk:"compliance_state"`
	ManagementState  types.String `tfsdk:"management_state"`
	EnrolledDateTime types.String `tfsdk:"enrolled_date_time"`
	LastSyncDateTime types.String `tfsdk:"last_sync_date_time"`
}

type graphManagedDevice struct {
	ID                string `json:"id"`
	DeviceName        string `json:"deviceName"`
	SerialNumber      string `json:"serialNumber"`
	OperatingSystem   string `json:"operatingSystem"`
	OSVersion         string `json:"osVersion"`
	Manufacturer      string `json:"manufacturer"`
	Model             string `json:"model"`
	UserID            string `json:"userId"`
	UserPrincipalName string `json:"userPrincipalName"`
	ComplianceState   string `json:"complianceState"`
	ManagementState   string `json:"managementState"`
	EnrolledDateTime  string `json:"enrolledDateTime"`
	LastSyncDateTime  string `json:"lastSyncDateTime"`
}

func NewManagedDeviceDataSource() datasource.DataSource {
	return &ManagedDeviceDataSource{}
}

func (d *ManagedDeviceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_device"
}

func (d *ManagedDeviceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves information about an Intune managed device.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique identifier of the managed device.",
			},
			"device_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The name of the device.",
			},
			"serial_number": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The serial number of the device.",
			},
			"operating_system": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The operating system of the device.",
			},
			"os_version": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The OS version of the device.",
			},
			"manufacturer": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The manufacturer of the device.",
			},
			"model": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The model of the device.",
			},
			"user_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the user enrolled to the device.",
			},
			"user_principal_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The UPN of the user enrolled to the device.",
			},
			"compliance_state": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The compliance state of the device.",
			},
			"management_state": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The management state of the device.",
			},
			"enrolled_date_time": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The date and time the device was enrolled.",
			},
			"last_sync_date_time": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The date and time of the last device sync.",
			},
		},
	}
}

func (d *ManagedDeviceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("expected *client.Client, got: %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *ManagedDeviceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ManagedDeviceDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result graphManagedDevice
	if err := d.client.Get(ctx, "/deviceManagement/managedDevices/"+data.ID.ValueString(), &result); err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError("Managed Device Not Found", fmt.Sprintf("device %s not found", data.ID.ValueString()))
			return
		}
		resp.Diagnostics.AddError("Read Managed Device Error", err.Error())
		return
	}

	data.DeviceName = types.StringValue(result.DeviceName)
	data.SerialNumber = types.StringValue(result.SerialNumber)
	data.OperatingSystem = types.StringValue(result.OperatingSystem)
	data.OSVersion = types.StringValue(result.OSVersion)
	data.Manufacturer = types.StringValue(result.Manufacturer)
	data.Model = types.StringValue(result.Model)
	data.UserID = types.StringValue(result.UserID)
	data.UserPrincipalName = types.StringValue(result.UserPrincipalName)
	data.ComplianceState = types.StringValue(result.ComplianceState)
	data.ManagementState = types.StringValue(result.ManagementState)
	data.EnrolledDateTime = types.StringValue(result.EnrolledDateTime)
	data.LastSyncDateTime = types.StringValue(result.LastSyncDateTime)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
