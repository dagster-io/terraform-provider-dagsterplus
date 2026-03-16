package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &codeLocationFromDocumentResource{}
var _ resource.ResourceWithImportState = &codeLocationFromDocumentResource{}

func NewCodeLocationFromDocumentResource() resource.Resource {
	return &codeLocationFromDocumentResource{}
}

type codeLocationFromDocumentResource struct {
	client *client.Client
}

type codeLocationFromDocumentResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Deployment types.String `tfsdk:"deployment"`
	Document   types.String `tfsdk:"document"`
	Name       types.String `tfsdk:"name"`
}

func (r *codeLocationFromDocumentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_code_location_from_document"
}

func (r *codeLocationFromDocumentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dagster+ code location defined by a raw JSON configuration document. " +
			"Use the dagsterplus_configuration_document data source to convert a YAML document to JSON.",
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
			"document": schema.StringAttribute{
				Description: "The code location configuration as a JSON document. " +
					"Must include a 'locationName' key. " +
					"Changing the locationName within the document requires replacing this resource.",
				Required: true,
			},
			"name": schema.StringAttribute{
				Description: "The code location name, extracted from the document.",
				Computed:    true,
			},
		},
	}
}

func (r *codeLocationFromDocumentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *codeLocationFromDocumentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan codeLocationFromDocumentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, name, err := client.ParseCodeLocationDocument(plan.Document.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid document", err.Error())
		return
	}

	loc, err := r.client.AddCodeLocation(ctx, plan.Deployment.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Error creating code location", err.Error())
		return
	}

	docJSON, err := client.CodeLocationToDocument(loc)
	if err != nil {
		resp.Diagnostics.AddError("Error serializing code location", err.Error())
		return
	}

	plan.ID = types.StringValue(plan.Deployment.ValueString() + "/" + name)
	plan.Name = types.StringValue(name)
	plan.Document = types.StringValue(docJSON)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *codeLocationFromDocumentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state codeLocationFromDocumentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	loc, err := r.client.GetCodeLocation(ctx, state.Deployment.ValueString(), state.Name.ValueString())
	if err != nil {
		// Location removed out-of-band; let Terraform detect the drift.
		resp.State.RemoveResource(ctx)
		return
	}

	docJSON, err := client.CodeLocationToDocument(loc)
	if err != nil {
		resp.Diagnostics.AddError("Error serializing code location", err.Error())
		return
	}

	state.Document = types.StringValue(docJSON)
	state.Name = types.StringValue(loc.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *codeLocationFromDocumentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan codeLocationFromDocumentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state codeLocationFromDocumentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, newName, err := client.ParseCodeLocationDocument(plan.Document.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid document", err.Error())
		return
	}

	// If locationName changed, delete the old one first.
	if newName != state.Name.ValueString() {
		if err := r.client.DeleteCodeLocation(ctx, state.Deployment.ValueString(), state.Name.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error deleting old code location", err.Error())
			return
		}
	}

	loc, err := r.client.UpdateCodeLocation(ctx, plan.Deployment.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Error updating code location", err.Error())
		return
	}

	docJSON, err := client.CodeLocationToDocument(loc)
	if err != nil {
		resp.Diagnostics.AddError("Error serializing code location", err.Error())
		return
	}

	plan.ID = types.StringValue(plan.Deployment.ValueString() + "/" + newName)
	plan.Name = types.StringValue(newName)
	plan.Document = types.StringValue(docJSON)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *codeLocationFromDocumentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state codeLocationFromDocumentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteCodeLocation(ctx, state.Deployment.ValueString(), state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting code location", err.Error())
	}
}

func (r *codeLocationFromDocumentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Expected ID in the form '{deployment}/{name}'.",
		)
		return
	}

	deployment := parts[0]
	name := parts[1]

	loc, err := r.client.GetCodeLocation(ctx, deployment, name)
	if err != nil {
		resp.Diagnostics.AddError("Error reading code location", err.Error())
		return
	}

	docJSON, err := client.CodeLocationToDocument(loc)
	if err != nil {
		resp.Diagnostics.AddError("Error serializing code location", err.Error())
		return
	}

	state := codeLocationFromDocumentResourceModel{
		ID:         types.StringValue(req.ID),
		Deployment: types.StringValue(deployment),
		Name:       types.StringValue(name),
		Document:   types.StringValue(docJSON),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
