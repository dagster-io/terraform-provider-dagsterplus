package resources

import (
	"context"
	"testing"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSecretFromModel_Basic(t *testing.T) {
	locs := types.ListValueMust(types.StringType, []attr.Value{
		types.StringValue("loc-a"),
		types.StringValue("loc-b"),
	})

	m := SecretResourceModel{
		ID:                            types.StringValue("secret-id-1"),
		SecretName:                    types.StringValue("MY_API_KEY"),
		SecretValue:                   types.StringValue("supersecret"),
		FullDeploymentScope:           types.BoolValue(true),
		AllBranchDeploymentsScope:     types.BoolValue(false),
		SpecificBranchDeploymentScope: types.StringValue("feature-branch"),
		LocalDeploymentScope:          types.BoolValue(true),
		LocationNames:                 locs,
	}

	s := secretFromModel(context.Background(), m)

	if s.SecretName != "MY_API_KEY" {
		t.Errorf("expected SecretName MY_API_KEY, got %q", s.SecretName)
	}
	if s.SecretValue != "supersecret" {
		t.Errorf("expected SecretValue supersecret, got %q", s.SecretValue)
	}
	if !s.FullDeploymentScope {
		t.Error("expected FullDeploymentScope true")
	}
	if s.AllBranchDeploymentsScope {
		t.Error("expected AllBranchDeploymentsScope false")
	}
	if s.SpecificBranchDeploymentScope != "feature-branch" {
		t.Errorf("expected SpecificBranchDeploymentScope feature-branch, got %q", s.SpecificBranchDeploymentScope)
	}
	if !s.LocalDeploymentScope {
		t.Error("expected LocalDeploymentScope true")
	}
	if len(s.LocationNames) != 2 {
		t.Fatalf("expected 2 location names, got %d", len(s.LocationNames))
	}
	if s.LocationNames[0] != "loc-a" || s.LocationNames[1] != "loc-b" {
		t.Errorf("unexpected location names: %v", s.LocationNames)
	}
	// ID is not mapped by secretFromModel (it is set separately on update)
	if s.ID != "" {
		t.Errorf("expected empty ID from secretFromModel, got %q", s.ID)
	}
}

func TestSecretFromModel_NullLocationNames(t *testing.T) {
	m := SecretResourceModel{
		SecretName:    types.StringValue("KEY"),
		SecretValue:   types.StringValue("val"),
		LocationNames: types.ListNull(types.StringType),
	}

	s := secretFromModel(context.Background(), m)
	if len(s.LocationNames) != 0 {
		t.Errorf("expected empty LocationNames for null list, got %v", s.LocationNames)
	}
}

func TestSecretToModel_Basic(t *testing.T) {
	ctx := context.Background()
	cs := &client.Secret{
		ID:                            "secret-xyz",
		SecretName:                    "DB_PASSWORD",
		SecretValue:                   "hunter2",
		FullDeploymentScope:           true,
		AllBranchDeploymentsScope:     false,
		SpecificBranchDeploymentScope: "",
		LocalDeploymentScope:          false,
		LocationNames:                 []string{"repo1", "repo2"},
	}

	model, diags := SecretToModel(ctx, cs)
	if diags.HasError() {
		t.Fatalf("SecretToModel returned diagnostics: %v", diags)
	}

	if model.ID.ValueString() != "secret-xyz" {
		t.Errorf("expected ID secret-xyz, got %q", model.ID.ValueString())
	}
	if model.SecretName.ValueString() != "DB_PASSWORD" {
		t.Errorf("expected SecretName DB_PASSWORD, got %q", model.SecretName.ValueString())
	}
	if model.SecretValue.ValueString() != "hunter2" {
		t.Errorf("expected SecretValue hunter2, got %q", model.SecretValue.ValueString())
	}
	if !model.FullDeploymentScope.ValueBool() {
		t.Error("expected FullDeploymentScope true")
	}
	if model.AllBranchDeploymentsScope.ValueBool() {
		t.Error("expected AllBranchDeploymentsScope false")
	}
	if model.SpecificBranchDeploymentScope.ValueString() != "" {
		t.Errorf("expected empty SpecificBranchDeploymentScope, got %q", model.SpecificBranchDeploymentScope.ValueString())
	}
	if model.LocalDeploymentScope.ValueBool() {
		t.Error("expected LocalDeploymentScope false")
	}

	var locs []string
	model.LocationNames.ElementsAs(ctx, &locs, false)
	if len(locs) != 2 || locs[0] != "repo1" || locs[1] != "repo2" {
		t.Errorf("unexpected LocationNames: %v", locs)
	}
}

func TestSecretToModel_EmptyLocationNames(t *testing.T) {
	ctx := context.Background()
	cs := &client.Secret{
		ID:            "s1",
		SecretName:    "EMPTY",
		LocationNames: []string{},
	}

	model, diags := SecretToModel(ctx, cs)
	if diags.HasError() {
		t.Fatalf("SecretToModel returned diagnostics: %v", diags)
	}
	if model.LocationNames.IsNull() {
		t.Error("expected non-null LocationNames for empty slice")
	}
	if model.LocationNames.IsUnknown() {
		t.Error("expected non-unknown LocationNames")
	}
}

func TestSecretRoundTrip(t *testing.T) {
	ctx := context.Background()
	original := &client.Secret{
		ID:                            "round-trip-id",
		SecretName:                    "ROUND_TRIP",
		SecretValue:                   "rt-value",
		FullDeploymentScope:           false,
		AllBranchDeploymentsScope:     true,
		SpecificBranchDeploymentScope: "my-branch",
		LocalDeploymentScope:          true,
		LocationNames:                 []string{"alpha"},
	}

	model, diags := SecretToModel(ctx, original)
	if diags.HasError() {
		t.Fatalf("SecretToModel diags: %v", diags)
	}

	s := secretFromModel(ctx, model)
	s.ID = original.ID // secretFromModel does not carry the ID

	if s.ID != original.ID {
		t.Errorf("ID mismatch: want %q got %q", original.ID, s.ID)
	}
	if s.SecretName != original.SecretName {
		t.Errorf("SecretName mismatch: want %q got %q", original.SecretName, s.SecretName)
	}
	if s.SecretValue != original.SecretValue {
		t.Errorf("SecretValue mismatch: want %q got %q", original.SecretValue, s.SecretValue)
	}
	if s.FullDeploymentScope != original.FullDeploymentScope {
		t.Errorf("FullDeploymentScope mismatch: want %v got %v", original.FullDeploymentScope, s.FullDeploymentScope)
	}
	if s.AllBranchDeploymentsScope != original.AllBranchDeploymentsScope {
		t.Errorf("AllBranchDeploymentsScope mismatch: want %v got %v", original.AllBranchDeploymentsScope, s.AllBranchDeploymentsScope)
	}
	if s.SpecificBranchDeploymentScope != original.SpecificBranchDeploymentScope {
		t.Errorf("SpecificBranchDeploymentScope mismatch: want %q got %q", original.SpecificBranchDeploymentScope, s.SpecificBranchDeploymentScope)
	}
	if s.LocalDeploymentScope != original.LocalDeploymentScope {
		t.Errorf("LocalDeploymentScope mismatch: want %v got %v", original.LocalDeploymentScope, s.LocalDeploymentScope)
	}
	if len(s.LocationNames) != 1 || s.LocationNames[0] != "alpha" {
		t.Errorf("LocationNames mismatch: want [alpha] got %v", s.LocationNames)
	}
}
