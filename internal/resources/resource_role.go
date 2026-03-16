package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &roleResource{}
var _ resource.ResourceWithImportState = &roleResource{}

// allCustomRolePermissions is the full set of valid permission values.
var allCustomRolePermissions = []string{
	"delete_runs",
	"edit_alerts",
	"edit_all_catalog_views",
	"edit_code_locations",
	"edit_concurrency_limits",
	"edit_custom_roles",
	"edit_deployment_permissions",
	"edit_deployment_settings",
	"edit_dynamic_partitions",
	"edit_external_asset_connections",
	"edit_insights_metrics",
	"edit_issues",
	"edit_secrets",
	"edit_sensor_cursors",
	"edit_users_and_teams",
	"manage_billing",
	"manage_branch_deployments",
	"manage_full_deployments",
	"manage_service_users",
	"manage_sso_and_scim",
	"read_and_edit_agent_tokens",
	"read_and_edit_all_user_tokens",
	"read_audit_log",
	"read_secret_values",
	"redeploy_code_locations",
	"report_asset_events",
	"start_and_stop_runs",
	"toggle_schedules",
	"toggle_sensors",
	"wipe_assets",
}

func NewRoleResource() resource.Resource {
	return &roleResource{}
}

type roleResource struct {
	client *client.Client
}

type RoleResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Icon        types.String `tfsdk:"icon"`
	RoleType    types.String `tfsdk:"role_type"`
	Permissions types.Set    `tfsdk:"permissions"`
}

func (r *roleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (r *roleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dagster+ custom role.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The role ID assigned by Dagster+.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the role.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "A human-readable description of the role.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"icon": schema.StringAttribute{
				Description: "Icon name for the role (e.g. an emoji).",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"role_type": schema.StringAttribute{
				Description: "The scope of the role: deployment or organization. Changing this forces a new resource.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("deployment", "organization"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"permissions": schema.SetAttribute{
				Description: "The permissions granted by this role. At least one permission is required.",
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.OneOf(allCustomRolePermissions...),
					),
				},
			},
		},
	}
}

func (r *roleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = c
}

func (r *roleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RoleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	role, diags := modelToRole(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateRole(ctx, role)
	if err != nil {
		resp.Diagnostics.AddError("Error creating role", err.Error())
		return
	}

	resp.Diagnostics.Append(RoleToModel(ctx, created, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *roleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	role, err := r.client.GetRole(ctx, state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading role", err.Error())
		return
	}

	resp.Diagnostics.Append(RoleToModel(ctx, role, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *roleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RoleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	role, diags := modelToRole(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.UpdateRole(ctx, role)
	if err != nil {
		resp.Diagnostics.AddError("Error updating role", err.Error())
		return
	}

	resp.Diagnostics.Append(RoleToModel(ctx, updated, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *roleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteRole(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting role", err.Error())
	}
}

func (r *roleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	role, err := r.client.GetRole(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing role", err.Error())
		return
	}

	var state RoleResourceModel
	resp.Diagnostics.Append(RoleToModel(ctx, role, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func modelToRole(ctx context.Context, model RoleResourceModel) (client.Role, diag.Diagnostics) {
	var perms []string
	diags := model.Permissions.ElementsAs(ctx, &perms, false)
	return client.Role{
		ID:          model.ID.ValueString(),
		Name:        model.Name.ValueString(),
		Description: model.Description.ValueString(),
		Icon:        model.Icon.ValueString(),
		RoleType:    model.RoleType.ValueString(),
		Permissions: perms,
	}, diags
}

func RoleToModel(ctx context.Context, role *client.Role, model *RoleResourceModel) diag.Diagnostics {
	model.ID = types.StringValue(role.ID)
	model.Name = types.StringValue(role.Name)
	model.Description = types.StringValue(role.Description)
	model.Icon = types.StringValue(role.Icon)
	model.RoleType = types.StringValue(role.RoleType)

	elems := make([]attr.Value, len(role.Permissions))
	for i, p := range role.Permissions {
		elems[i] = types.StringValue(p)
	}
	var diags diag.Diagnostics
	model.Permissions, diags = types.SetValue(types.StringType, elems)
	return diags
}
