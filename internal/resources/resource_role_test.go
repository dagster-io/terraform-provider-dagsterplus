package resources

import (
	"context"
	"testing"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestModelToRole_Basic(t *testing.T) {
	ctx := context.Background()

	permsSet := types.SetValueMust(types.StringType, []attr.Value{
		types.StringValue("edit_alerts"),
		types.StringValue("edit_secrets"),
	})

	model := RoleResourceModel{
		ID:          types.StringValue("role-id-1"),
		Name:        types.StringValue("my-role"),
		Description: types.StringValue("A test role"),
		Icon:        types.StringValue("star"),
		RoleType:    types.StringValue("deployment"),
		Permissions: permsSet,
	}

	role, diags := modelToRole(ctx, model)
	if diags.HasError() {
		t.Fatalf("modelToRole returned diags: %v", diags)
	}

	if role.ID != "role-id-1" {
		t.Errorf("expected ID role-id-1, got %q", role.ID)
	}
	if role.Name != "my-role" {
		t.Errorf("expected Name my-role, got %q", role.Name)
	}
	if role.Description != "A test role" {
		t.Errorf("expected Description 'A test role', got %q", role.Description)
	}
	if role.Icon != "star" {
		t.Errorf("expected Icon star, got %q", role.Icon)
	}
	if role.RoleType != "deployment" {
		t.Errorf("expected RoleType deployment, got %q", role.RoleType)
	}
	if len(role.Permissions) != 2 {
		t.Fatalf("expected 2 permissions, got %d", len(role.Permissions))
	}
	if !containsString(role.Permissions, "edit_alerts") {
		t.Error("expected edit_alerts in permissions")
	}
	if !containsString(role.Permissions, "edit_secrets") {
		t.Error("expected edit_secrets in permissions")
	}
}

func TestModelToRole_OrgType(t *testing.T) {
	ctx := context.Background()

	permsSet := types.SetValueMust(types.StringType, []attr.Value{
		types.StringValue("manage_billing"),
	})

	model := RoleResourceModel{
		ID:          types.StringValue(""),
		Name:        types.StringValue("org-role"),
		Description: types.StringValue(""),
		Icon:        types.StringValue(""),
		RoleType:    types.StringValue("organization"),
		Permissions: permsSet,
	}

	role, diags := modelToRole(ctx, model)
	if diags.HasError() {
		t.Fatalf("modelToRole returned diags: %v", diags)
	}

	if role.RoleType != "organization" {
		t.Errorf("expected RoleType organization, got %q", role.RoleType)
	}
}

func TestRoleToModel_Basic(t *testing.T) {
	ctx := context.Background()

	clientRole := &client.Role{
		ID:          "r-abc",
		Name:        "admin-role",
		Description: "Full admin",
		Icon:        "shield",
		RoleType:    "deployment",
		Permissions: []string{"edit_alerts", "start_and_stop_runs", "wipe_assets"},
	}

	var model RoleResourceModel
	diags := RoleToModel(ctx, clientRole, &model)
	if diags.HasError() {
		t.Fatalf("RoleToModel returned diags: %v", diags)
	}

	if model.ID.ValueString() != "r-abc" {
		t.Errorf("expected ID r-abc, got %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "admin-role" {
		t.Errorf("expected Name admin-role, got %q", model.Name.ValueString())
	}
	if model.Description.ValueString() != "Full admin" {
		t.Errorf("expected Description 'Full admin', got %q", model.Description.ValueString())
	}
	if model.Icon.ValueString() != "shield" {
		t.Errorf("expected Icon shield, got %q", model.Icon.ValueString())
	}
	if model.RoleType.ValueString() != "deployment" {
		t.Errorf("expected RoleType deployment, got %q", model.RoleType.ValueString())
	}
	if model.Permissions.IsNull() || model.Permissions.IsUnknown() {
		t.Fatal("expected non-null, non-unknown Permissions set")
	}

	var perms []string
	model.Permissions.ElementsAs(ctx, &perms, false)
	if len(perms) != 3 {
		t.Fatalf("expected 3 permissions, got %d", len(perms))
	}
	if !containsString(perms, "edit_alerts") {
		t.Error("expected edit_alerts in permissions")
	}
}

func TestRoleRoundTrip(t *testing.T) {
	ctx := context.Background()

	original := &client.Role{
		ID:          "round-trip-role",
		Name:        "rt-role",
		Description: "round trip test",
		Icon:        "",
		RoleType:    "organization",
		Permissions: []string{"manage_billing", "edit_users_and_teams"},
	}

	var model RoleResourceModel
	diags := RoleToModel(ctx, original, &model)
	if diags.HasError() {
		t.Fatalf("RoleToModel diags: %v", diags)
	}

	role, diags := modelToRole(ctx, model)
	if diags.HasError() {
		t.Fatalf("modelToRole diags: %v", diags)
	}

	if role.ID != original.ID {
		t.Errorf("ID mismatch: want %q got %q", original.ID, role.ID)
	}
	if role.Name != original.Name {
		t.Errorf("Name mismatch: want %q got %q", original.Name, role.Name)
	}
	if role.Description != original.Description {
		t.Errorf("Description mismatch: want %q got %q", original.Description, role.Description)
	}
	if role.RoleType != original.RoleType {
		t.Errorf("RoleType mismatch: want %q got %q", original.RoleType, role.RoleType)
	}
	if len(role.Permissions) != len(original.Permissions) {
		t.Errorf("Permissions count mismatch: want %d got %d", len(original.Permissions), len(role.Permissions))
	}
	for _, p := range original.Permissions {
		if !containsString(role.Permissions, p) {
			t.Errorf("missing permission %q in round-trip result", p)
		}
	}
}

func TestRoleToModel_EmptyPermissions(t *testing.T) {
	ctx := context.Background()

	clientRole := &client.Role{
		ID:          "r-empty",
		Name:        "empty-role",
		RoleType:    "deployment",
		Permissions: []string{},
	}

	var model RoleResourceModel
	diags := RoleToModel(ctx, clientRole, &model)
	if diags.HasError() {
		t.Fatalf("RoleToModel returned diags: %v", diags)
	}

	if model.Permissions.IsNull() || model.Permissions.IsUnknown() {
		t.Fatal("expected non-null Permissions for empty slice")
	}
	var perms []string
	model.Permissions.ElementsAs(ctx, &perms, false)
	if len(perms) != 0 {
		t.Errorf("expected 0 permissions, got %d", len(perms))
	}
}
