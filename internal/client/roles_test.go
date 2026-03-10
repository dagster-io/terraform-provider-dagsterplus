package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
)

func mockRawRole(id, name, scope string, perms []any) map[string]any {
	return map[string]any{
		"id":              id,
		"name":            name,
		"description":     "test role",
		"iconName":        "",
		"deploymentScope": scope,
		"permissions":     perms,
	}
}

func roleFixture() client.Role {
	return client.Role{
		Name:        "my-role",
		Description: "test role",
		Icon:        "",
		RoleType:    "deployment",
		Permissions: []string{"edit_alerts"},
	}
}

func TestCreateRole_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "CreateRole") {
			t.Errorf("expected CreateRole mutation, got: %s", b.Query)
		}
		if b.Variables["name"] != "my-role" {
			t.Errorf("name = %v, want my-role", b.Variables["name"])
		}
		if b.Variables["scope"] != "DEPLOYMENT" {
			t.Errorf("scope = %v, want DEPLOYMENT", b.Variables["scope"])
		}
		gqlOK(w, map[string]any{
			"createCustomRole": map[string]any{
				"__typename": "CreateOrUpdateCustomRoleSuccess",
				"customRole": mockRawRole("role-123", "my-role", "DEPLOYMENT", []any{"EDIT_ALERTS"}),
			},
		})
	}))
	defer srv.Close()

	role, err := newClient(srv).CreateRole(context.Background(), roleFixture())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if role.ID != "role-123" {
		t.Errorf("ID = %q, want role-123", role.ID)
	}
	if role.Name != "my-role" {
		t.Errorf("Name = %q, want my-role", role.Name)
	}
	if role.RoleType != "deployment" {
		t.Errorf("RoleType = %q, want deployment", role.RoleType)
	}
	if len(role.Permissions) != 1 || role.Permissions[0] != "edit_alerts" {
		t.Errorf("Permissions = %v, want [edit_alerts]", role.Permissions)
	}
}

func TestCreateRole_UnexpectedTypename(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"createCustomRole": map[string]any{
				"__typename": "CreateOrUpdateCustomRoleFailure",
				"customRole": nil,
			},
		})
	}))
	defer srv.Close()

	_, err := newClient(srv).CreateRole(context.Background(), roleFixture())
	if err == nil {
		t.Fatal("expected error for unexpected typename, got nil")
	}
}

func TestListRoles_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "ListRoles") {
			t.Errorf("expected ListRoles query, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"customRoles": []any{
				mockRawRole("role-1", "role-one", "DEPLOYMENT", []any{"EDIT_ALERTS"}),
				mockRawRole("role-2", "role-two", "ORGANIZATION", []any{"MANAGE_BILLING"}),
			},
		})
	}))
	defer srv.Close()

	roles, err := newClient(srv).ListRoles(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(roles) != 2 {
		t.Fatalf("len(roles) = %d, want 2", len(roles))
	}
	if roles[0].ID != "role-1" {
		t.Errorf("roles[0].ID = %q, want role-1", roles[0].ID)
	}
	if roles[1].RoleType != "organization" {
		t.Errorf("roles[1].RoleType = %q, want organization", roles[1].RoleType)
	}
}

func TestGetRole_Found(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"customRoles": []any{
				mockRawRole("role-123", "my-role", "DEPLOYMENT", []any{"EDIT_ALERTS"}),
			},
		})
	}))
	defer srv.Close()

	role, err := newClient(srv).GetRole(context.Background(), "role-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if role.Name != "my-role" {
		t.Errorf("Name = %q, want my-role", role.Name)
	}
}

func TestGetRole_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{"customRoles": []any{}})
	}))
	defer srv.Close()

	_, err := newClient(srv).GetRole(context.Background(), "role-missing")
	if err == nil {
		t.Fatal("expected error for not-found role, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention not found, got: %v", err)
	}
}

func TestGetRoleByName_Found(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"customRoles": []any{
				mockRawRole("role-123", "my-role", "DEPLOYMENT", []any{"EDIT_ALERTS"}),
			},
		})
	}))
	defer srv.Close()

	role, err := newClient(srv).GetRoleByName(context.Background(), "my-role")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if role.ID != "role-123" {
		t.Errorf("ID = %q, want role-123", role.ID)
	}
}

func TestGetRoleByName_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{"customRoles": []any{}})
	}))
	defer srv.Close()

	_, err := newClient(srv).GetRoleByName(context.Background(), "missing-role")
	if err == nil {
		t.Fatal("expected error for not-found role, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention not found, got: %v", err)
	}
}

func TestUpdateRole_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "UpdateRole") {
			t.Errorf("expected UpdateRole mutation, got: %s", b.Query)
		}
		if b.Variables["id"] != "role-123" {
			t.Errorf("id = %v, want role-123", b.Variables["id"])
		}
		if b.Variables["name"] != "my-role-renamed" {
			t.Errorf("name = %v, want my-role-renamed", b.Variables["name"])
		}
		gqlOK(w, map[string]any{
			"updateCustomRole": map[string]any{
				"__typename": "CreateOrUpdateCustomRoleSuccess",
				"customRole": mockRawRole("role-123", "my-role-renamed", "DEPLOYMENT", []any{"EDIT_ALERTS", "DELETE_RUNS"}),
			},
		})
	}))
	defer srv.Close()

	updated, err := newClient(srv).UpdateRole(context.Background(), client.Role{
		ID:          "role-123",
		Name:        "my-role-renamed",
		Description: "test role",
		RoleType:    "deployment",
		Permissions: []string{"edit_alerts", "delete_runs"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "my-role-renamed" {
		t.Errorf("Name = %q, want my-role-renamed", updated.Name)
	}
	if len(updated.Permissions) != 2 {
		t.Errorf("len(Permissions) = %d, want 2", len(updated.Permissions))
	}
}

func TestUpdateRole_UnexpectedTypename(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"updateCustomRole": map[string]any{
				"__typename": "CreateOrUpdateCustomRoleFailure",
				"customRole": nil,
			},
		})
	}))
	defer srv.Close()

	_, err := newClient(srv).UpdateRole(context.Background(), client.Role{
		ID: "role-123", Name: "x", RoleType: "deployment",
	})
	if err == nil {
		t.Fatal("expected error for unexpected typename, got nil")
	}
}

func TestDeleteRole_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "DeleteRole") {
			t.Errorf("expected DeleteRole mutation, got: %s", b.Query)
		}
		if b.Variables["id"] != "role-123" {
			t.Errorf("id = %v, want role-123", b.Variables["id"])
		}
		gqlOK(w, map[string]any{
			"deleteCustomRole": map[string]any{"__typename": "DeleteCustomRoleSuccess"},
		})
	}))
	defer srv.Close()

	if err := newClient(srv).DeleteRole(context.Background(), "role-123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteRole_GraphQLError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlErr(w, "role not found")
	}))
	defer srv.Close()

	err := newClient(srv).DeleteRole(context.Background(), "role-123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "role not found") {
		t.Errorf("error should include API message, got: %v", err)
	}
}
