package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func mockDeployment() map[string]any {
	return map[string]any{
		"deploymentId":   42,
		"deploymentName": "prod",
		"deploymentType": "PROD",
		"agentType":      "HYBRID",
	}
}

func TestCreateDeployment_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "CreateDeployment") {
			t.Errorf("expected CreateDeployment mutation, got: %s", b.Query)
		}
		if b.Variables["name"] != "prod" {
			t.Errorf("name = %v, want prod", b.Variables["name"])
		}
		if b.Variables["agentType"] != "HYBRID" {
			t.Errorf("agentType = %v, want HYBRID", b.Variables["agentType"])
		}
		gqlOK(w, map[string]any{
			"createDeployment": map[string]any{
				"__typename":     "DagsterCloudDeployment",
				"deploymentId":   99,
				"deploymentName": "prod",
				"deploymentType": "PROD",
				"agentType":      "HYBRID",
			},
		})
	}))
	defer srv.Close()

	dep, err := newClient(srv).CreateDeployment(context.Background(), "prod", "HYBRID")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dep.Name != "prod" {
		t.Errorf("Name = %q, want prod", dep.Name)
	}
	if dep.Type != "PROD" {
		t.Errorf("Type = %q, want PROD", dep.Type)
	}
	if dep.IntID != 99 {
		t.Errorf("IntID = %d, want 99", dep.IntID)
	}
	if dep.AgentType != "HYBRID" {
		t.Errorf("AgentType = %q, want HYBRID", dep.AgentType)
	}
}

// An empty agent type must be sent as null so the organization default applies,
// rather than being coerced into an invalid empty enum value.
func TestCreateDeployment_OmittedAgentType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if got, ok := b.Variables["agentType"]; !ok || got != nil {
			t.Errorf("agentType = %#v, want null", got)
		}
		gqlOK(w, map[string]any{
			"createDeployment": map[string]any{
				"__typename":     "DagsterCloudDeployment",
				"deploymentId":   99,
				"deploymentName": "prod",
				"deploymentType": "PROD",
				"agentType":      "SERVERLESS",
			},
		})
	}))
	defer srv.Close()

	dep, err := newClient(srv).CreateDeployment(context.Background(), "prod", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dep.AgentType != "SERVERLESS" {
		t.Errorf("AgentType = %q, want SERVERLESS", dep.AgentType)
	}
}

func TestCreateDeployment_DuplicateDeploymentError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"createDeployment": map[string]any{
				"__typename": "DuplicateDeploymentError",
				"message":    "deployment prod already exists",
			},
		})
	}))
	defer srv.Close()

	_, err := newClient(srv).CreateDeployment(context.Background(), "prod", "HYBRID")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "deployment prod already exists") {
		t.Errorf("error should include API message, got: %v", err)
	}
}

func TestCreateDeployment_GraphQLError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlErr(w, "deployment already exists")
	}))
	defer srv.Close()

	_, err := newClient(srv).CreateDeployment(context.Background(), "prod", "SERVERLESS")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "deployment already exists") {
		t.Errorf("error should include API message, got: %v", err)
	}
}

func TestCreateDeployment_NilDeployment(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"createDeployment": map[string]any{"deployment": nil},
		})
	}))
	defer srv.Close()

	_, err := newClient(srv).CreateDeployment(context.Background(), "prod", "SERVERLESS")
	if err == nil {
		t.Fatal("expected error for nil deployment, got nil")
	}
}

func TestListDeployments_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "ListDeployments") {
			t.Errorf("expected ListDeployments query, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"deployments": []any{mockDeployment()},
		})
	}))
	defer srv.Close()

	deps, err := newClient(srv).ListDeployments(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(deps) != 1 {
		t.Fatalf("len = %d, want 1", len(deps))
	}
	if deps[0].Name != "prod" {
		t.Errorf("Name = %q, want prod", deps[0].Name)
	}
	if deps[0].IntID != 42 {
		t.Errorf("IntID = %d, want 42", deps[0].IntID)
	}
	if deps[0].AgentType != "HYBRID" {
		t.Errorf("AgentType = %q, want HYBRID", deps[0].AgentType)
	}
}

func TestListDeployments_Empty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{"deployments": []any{}})
	}))
	defer srv.Close()

	deps, err := newClient(srv).ListDeployments(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(deps) != 0 {
		t.Errorf("expected empty slice, got %d items", len(deps))
	}
}

