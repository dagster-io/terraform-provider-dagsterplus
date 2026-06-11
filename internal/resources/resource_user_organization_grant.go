package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &userOrganizationGrantResource{}
var _ resource.ResourceWithImportState = &userOrganizationGrantResource{}

func NewUserOrganizationGrantResource() resource.Resource {
	return &userOrganizationGrantResource{}
}

type userOrganizationGrantResource struct {
	client *client.Client
}

type userOrganizationGrantResourceModel struct {
	ID           types.String `tfsdk:"id"`
	UserID       types.String `tfsdk:"user_id"`
	Grant        types.String `tfsdk:"grant"`
	CustomRoleID types.String `tfsdk:"custom_role_id"`
}

func (r *userOrganizationGrantResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_organization_grant"
}

func (r *userOrganizationGrantResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an organization-level permission grant for a Dagster+ user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier (the user_id).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.StringAttribute{
				Description: "The ID of the user to grant access.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"grant": schema.StringAttribute{
				Description: withEnumValues(orgGrantDescription, orgGrantLevels),
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(orgGrantLevels...),
					stringvalidator.ExactlyOneOf(path.MatchRoot("custom_role_id")),
				},
			},
			"custom_role_id": schema.StringAttribute{
				Description: "The ID of a custom role to assign. Exactly one of grant or custom_role_id must be set.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("grant")),
				},
			},
		},
	}
}

func (r *userOrganizationGrantResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *userOrganizationGrantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userOrganizationGrantResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.GetUser(ctx, plan.UserID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error resolving user", err.Error())
		return
	}

	grantStr, customRoleID := resolveGrantFields(plan.Grant, plan.CustomRoleID)
	result, err := r.client.SetUserGrant(ctx, client.UserGrant{
		UserID:          user.ID,
		Email:           user.Email,
		DeploymentScope: "organization",
		Grant:           grantStr,
		CustomRoleID:    customRoleID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error setting organization grant", err.Error())
		return
	}

	plan.ID = types.StringValue(plan.UserID.ValueString())
	plan.Grant, plan.CustomRoleID = grantFieldsFromAPI(result.Grant, result.CustomRoleID)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *userOrganizationGrantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userOrganizationGrantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	grant, err := r.client.GetUserGrant(ctx, state.UserID.ValueString(), "organization", 0)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading organization grant", err.Error())
		return
	}

	state.Grant, state.CustomRoleID = grantFieldsFromAPI(grant.Grant, grant.CustomRoleID)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *userOrganizationGrantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan userOrganizationGrantResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.GetUser(ctx, plan.UserID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error resolving user", err.Error())
		return
	}

	grantStr, customRoleID := resolveGrantFields(plan.Grant, plan.CustomRoleID)
	result, err := r.client.SetUserGrant(ctx, client.UserGrant{
		UserID:          user.ID,
		Email:           user.Email,
		DeploymentScope: "organization",
		Grant:           grantStr,
		CustomRoleID:    customRoleID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating organization grant", err.Error())
		return
	}

	plan.Grant, plan.CustomRoleID = grantFieldsFromAPI(result.Grant, result.CustomRoleID)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *userOrganizationGrantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userOrganizationGrantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.GetUser(ctx, state.UserID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError("Error resolving user", err.Error())
		return
	}

	if err := r.client.DeleteUserGrant(ctx, user.Email, "organization", 0); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError("Error removing organization grant", err.Error())
	}
}

func (r *userOrganizationGrantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	userID := req.ID

	grant, err := r.client.GetUserGrant(ctx, userID, "organization", 0)
	if err != nil {
		resp.Diagnostics.AddError("Error reading organization grant", err.Error())
		return
	}

	grantVal, customRoleIDVal := grantFieldsFromAPI(grant.Grant, grant.CustomRoleID)
	state := userOrganizationGrantResourceModel{
		ID:           types.StringValue(userID),
		UserID:       types.StringValue(userID),
		Grant:        grantVal,
		CustomRoleID: customRoleIDVal,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
