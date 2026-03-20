package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure codeLocationResource satisfies the resource.Resource interface.
var _ resource.Resource = &codeLocationResource{}
var _ resource.ResourceWithImportState = &codeLocationResource{}
var _ resource.ResourceWithConfigValidators = &codeLocationResource{}

// NewCodeLocationResource returns a new code location resource.
func NewCodeLocationResource() resource.Resource {
	return &codeLocationResource{}
}

// codeLocationResource defines the resource implementation.
type codeLocationResource struct {
	client *client.Client
}

// CodeSourceModel represents the code_source nested block.
type CodeSourceModel struct {
	PythonFile  types.String `tfsdk:"python_file"`
	PackageName types.String `tfsdk:"package_name"`
	ModuleName  types.String `tfsdk:"module_name"`
}

// GitModel represents the git nested block.
type GitModel struct {
	CommitHash types.String `tfsdk:"commit_hash"`
	URL        types.String `tfsdk:"url"`
}

// codeLocationResourceModel describes the resource data model.
type codeLocationResourceModel struct {
	ID               types.String    `tfsdk:"id"`
	DeploymentName   types.String    `tfsdk:"deployment"`
	Name             types.String    `tfsdk:"name"`
	Image            types.String    `tfsdk:"image"`
	CodeSource       CodeSourceModel `tfsdk:"code_source"`
	WorkingDirectory types.String    `tfsdk:"working_directory"`
	ExecutablePath   types.String    `tfsdk:"executable_path"`
	Attribute        types.String    `tfsdk:"attribute"`
	AgentQueue       types.String    `tfsdk:"agent_queue"`
	Git              *GitModel       `tfsdk:"git"`
}

func (r *codeLocationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_code_location"
}

func (r *codeLocationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dagster+ code location within a deployment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier in the form '{deployment}/{name}'.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"deployment": schema.StringAttribute{
				Description: "The name of the deployment this code location belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the code location.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"image": schema.StringAttribute{
				Description: "The Docker image to use for this code location. Exactly one of image or git must be specified.",
				Optional:    true,
			},
			"working_directory": schema.StringAttribute{
				Description: "The working directory inside the container.",
				Optional:    true,
				Computed:    true,
			},
			"executable_path": schema.StringAttribute{
				Description: "Path to the Python executable inside the container.",
				Optional:    true,
				Computed:    true,
			},
			"attribute": schema.StringAttribute{
				Description: "Python attribute containing the Definitions object.",
				Optional:    true,
				Computed:    true,
			},
			"agent_queue": schema.StringAttribute{
				Description: "The agent queue to use for this code location.",
				Optional:    true,
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"code_source": schema.SingleNestedBlock{
				Description: "Specifies where Dagster finds the code.",
				Attributes: map[string]schema.Attribute{
					"python_file": schema.StringAttribute{
						Description: "Path to a Python file containing the repository. Mutually exclusive with package_name and module_name.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.ConflictsWith(
								path.MatchRelative().AtParent().AtName("package_name"),
								path.MatchRelative().AtParent().AtName("module_name"),
							),
						},
					},
					"package_name": schema.StringAttribute{
						Description: "Python package name containing the repository. Mutually exclusive with python_file and module_name.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.ConflictsWith(
								path.MatchRelative().AtParent().AtName("python_file"),
								path.MatchRelative().AtParent().AtName("module_name"),
							),
						},
					},
					"module_name": schema.StringAttribute{
						Description: "Python module name containing the repository. Mutually exclusive with python_file and package_name.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.ConflictsWith(
								path.MatchRelative().AtParent().AtName("python_file"),
								path.MatchRelative().AtParent().AtName("package_name"),
							),
						},
					},
				},
			},
			"git": schema.SingleNestedBlock{
				Description: "Git-based code location. Exactly one of image or git must be specified.",
				Attributes: map[string]schema.Attribute{
					"commit_hash": schema.StringAttribute{
						Description: "The git commit hash to use.",
						Optional:    true,
					},
					"url": schema.StringAttribute{
						Description: "The URL of the git repository.",
						Optional:    true,
					},
				},
			},
		},
	}
}

func (r *codeLocationResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("image"),
			path.MatchRoot("git"),
		),
	}
}

