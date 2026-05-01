package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
)

func TestSetServiceUserGrant_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "SetServiceUserGrant") {
			t.Errorf("expected SetServiceUserGrant mutation, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"createOrUpdateServiceUserPermissions": map[string]any{
				"__typename": "ServiceUserWithScopedPermissionGrants",
				"id":         "su-id-1",
				"organizationPermissionGrant": map[string]any{
					"grant": "", "customRoleId": "", "deploymentScope": "ORGANIZATION", "deploymentId": 0,
				},
				"allBranchDeploymentsPermissionGrant": map[string]any{
					"grant": "", "customRoleId": "", "deploymentScope": "ALL_BRANCH_DEPLOYMENTS", "deploymentId": 0,
				},
				"deploymentPermissionGrants": []any{
					map[string]any{
						"grant": "VIEWER", "customRoleId": "", "deploymentScope": "DEPLOYMENT",
						"deploymentId": 42, "locationGrants": []any{},
					},
				},
			},
		})
	}))
	defer srv.Close()

	g, err := newClient(srv).SetServiceUserGrant(context.Background(), client.ServiceUserGrant{
		ServiceUserID:   "su-id-1",
		DeploymentScope: "deployment",
		DeploymentID:    42,
		Grant:           "VIEWER",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Grant != "VIEWER" {
		t.Errorf("Grant = %q, want VIEWER", g.Grant)
	}
	if g.DeploymentID != 42 {
		t.Errorf("DeploymentID = %d, want 42", g.DeploymentID)
	}
}

func TestSetServiceUserGrant_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"createOrUpdateServiceUserPermissions": map[string]any{
				"__typename": "ServiceUserNotFoundError",
				"message":    "service user not found",
			},
		})
	}))
	defer srv.Close()

	_, err := newClient(srv).SetServiceUserGrant(context.Background(), client.ServiceUserGrant{
		ServiceUserID:   "missing-id",
		DeploymentScope: "deployment",
		DeploymentID:    42,
		Grant:           "VIEWER",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "SetServiceUserGrant") {
		t.Errorf("error should mention SetServiceUserGrant, got: %v", err)
	}
}

func TestGetServiceUserGrant_Found(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "GetServiceUserGrants") {
			t.Errorf("expected GetServiceUserGrants query, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"serviceUser": map[string]any{
				"__typename":                          "ServiceUserWithScopedPermissionGrants",
				"id":                                  "su-id-1",
				"organizationPermissionGrant":         map[string]any{"grant": "", "customRoleId": "", "deploymentScope": "ORGANIZATION", "deploymentId": 0},
				"allBranchDeploymentsPermissionGrant": map[string]any{"grant": "", "customRoleId": "", "deploymentScope": "ALL_BRANCH_DEPLOYMENTS", "deploymentId": 0},
				"deploymentPermissionGrants": []any{
					map[string]any{
						"grant": "EDITOR", "customRoleId": "", "deploymentScope": "DEPLOYMENT",
						"deploymentId": 99, "locationGrants": []any{},
					},
				},
			},
		})
	}))
	defer srv.Close()

	g, err := newClient(srv).GetServiceUserGrant(context.Background(), "su-id-1", "deployment", 99)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Grant != "EDITOR" {
		t.Errorf("Grant = %q, want EDITOR", g.Grant)
	}
	if g.ServiceUserID != "su-id-1" {
		t.Errorf("ServiceUserID = %q, want su-id-1", g.ServiceUserID)
	}
}

func TestGetServiceUserGrant_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"serviceUser": map[string]any{
				"__typename":                          "ServiceUserWithScopedPermissionGrants",
				"id":                                  "su-id-1",
				"organizationPermissionGrant":         map[string]any{"grant": "", "customRoleId": "", "deploymentScope": "ORGANIZATION", "deploymentId": 0},
				"allBranchDeploymentsPermissionGrant": map[string]any{"grant": "", "customRoleId": "", "deploymentScope": "ALL_BRANCH_DEPLOYMENTS", "deploymentId": 0},
				"deploymentPermissionGrants":          []any{},
			},
		})
	}))
	defer srv.Close()

	_, err := newClient(srv).GetServiceUserGrant(context.Background(), "su-id-1", "deployment", 99)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestGetServiceUserGrant_ServiceUserNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"serviceUser": map[string]any{
				"__typename": "ServiceUserNotFoundError",
				"message":    "service user not found",
			},
		})
	}))
	defer srv.Close()

	_, err := newClient(srv).GetServiceUserGrant(context.Background(), "missing-id", "deployment", 99)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestDeleteServiceUserGrant_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "DeleteServiceUserGrant") {
			t.Errorf("expected DeleteServiceUserGrant mutation, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"removeServiceUserPermissions": map[string]any{
				"__typename": "ServiceUserWithScopedPermissionGrants",
				"id":         "su-id-1",
			},
		})
	}))
	defer srv.Close()

	err := newClient(srv).DeleteServiceUserGrant(context.Background(), "su-id-1", "deployment", 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteServiceUserGrant_GraphQLError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlErr(w, "unauthorized")
	}))
	defer srv.Close()

	err := newClient(srv).DeleteServiceUserGrant(context.Background(), "su-id-1", "deployment", 42)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "DeleteServiceUserGrant") {
		t.Errorf("error should mention DeleteServiceUserGrant, got: %v", err)
	}
}
