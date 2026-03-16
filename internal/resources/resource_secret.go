package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &secretResource{}
var _ resource.ResourceWithImportState = &secretResource{}

func NewSecretResource() resource.Resource {
	return &secretResource{}
}

type secretResource struct {
	client *client.Client
}

type SecretResourceModel struct {
	ID                            types.String `tfsdk:"id"`
	SecretName                    types.String `tfsdk:"secret_name"`
	SecretValue                   types.String `tfsdk:"secret_value"`
	FullDeploymentScope           types.Bool   `tfsdk:"full_deployment_scope"`
	AllBranchDeploymentsScope     types.Bool   `tfsdk:"all_branch_deployments_scope"`
	SpecificBranchDeploymentScope types.String `tfsdk:"specific_branch_deployment_scope"`
	LocalDeploymentScope          types.Bool   `tfsdk:"local_deployment_scope"`
	LocationNames                 types.List   `tfsdk:"location_names"`
}

func (r *secretResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

func (r *secretResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dagster+ environment secret.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The secret ID assigned by Dagster+.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"secret_name": schema.StringAttribute{
				Description: "The name of the secret (e.g. MY_API_KEY).",
				Required:    true,
			},
			"secret_value": schema.StringAttribute{
				Description: "The value of the secret.",
				Required:    true,
				Sensitive:   true,
			},
			"full_deployment_scope": schema.BoolAttribute{
				Description: "Whether the secret is available to the full deployment.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"all_branch_deployments_scope": schema.BoolAttribute{
				Description: "Whether the secret is available to all branch deployments.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"specific_branch_deployment_scope": schema.StringAttribute{
				Description: "A specific branch deployment name to scope the secret to.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"local_deployment_scope": schema.BoolAttribute{
				Description: "Whether the secret is available to local deployments.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"location_names": schema.ListAttribute{
				Description: "Code location names the secret is scoped to. Empty means all locations.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *secretResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *secretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SecretResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.CreateSecret(ctx, secretFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error creating secret", err.Error())
		return
	}

	state, diags := SecretToModel(ctx, s)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *secretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SecretResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.GetSecretByID(ctx, state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading secret", err.Error())
		return
	}

	newState, diags := SecretToModel(ctx, s)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

func (r *secretResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SecretResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state SecretResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	secret := secretFromModel(plan)
	secret.ID = state.ID.ValueString()

	s, err := r.client.UpdateSecret(ctx, secret)
	if err != nil {
		resp.Diagnostics.AddError("Error updating secret", err.Error())
		return
	}

	newState, diags := SecretToModel(ctx, s)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

func (r *secretResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SecretResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteSecret(ctx, state.ID.ValueString()); err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "PythonError") {
			return
		}
		resp.Diagnostics.AddError("Error deleting secret", err.Error())
	}
}

func (r *secretResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	s, err := r.client.GetSecretByID(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing secret", err.Error())
		return
	}

	state, diags := SecretToModel(ctx, s)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func secretFromModel(m SecretResourceModel) client.Secret {
	var locationNames []string
	if !m.LocationNames.IsNull() && !m.LocationNames.IsUnknown() {
		m.LocationNames.ElementsAs(context.Background(), &locationNames, false)
	}
	return client.Secret{
		SecretName:                    m.SecretName.ValueString(),
		SecretValue:                   m.SecretValue.ValueString(),
		FullDeploymentScope:           m.FullDeploymentScope.ValueBool(),
		AllBranchDeploymentsScope:     m.AllBranchDeploymentsScope.ValueBool(),
		SpecificBranchDeploymentScope: m.SpecificBranchDeploymentScope.ValueString(),
		LocalDeploymentScope:          m.LocalDeploymentScope.ValueBool(),
		LocationNames:                 locationNames,
	}
}

func SecretToModel(ctx context.Context, s *client.Secret) (SecretResourceModel, diag.Diagnostics) {
	locs, diags := types.ListValueFrom(ctx, types.StringType, s.LocationNames)
	return SecretResourceModel{
		ID:                            types.StringValue(s.ID),
		SecretName:                    types.StringValue(s.SecretName),
		SecretValue:                   types.StringValue(s.SecretValue),
		FullDeploymentScope:           types.BoolValue(s.FullDeploymentScope),
		AllBranchDeploymentsScope:     types.BoolValue(s.AllBranchDeploymentsScope),
		SpecificBranchDeploymentScope: types.StringValue(s.SpecificBranchDeploymentScope),
		LocalDeploymentScope:          types.BoolValue(s.LocalDeploymentScope),
		LocationNames:                 locs,
	}, diags
}
