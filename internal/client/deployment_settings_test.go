package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetDeploymentSettings_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "GetDeploymentSettings") {
			t.Errorf("expected GetDeploymentSettings query, got: %s", b.Query)
		}
		if r.URL.Path != "/prod/graphql" {
			t.Errorf("path = %q, want /prod/graphql", r.URL.Path)
		}
		gqlOK(w, map[string]any{
			"deploymentSettings": map[string]any{
				"settings": map[string]any{"run_queue": map[string]any{"max_concurrent_runs": 10}},
			},
		})
	}))
	defer srv.Close()

	settings, err := newClient(srv).GetDeploymentSettings(context.Background(), "prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(settings, "run_queue") {
		t.Errorf("settings should contain run_queue, got: %s", settings)
	}
}

func TestGetDeploymentSettings_NullSettings(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"deploymentSettings": map[string]any{
				"settings": nil,
			},
		})
	}))
	defer srv.Close()

	settings, err := newClient(srv).GetDeploymentSettings(context.Background(), "prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if settings != "{}" {
		t.Errorf("settings = %q, want {}", settings)
	}
}

func TestSetDeploymentSettings_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "SetDeploymentSettings") {
			t.Errorf("expected SetDeploymentSettings mutation, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"setDeploymentSettings": map[string]any{
				"__typename": "DeploymentSettings",
				"settings":   map[string]any{},
			},
		})
	}))
	defer srv.Close()

	if err := newClient(srv).SetDeploymentSettings(context.Background(), "prod", `{"run_queue":{"max_concurrent_runs":10}}`); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetDeploymentSettings_UnexpectedType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"setDeploymentSettings": map[string]any{
				"__typename": "UnauthorizedError",
			},
		})
	}))
	defer srv.Close()

	err := newClient(srv).SetDeploymentSettings(context.Background(), "prod", "{}")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSetDeploymentSettings_GraphQLError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlErr(w, "unauthorized")
	}))
	defer srv.Close()

	err := newClient(srv).SetDeploymentSettings(context.Background(), "prod", "{}")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
