package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetOrganizationSettings_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "GetOrganizationSettings") {
			t.Errorf("expected GetOrganizationSettings query, got: %s", b.Query)
		}
		if r.URL.Path != "/graphql" {
			t.Errorf("path = %q, want /graphql", r.URL.Path)
		}
		gqlOK(w, map[string]any{
			"organizationSettings": map[string]any{
				"settings": map[string]any{"sso_default_role": "VIEWER"},
			},
		})
	}))
	defer srv.Close()

	settings, err := newClient(srv).GetOrganizationSettings(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(settings, "sso_default_role") {
		t.Errorf("settings should contain sso_default_role, got: %s", settings)
	}
}

func TestGetOrganizationSettings_NullSettings(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"organizationSettings": map[string]any{
				"settings": nil,
			},
		})
	}))
	defer srv.Close()

	settings, err := newClient(srv).GetOrganizationSettings(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if settings != "{}" {
		t.Errorf("settings = %q, want {}", settings)
	}
}

func TestSetOrganizationSettings_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "SetOrganizationSettings") {
			t.Errorf("expected SetOrganizationSettings mutation, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"setOrganizationSettings": map[string]any{
				"__typename": "OrganizationSettings",
				"settings":   map[string]any{},
			},
		})
	}))
	defer srv.Close()

	if err := newClient(srv).SetOrganizationSettings(context.Background(), `{"sso_default_role":"VIEWER"}`); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetOrganizationSettings_UnexpectedType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"setOrganizationSettings": map[string]any{
				"__typename": "UnauthorizedError",
			},
		})
	}))
	defer srv.Close()

	err := newClient(srv).SetOrganizationSettings(context.Background(), "{}")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
