package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
)

func mockTeamPermsOrg(teamID, grant string) map[string]any {
	return map[string]any{
		"team": map[string]any{"id": teamID},
		"organizationPermissionGrant": map[string]any{
			"grant":           grant,
			"customRoleId":    "",
			"deploymentScope": "ORGANIZATION",
			"deploymentId":    nil,
		},
		"allBranchDeploymentsPermissionGrant": nil,
		"deploymentPermissionGrants":          []any{},
	}
}

func mockTeamPermsDeployment(teamID string, deploymentID int, grant string) map[string]any {
	return map[string]any{
		"team":                                map[string]any{"id": teamID},
		"organizationPermissionGrant":         nil,
		"allBranchDeploymentsPermissionGrant": nil,
		"deploymentPermissionGrants": []any{
			map[string]any{
				"grant":           grant,
				"customRoleId":    "",
				"deploymentScope": "DEPLOYMENT",
				"deploymentId":    deploymentID,
			},
		},
	}
}

func TestSetTeamGrant_OrgScope_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "SetTeamGrant") {
			t.Errorf("expected SetTeamGrant mutation, got: %s", b.Query)
		}
		if b.Variables["teamId"] != "team-abc" {
			t.Errorf("teamId = %v, want team-abc", b.Variables["teamId"])
		}
		if b.Variables["scope"] != "ORGANIZATION" {
			t.Errorf("scope = %v, want ORGANIZATION", b.Variables["scope"])
		}
		if b.Variables["grant"] != "VIEWER" {
			t.Errorf("grant = %v, want VIEWER", b.Variables["grant"])
		}
		gqlOK(w, map[string]any{
			"createOrUpdateTeamPermission": map[string]any{
				"__typename":      "CreateOrUpdateTeamPermissionSuccess",
				"teamPermissions": mockTeamPermsOrg("team-abc", "VIEWER"),
			},
		})
	}))
	defer srv.Close()

	g, err := newClient(srv).SetTeamGrant(context.Background(), client.TeamGrant{
		TeamID:          "team-abc",
		DeploymentScope: "organization",
		Grant:           "VIEWER",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Grant != "VIEWER" {
		t.Errorf("Grant = %q, want VIEWER", g.Grant)
	}
	if g.DeploymentScope != "organization" {
		t.Errorf("DeploymentScope = %q, want organization", g.DeploymentScope)
	}
}

func TestSetTeamGrant_DeploymentScope_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if b.Variables["scope"] != "DEPLOYMENT" {
			t.Errorf("scope = %v, want DEPLOYMENT", b.Variables["scope"])
		}
		if b.Variables["deploymentId"] != float64(42) {
			t.Errorf("deploymentId = %v, want 42", b.Variables["deploymentId"])
		}
		gqlOK(w, map[string]any{
			"createOrUpdateTeamPermission": map[string]any{
				"__typename":      "CreateOrUpdateTeamPermissionSuccess",
				"teamPermissions": mockTeamPermsDeployment("team-abc", 42, "EDITOR"),
			},
		})
	}))
	defer srv.Close()

	g, err := newClient(srv).SetTeamGrant(context.Background(), client.TeamGrant{
		TeamID:          "team-abc",
		DeploymentScope: "deployment",
		DeploymentID:    42,
		Grant:           "EDITOR",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Grant != "EDITOR" {
		t.Errorf("Grant = %q, want EDITOR", g.Grant)
	}
	if g.DeploymentID != 42 {
		t.Errorf("DeploymentID = %d, want 42", g.DeploymentID)
	}
}

func TestSetTeamGrant_UnexpectedTypename(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"createOrUpdateTeamPermission": map[string]any{
				"__typename":      "CreateOrUpdateTeamPermissionFailure",
				"teamPermissions": nil,
			},
		})
	}))
	defer srv.Close()

	_, err := newClient(srv).SetTeamGrant(context.Background(), client.TeamGrant{
		TeamID: "team-abc", DeploymentScope: "organization", Grant: "VIEWER",
	})
	if err == nil {
		t.Fatal("expected error for unexpected typename, got nil")
	}
}

func TestGetTeamGrant_Found(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "ListTeamGrants") {
			t.Errorf("expected ListTeamGrants query, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"teamPermissions": []any{
				mockTeamPermsOrg("team-abc", "ADMIN"),
			},
		})
	}))
	defer srv.Close()

	g, err := newClient(srv).GetTeamGrant(context.Background(), "team-abc", "organization", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Grant != "ADMIN" {
		t.Errorf("Grant = %q, want ADMIN", g.Grant)
	}
}

func TestGetTeamGrant_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{"teamPermissions": []any{}})
	}))
	defer srv.Close()

	_, err := newClient(srv).GetTeamGrant(context.Background(), "team-missing", "organization", 0)
	if err == nil {
		t.Fatal("expected error for not-found grant, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention not found, got: %v", err)
	}
}

func TestDeleteTeamGrant_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "DeleteTeamGrant") {
			t.Errorf("expected DeleteTeamGrant mutation, got: %s", b.Query)
		}
		if b.Variables["teamId"] != "team-abc" {
			t.Errorf("teamId = %v, want team-abc", b.Variables["teamId"])
		}
		if b.Variables["scope"] != "ORGANIZATION" {
			t.Errorf("scope = %v, want ORGANIZATION", b.Variables["scope"])
		}
		gqlOK(w, map[string]any{
			"removeTeamPermission": map[string]any{
				"__typename": "RemoveTeamPermissionSuccess",
			},
		})
	}))
	defer srv.Close()

	if err := newClient(srv).DeleteTeamGrant(context.Background(), "team-abc", "organization", 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteTeamGrant_UnexpectedTypename(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"removeTeamPermission": map[string]any{
				"__typename": "DeleteTeamPermissionFailure",
			},
		})
	}))
	defer srv.Close()

	err := newClient(srv).DeleteTeamGrant(context.Background(), "team-abc", "organization", 0)
	if err == nil {
		t.Fatal("expected error for unexpected typename, got nil")
	}
}

func TestDeleteTeamGrant_GraphQLError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlErr(w, "team not found")
	}))
	defer srv.Close()

	err := newClient(srv).DeleteTeamGrant(context.Background(), "team-abc", "organization", 0)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "team not found") {
		t.Errorf("error should include API message, got: %v", err)
	}
}