func TestGetDeployment_Found(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"deployments": []any{
				mockDeployment(),
				map[string]any{
					"id": "dep-456", "deploymentName": "staging",
					"deploymentType": "BRANCH", "createdAt": "2024-01-01T00:00:00Z",
				},
			},
		})
	}))
	defer srv.Close()

	dep, err := newClient(srv).GetDeployment(context.Background(), "prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dep.Name != "prod" {
		t.Errorf("Name = %q, want prod", dep.Name)
	}
}

func TestGetDeployment_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{"deployments": []any{}})
	}))
	defer srv.Close()

	_, err := newClient(srv).GetDeployment(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing deployment, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention not found, got: %v", err)
	}
}

func TestDeleteDeployment_Success(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		calls++
		if calls == 1 {
			// First: ListDeployments to resolve the integer ID.
			if !strings.Contains(b.Query, "ListDeployments") {
				t.Errorf("expected ListDeployments query, got: %s", b.Query)
			}
			gqlOK(w, map[string]any{
				"deployments": []any{
					map[string]any{"deploymentId": 42, "deploymentName": "prod", "deploymentType": "PROD"},
				},
			})
		} else {
			// Second: DeleteDeployment by integer ID.
			if !strings.Contains(b.Query, "DeleteDeployment") {
				t.Errorf("expected DeleteDeployment mutation, got: %s", b.Query)
			}
			if b.Variables["id"] != float64(42) {
				t.Errorf("id = %v, want 42", b.Variables["id"])
			}
			gqlOK(w, map[string]any{
				"deleteDeployment": map[string]any{"__typename": "DagsterCloudDeployment"},
			})
		}
	}))
	defer srv.Close()

	if err := newClient(srv).DeleteDeployment(context.Background(), "prod"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteDeployment_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlErr(w, "deployment not found")
	}))
	defer srv.Close()

	err := newClient(srv).DeleteDeployment(context.Background(), "prod")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUpdateDeploymentAgentType_Success(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		calls++
		if calls == 1 {
			// First: ListDeployments to resolve the integer ID.
			if !strings.Contains(b.Query, "ListDeployments") {
				t.Errorf("expected ListDeployments query, got: %s", b.Query)
			}
			gqlOK(w, map[string]any{"deployments": []any{mockDeployment()}})
			return
		}
		if !strings.Contains(b.Query, "UpdateDeploymentAgentType") {
			t.Errorf("expected UpdateDeploymentAgentType mutation, got: %s", b.Query)
		}
		if b.Variables["id"] != float64(42) {
			t.Errorf("id = %v, want 42", b.Variables["id"])
		}
		if b.Variables["agentType"] != "SERVERLESS" {
			t.Errorf("agentType = %v, want SERVERLESS", b.Variables["agentType"])
		}
		gqlOK(w, map[string]any{
			"updateDeploymentAgentType": map[string]any{
				"__typename":     "DagsterCloudDeployment",
				"deploymentId":   42,
				"deploymentName": "prod",
				"deploymentType": "PROD",
				"agentType":      "SERVERLESS",
			},
		})
	}))
	defer srv.Close()

	dep, err := newClient(srv).UpdateDeploymentAgentType(context.Background(), "prod", "SERVERLESS")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dep.AgentType != "SERVERLESS" {
		t.Errorf("AgentType = %q, want SERVERLESS", dep.AgentType)
	}
	if dep.IntID != 42 {
		t.Errorf("IntID = %d, want 42", dep.IntID)
	}
}

func TestUpdateDeploymentAgentType_NotFoundError(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			gqlOK(w, map[string]any{"deployments": []any{mockDeployment()}})
			return
		}
		gqlOK(w, map[string]any{
			"updateDeploymentAgentType": map[string]any{
				"__typename": "DeploymentNotFoundError",
				"message":    "deployment 42 not found",
			},
		})
	}))
	defer srv.Close()

	_, err := newClient(srv).UpdateDeploymentAgentType(context.Background(), "prod", "SERVERLESS")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "deployment 42 not found") {
		t.Errorf("error should include API message, got: %v", err)
	}
}

func TestUpdateDeploymentAgentType_UnknownDeployment(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{"deployments": []any{}})
	}))
	defer srv.Close()

	_, err := newClient(srv).UpdateDeploymentAgentType(context.Background(), "missing", "HYBRID")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention not found, got: %v", err)
	}
}