func (r *codeLocationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func modelToInput(plan codeLocationResourceModel) client.CodeLocationInput {
	input := client.CodeLocationInput{
		Name:             plan.Name.ValueString(),
		Image:            plan.Image.ValueString(),
		WorkingDirectory: plan.WorkingDirectory.ValueString(),
		ExecutablePath:   plan.ExecutablePath.ValueString(),
		Attribute:        plan.Attribute.ValueString(),
		AgentQueue:       plan.AgentQueue.ValueString(),
		CodeSource: client.CodeSource{
			PythonFile:  plan.CodeSource.PythonFile.ValueString(),
			PackageName: plan.CodeSource.PackageName.ValueString(),
			ModuleName:  plan.CodeSource.ModuleName.ValueString(),
		},
	}
	if plan.Git != nil {
		input.CommitHash = plan.Git.CommitHash.ValueString()
		input.URL = plan.Git.URL.ValueString()
	}
	return input
}

// CodeSourceStringVal converts an API string to a types.String, using null for
// empty strings so that Optional (non-Computed) schema fields round-trip correctly.
func CodeSourceStringVal(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

func applyCodeLocation(state *codeLocationResourceModel, cl *client.CodeLocation) {
	state.ID = types.StringValue(cl.ID)
	state.Name = types.StringValue(cl.Name)
	if cl.Image != "" {
		state.Image = types.StringValue(cl.Image)
	} else {
		state.Image = types.StringNull()
	}
	state.WorkingDirectory = types.StringValue(cl.WorkingDirectory)
	state.ExecutablePath = types.StringValue(cl.ExecutablePath)
	state.Attribute = types.StringValue(cl.Attribute)
	state.AgentQueue = types.StringValue(cl.AgentQueue)
	state.CodeSource = CodeSourceModel{
		PythonFile:  CodeSourceStringVal(cl.CodeSource.PythonFile),
		PackageName: CodeSourceStringVal(cl.CodeSource.PackageName),
		ModuleName:  CodeSourceStringVal(cl.CodeSource.ModuleName),
	}
	if cl.CommitHash != "" || cl.URL != "" {
		state.Git = &GitModel{
			CommitHash: types.StringValue(cl.CommitHash),
			URL:        types.StringValue(cl.URL),
		}
	} else {
		state.Git = nil
	}
}

func requireGitFields(git *GitModel, diagnostics interface{ AddError(string, string) }) bool {
	if git == nil {
		return true
	}
	ok := true
	if git.CommitHash.IsNull() || git.CommitHash.ValueString() == "" {
		diagnostics.AddError("Missing git.commit_hash", "git.commit_hash is required when the git block is specified.")
		ok = false
	}
	if git.URL.IsNull() || git.URL.ValueString() == "" {
		diagnostics.AddError("Missing git.url", "git.url is required when the git block is specified.")
		ok = false
	}
	return ok
}

func (r *codeLocationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan codeLocationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !requireGitFields(plan.Git, &resp.Diagnostics) {
		return
	}

	cl, err := r.client.AddCodeLocation(ctx, plan.DeploymentName.ValueString(), modelToInput(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error creating code location", err.Error())
		return
	}

	applyCodeLocation(&plan, cl)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *codeLocationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state codeLocationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cl, err := r.client.GetCodeLocation(ctx, state.DeploymentName.ValueString(), state.Name.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading code location", err.Error())
		return
	}

	// Always update fields the API reliably returns.
	state.ID = types.StringValue(cl.ID)
	state.Name = types.StringValue(cl.Name)
	if cl.Image != "" {
		state.Image = types.StringValue(cl.Image)
	} else {
		state.Image = types.StringNull()
	}

	// For optional fields and code_source, only overwrite state when the API
	// returns a non-empty value. The serialized metadata may omit or rename
	// these fields, so preserving the stored state avoids spurious drift.
	if cl.WorkingDirectory != "" {
		state.WorkingDirectory = types.StringValue(cl.WorkingDirectory)
	}
	if cl.ExecutablePath != "" {
		state.ExecutablePath = types.StringValue(cl.ExecutablePath)
	}
	if cl.Attribute != "" {
		state.Attribute = types.StringValue(cl.Attribute)
	}
	if cl.AgentQueue != "" {
		state.AgentQueue = types.StringValue(cl.AgentQueue)
	}
	if cl.CommitHash != "" || cl.URL != "" {
		state.Git = &GitModel{
			CommitHash: types.StringValue(cl.CommitHash),
			URL:        types.StringValue(cl.URL),
		}
	}
	cs := cl.CodeSource
	if cs.PythonFile != "" || cs.PackageName != "" || cs.ModuleName != "" {
		state.CodeSource = CodeSourceModel{
			PythonFile:  CodeSourceStringVal(cs.PythonFile),
			PackageName: CodeSourceStringVal(cs.PackageName),
			ModuleName:  CodeSourceStringVal(cs.ModuleName),
		}
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *codeLocationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan codeLocationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !requireGitFields(plan.Git, &resp.Diagnostics) {
		return
	}

	cl, err := r.client.UpdateCodeLocation(ctx, plan.DeploymentName.ValueString(), modelToInput(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error updating code location", err.Error())
		return
	}

	applyCodeLocation(&plan, cl)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *codeLocationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state codeLocationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteCodeLocation(ctx, state.DeploymentName.ValueString(), state.Name.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError("Error deleting code location", err.Error())
	}
}

func (r *codeLocationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID format: "{deployment}/{location_name}"
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Import ID must be in the format '{deployment}/{location_name}'.",
		)
		return
	}

	deployment, name := parts[0], parts[1]

	cl, err := r.client.GetCodeLocation(ctx, deployment, name)
	if err != nil {
		resp.Diagnostics.AddError("Error importing code location", err.Error())
		return
	}

	state := codeLocationResourceModel{
		DeploymentName: types.StringValue(deployment),
	}
	applyCodeLocation(&state, cl)

	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
