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
		"id":             "dep-123",
		"deploymentName": "prod",
		"deploymentType": "PROD",
		"createdAt":      "2024-01-15T10:00:00Z",
	}
}

func TestCreateDeployment_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "CreateDeployment") {
			t.Errorf("expected CreateDeployment mutation, got: %s", b.Query)
		}
		if b.Variables["deploymentName"] != "prod" {
			t.Errorf("deploymentName = %v, want prod", b.Variables["deploymentName"])
		}
		if b.Variables["deploymentType"] != "PROD" {
			t.Errorf("deploymentType = %v, want PROD", b.Variables["deploymentType"])
		}
		gqlOK(w, map[string]any{
			"createDeployment": map[string]any{
				"deployment": mockDeployment(),
			},
		})
	}))
	defer srv.Close()

	dep, err := newClient(srv).CreateDeployment(context.Background(), "prod", "PROD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dep.Name != "prod" {
		t.Errorf("Name = %q, want prod", dep.Name)
	}
	if dep.Type != "PROD" {
		t.Errorf("Type = %q, want PROD", dep.Type)
	}
	if dep.CreatedAt != "2024-01-15T10:00:00Z" {
		t.Errorf("CreatedAt = %q, unexpected", dep.CreatedAt)
	}
}

func TestCreateDeployment_GraphQLError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlErr(w, "deployment already exists")
	}))
	defer srv.Close()

	_, err := newClient(srv).CreateDeployment(context.Background(), "prod", "PROD")
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

	_, err := newClient(srv).CreateDeployment(context.Background(), "prod", "PROD")
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
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "DeleteDeployment") {
			t.Errorf("expected DeleteDeployment mutation, got: %s", b.Query)
		}
		if b.Variables["deploymentName"] != "prod" {
			t.Errorf("deploymentName = %v, want prod", b.Variables["deploymentName"])
		}
		gqlOK(w, map[string]any{
			"deleteDeployment": map[string]any{"success": true},
		})
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
