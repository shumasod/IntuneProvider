package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shumasod/IntuneProvider/internal/client"
)

var _ resource.Resource = &DeviceGroupResource{}

// DeviceGroupResource manages Azure AD groups used for Intune device targeting.
type DeviceGroupResource struct {
	client *client.Client
}

// DeviceGroupModel describes the resource data model.
type DeviceGroupModel struct {
	ID              types.String `tfsdk:"id"`
	DisplayName     types.String `tfsdk:"display_name"`
	Description     types.String `tfsdk:"description"`
	SecurityEnabled types.Bool   `tfsdk:"security_enabled"`
	MailEnabled     types.Bool   `tfsdk:"mail_enabled"`
	MailNickname    types.String `tfsdk:"mail_nickname"`
	GroupTypes      types.List   `tfsdk:"group_types"`
}

type graphGroup struct {
	ID              string   `json:"id,omitempty"`
	DisplayName     string   `json:"displayName"`
	Description     string   `json:"description,omitempty"`
	SecurityEnabled bool     `json:"securityEnabled"`
	MailEnabled     bool     `json:"mailEnabled"`
	MailNickname    string   `json:"mailNickname"`
	GroupTypes      []string `json:"groupTypes,omitempty"`
}

func NewDeviceGroupResource() resource.Resource {
	return &DeviceGroupResource{}
}

func (r *DeviceGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_group"
}

func (r *DeviceGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Azure AD security group for use with Intune device policies.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the group.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"display_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The display name of the group.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Description of the group.",
			},
			"security_enabled": schema.BoolAttribute{
				Required:            true,
				MarkdownDescription: "Whether the group is a security group.",
			},
			"mail_enabled": schema.BoolAttribute{
				Required:            true,
				MarkdownDescription: "Whether the group is mail-enabled.",
			},
			"mail_nickname": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Mail alias for the group.",
			},
			"group_types": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Group types (e.g. 'DynamicMembership', 'Unified').",
			},
		},
	}
}

func (r *DeviceGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DeviceGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DeviceGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var groupTypes []string
	resp.Diagnostics.Append(data.GroupTypes.ElementsAs(ctx, &groupTypes, false)...)

	payload := &graphGroup{
		DisplayName:     data.DisplayName.ValueString(),
		Description:     data.Description.ValueString(),
		SecurityEnabled: data.SecurityEnabled.ValueBool(),
		MailEnabled:     data.MailEnabled.ValueBool(),
		MailNickname:    data.MailNickname.ValueString(),
		GroupTypes:      groupTypes,
	}

	var result graphGroup
	if err := r.client.Post(ctx, "/groups", payload, &result); err != nil {
		resp.Diagnostics.AddError("Create Device Group Error", err.Error())
		return
	}

	data.ID = types.StringValue(result.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeviceGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DeviceGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result graphGroup
	if err := r.client.Get(ctx, "/groups/"+data.ID.ValueString(), &result); err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Device Group Error", err.Error())
		return
	}

	data.DisplayName = types.StringValue(result.DisplayName)
	data.Description = types.StringValue(result.Description)
	data.SecurityEnabled = types.BoolValue(result.SecurityEnabled)
	data.MailEnabled = types.BoolValue(result.MailEnabled)
	data.MailNickname = types.StringValue(result.MailNickname)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeviceGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DeviceGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := &graphGroup{
		DisplayName:  data.DisplayName.ValueString(),
		Description:  data.Description.ValueString(),
		MailNickname: data.MailNickname.ValueString(),
	}

	if err := r.client.Patch(ctx, "/groups/"+data.ID.ValueString(), payload); err != nil {
		resp.Diagnostics.AddError("Update Device Group Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeviceGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DeviceGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Delete(ctx, "/groups/"+data.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Delete Device Group Error", err.Error())
	}
}
