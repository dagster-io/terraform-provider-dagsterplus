package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
)

// agentTokenPermissions builds a DagsterCloudScopedPermissionGrants payload.
// Absent grants are represented with an empty grant string.
func agentTokenPermissions(org, allBranch map[string]any, deployment []any) map[string]any {
	if org == nil {
		org = map[string]any{"grant": "", "deploymentScope": "ORGANIZATION", "deploymentId": 0}
	}
	if allBranch == nil {
		allBranch = map[string]any{"grant": "", "deploymentScope": "ALL_BRANCH_DEPLOYMENTS", "deploymentId": 0}
	}
	return map[string]any{
		"organizationPermissionGrant":         org,
		"allBranchDeploymentsPermissionGrant": allBranch,
		"deploymentPermissionGrants":          deployment,
	}
}

func TestSetAgentTokenGrant_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "SetAgentGrant") {
			t.Errorf("expected SetAgentGrant mutation, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"createOrUpdateAgentPermissions": map[string]any{
				"__typename": "DagsterCloudAgentToken",
				"id":         123,
				"permissions": agentTokenPermissions(nil, nil, []any{
					map[string]any{"grant": "AGENT", "deploymentScope": "DEPLOYMENT", "deploymentId": 42},
				}),
			},
		})
	}))
	defer srv.Close()

	g, err := newClient(srv).SetAgentTokenGrant(context.Background(), client.AgentTokenGrant{
		AgentTokenID:    "123",
		DeploymentScope: "deployment",
		DeploymentID:    42,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.DeploymentID != 42 {
		t.Errorf("DeploymentID = %d, want 42", g.DeploymentID)
	}
	if g.DeploymentScope != "deployment" {
		t.Errorf("DeploymentScope = %q, want deployment", g.DeploymentScope)
	}
}

func TestSetAgentTokenGrant_Organization(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"createOrUpdateAgentPermissions": map[string]any{
				"__typename": "DagsterCloudAgentToken",
				"id":         123,
				"permissions": agentTokenPermissions(
					map[string]any{"grant": "AGENT", "deploymentScope": "ORGANIZATION", "deploymentId": 0},
					nil, []any{}),
			},
		})
	}))
	defer srv.Close()

	g, err := newClient(srv).SetAgentTokenGrant(context.Background(), client.AgentTokenGrant{
		AgentTokenID:    "123",
		DeploymentScope: "organization",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.DeploymentScope != "organization" {
		t.Errorf("DeploymentScope = %q, want organization", g.DeploymentScope)
	}
}

func TestSetAgentTokenGrant_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"createOrUpdateAgentPermissions": map[string]any{
				"__typename": "UnauthorizedError",
				"message":    "unauthorized",
			},
		})
	}))
	defer srv.Close()

	_, err := newClient(srv).SetAgentTokenGrant(context.Background(), client.AgentTokenGrant{
		AgentTokenID: "123", DeploymentScope: "deployment", DeploymentID: 42,
	})
	if err == nil {
		t.Fatal("expected error for unauthorized, got nil")
	}
	if strings.Contains(err.Error(), "unexpected result type") {
		t.Errorf("UnauthorizedError should be handled explicitly, got: %v", err)
	}
}

func TestSetAgentTokenGrant_InvalidID(t *testing.T) {
	_, err := (&client.Client{}).SetAgentTokenGrant(context.Background(), client.AgentTokenGrant{
		AgentTokenID: "not-a-number", DeploymentScope: "deployment",
	})
	if err == nil {
		t.Fatal("expected error for non-numeric agent token id, got nil")
	}
	if !strings.Contains(err.Error(), "invalid agent token id") {
		t.Errorf("error should mention invalid id, got: %v", err)
	}
}

func TestGetAgentTokenGrant_Found(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "GetAgentTokenGrants") {
			t.Errorf("expected GetAgentTokenGrants query, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"agentTokenOrError": map[string]any{
				"__typename": "DagsterCloudAgentToken",
				"id":         99,
				"permissions": agentTokenPermissions(nil, nil, []any{
					map[string]any{"grant": "AGENT", "deploymentScope": "DEPLOYMENT", "deploymentId": 7},
				}),
			},
		})
	}))
	defer srv.Close()

	g, err := newClient(srv).GetAgentTokenGrant(context.Background(), "99", "deployment", 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.AgentTokenID != "99" {
		t.Errorf("AgentTokenID = %q, want 99", g.AgentTokenID)
	}
	if g.DeploymentID != 7 {
		t.Errorf("DeploymentID = %d, want 7", g.DeploymentID)
	}
}

func TestGetAgentTokenGrant_OrganizationAbsent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"agentTokenOrError": map[string]any{
				"__typename":  "DagsterCloudAgentToken",
				"id":          99,
				"permissions": agentTokenPermissions(nil, nil, []any{}),
			},
		})
	}))
	defer srv.Close()

	_, err := newClient(srv).GetAgentTokenGrant(context.Background(), "99", "organization", 0)
	if err == nil {
		t.Fatal("expected error for absent organization grant, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestGetAgentTokenGrant_TokenNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"agentTokenOrError": map[string]any{
				"__typename": "PythonError",
			},
		})
	}))
	defer srv.Close()

	_, err := newClient(srv).GetAgentTokenGrant(context.Background(), "404", "deployment", 7)
	if err == nil {
		t.Fatal("expected error for token not found, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestListAgentTokenDeploymentGrants_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"agentTokenOrError": map[string]any{
				"__typename": "DagsterCloudAgentToken",
				"id":         99,
				"permissions": agentTokenPermissions(nil, nil, []any{
					map[string]any{"grant": "AGENT", "deploymentScope": "DEPLOYMENT", "deploymentId": 1},
					map[string]any{"grant": "AGENT", "deploymentScope": "DEPLOYMENT", "deploymentId": 2},
				}),
			},
		})
	}))
	defer srv.Close()

	grants, err := newClient(srv).ListAgentTokenDeploymentGrants(context.Background(), "99")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(grants) != 2 {
		t.Fatalf("len(grants) = %d, want 2", len(grants))
	}
}

func TestDeleteAgentTokenGrant_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "DeleteAgentGrant") {
			t.Errorf("expected DeleteAgentGrant mutation, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"removeAgentPermissions": map[string]any{
				"__typename": "DagsterCloudAgentToken",
				"id":         123,
			},
		})
	}))
	defer srv.Close()

	if err := newClient(srv).DeleteAgentTokenGrant(context.Background(), "123", "organization", 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteAgentTokenGrant_GraphQLError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlErr(w, "unauthorized")
	}))
	defer srv.Close()

	err := newClient(srv).DeleteAgentTokenGrant(context.Background(), "123", "deployment", 42)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "DeleteAgentTokenGrant") {
		t.Errorf("error should mention DeleteAgentTokenGrant, got: %v", err)
	}
}
