package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// --- mock helpers ---

func mockOrganization() map[string]any {
	return map[string]any{
		"id":       1,
		"publicId": "abc123",
		"name":     "test-org",
		"status":   "ACTIVE",
		"metadata": map[string]any{
			"slackAppInstallation":  nil,
			"githubAppInstallation": nil,
		},
	}
}

func mockOrganizationWithGitHub() map[string]any {
	org := mockOrganization()
	org["metadata"] = map[string]any{
		"slackAppInstallation": nil,
		"githubAppInstallation": map[string]any{
			"accountName": "my-github-org",
			"appId":       "app-123",
			"settingsUrl": "https://github.com/settings/installations/123",
			"repos":       []any{"repo-a", "repo-b"},
		},
	}
	return org
}

// --- GetOrganization ---

func TestGetOrganization_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "GetOrganization") {
			t.Errorf("expected GetOrganization query, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{"organization": mockOrganization()})
	}))
	defer srv.Close()

	org, err := newClient(srv).GetOrganization(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if org.Name != "test-org" {
		t.Errorf("Name = %q, want test-org", org.Name)
	}
	if org.Status != "ACTIVE" {
		t.Errorf("Status = %q, want ACTIVE", org.Status)
	}
	if org.GitHub != nil {
		t.Errorf("GitHub should be nil when not installed")
	}
	if org.Slack != nil {
		t.Errorf("Slack should be nil when not installed")
	}
}

func TestGetOrganization_WithGitHub(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{"organization": mockOrganizationWithGitHub()})
	}))
	defer srv.Close()

	org, err := newClient(srv).GetOrganization(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if org.GitHub == nil {
		t.Fatal("GitHub should be non-nil when installed")
	}
	if org.GitHub.AccountName != "my-github-org" {
		t.Errorf("AccountName = %q, want my-github-org", org.GitHub.AccountName)
	}
	if len(org.GitHub.Repos) != 2 {
		t.Errorf("Repos len = %d, want 2", len(org.GitHub.Repos))
	}
}

func TestGetOrganization_GraphQLError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlErr(w, "unauthorized")
	}))
	defer srv.Close()

	_, err := newClient(srv).GetOrganization(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "GetOrganization") {
		t.Errorf("error should include GetOrganization, got: %v", err)
	}
}

func TestGetOrganization_UsesOrgEndpoint(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gqlOK(w, map[string]any{"organization": mockOrganization()})
	}))
	defer srv.Close()

	if _, err := newClient(srv).GetOrganization(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/graphql" {
		t.Errorf("path = %q, want /graphql (org-level endpoint)", gotPath)
	}
}

// --- GetScimSyncEnabled ---

func TestGetScimSyncEnabled_True(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "GetScimSyncEnabled") {
			t.Errorf("expected GetScimSyncEnabled query, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{"scimSyncEnabled": true})
	}))
	defer srv.Close()

	enabled, err := newClient(srv).GetScimSyncEnabled(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !enabled {
		t.Errorf("expected enabled=true, got false")
	}
}

func TestGetScimSyncEnabled_False(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{"scimSyncEnabled": false})
	}))
	defer srv.Close()

	enabled, err := newClient(srv).GetScimSyncEnabled(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if enabled {
		t.Errorf("expected enabled=false, got true")
	}
}

func TestGetScimSyncEnabled_GraphQLError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlErr(w, "permission denied")
	}))
	defer srv.Close()

	_, err := newClient(srv).GetScimSyncEnabled(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "GetScimSyncEnabled") {
		t.Errorf("error should include GetScimSyncEnabled, got: %v", err)
	}
}

// --- SetScimSyncEnabled ---

func TestSetScimSyncEnabled_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "SetScimSyncEnabled") {
			t.Errorf("expected SetScimSyncEnabled mutation, got: %s", b.Query)
		}
		if b.Variables["enabled"] != true {
			t.Errorf("enabled = %v, want true", b.Variables["enabled"])
		}
		gqlOK(w, map[string]any{
			"setScimSyncEnabled": map[string]any{
				"__typename": "SetScimSyncEnabledSuccess",
				"enabled":    true,
			},
		})
	}))
	defer srv.Close()

	if err := newClient(srv).SetScimSyncEnabled(context.Background(), true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetScimSyncEnabled_UnexpectedTypename(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"setScimSyncEnabled": map[string]any{
				"__typename": "PythonError",
				"message":    "something went wrong",
			},
		})
	}))
	defer srv.Close()

	err := newClient(srv).SetScimSyncEnabled(context.Background(), true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "SetScimSyncEnabled") {
		t.Errorf("error should include SetScimSyncEnabled, got: %v", err)
	}
}

// --- GetAtlanIntegration ---

func TestGetAtlanIntegration_Set(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "GetAtlanIntegrationSettings") {
			t.Errorf("expected GetAtlanIntegrationSettings query, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"atlanIntegration": map[string]any{
				"atlanIntegrationSettingsOrError": map[string]any{
					"__typename": "AtlanIntegrationSettings",
					"token":      "my-token",
					"domain":     "my-org.atlan.com",
				},
			},
		})
	}))
	defer srv.Close()

	atlan, err := newClient(srv).GetAtlanIntegration(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atlan == nil {
		t.Fatal("expected non-nil AtlanIntegration")
	}
	if atlan.Token != "my-token" {
		t.Errorf("Token = %q, want my-token", atlan.Token)
	}
	if atlan.Domain != "my-org.atlan.com" {
		t.Errorf("Domain = %q, want my-org.atlan.com", atlan.Domain)
	}
}

func TestGetAtlanIntegration_Unset(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"atlanIntegration": map[string]any{
				"atlanIntegrationSettingsOrError": map[string]any{
					"__typename": "AtlanIntegrationSettingsUnset",
					"message":    "not configured",
				},
			},
		})
	}))
	defer srv.Close()

	atlan, err := newClient(srv).GetAtlanIntegration(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atlan != nil {
		t.Errorf("expected nil when unset, got %+v", atlan)
	}
}

func TestGetAtlanIntegration_GraphQLError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlErr(w, "unauthorized")
	}))
	defer srv.Close()

	_, err := newClient(srv).GetAtlanIntegration(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "GetAtlanIntegration") {
		t.Errorf("error should include GetAtlanIntegration, got: %v", err)
	}
}

// --- SetAtlanIntegration ---

func TestSetAtlanIntegration_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "SetAtlanIntegrationSettings") {
			t.Errorf("expected SetAtlanIntegrationSettings mutation, got: %s", b.Query)
		}
		if b.Variables["token"] != "my-token" {
			t.Errorf("token = %v, want my-token", b.Variables["token"])
		}
		if b.Variables["domain"] != "my-org.atlan.com" {
			t.Errorf("domain = %v, want my-org.atlan.com", b.Variables["domain"])
		}
		gqlOK(w, map[string]any{
			"setAtlanIntegrationSettings": map[string]any{
				"__typename":   "SetAtlanIntegrationSettingsSuccess",
				"organization": "test-org",
				"success":      true,
			},
		})
	}))
	defer srv.Close()

	if err := newClient(srv).SetAtlanIntegration(context.Background(), "my-token", "my-org.atlan.com"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetAtlanIntegration_UnexpectedTypename(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"setAtlanIntegrationSettings": map[string]any{
				"__typename": "UnauthorizedError",
				"message":    "not allowed",
			},
		})
	}))
	defer srv.Close()

	err := newClient(srv).SetAtlanIntegration(context.Background(), "t", "d")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "SetAtlanIntegration") {
		t.Errorf("error should include SetAtlanIntegration, got: %v", err)
	}
}

// --- DeleteAtlanIntegration ---

func TestDeleteAtlanIntegration_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "DeleteAtlanIntegrationSettings") {
			t.Errorf("expected DeleteAtlanIntegrationSettings mutation, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"deleteAtlanIntegrationSettings": map[string]any{
				"__typename":   "DeleteAtlanIntegrationSuccess",
				"organization": "test-org",
				"success":      true,
			},
		})
	}))
	defer srv.Close()

	if err := newClient(srv).DeleteAtlanIntegration(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteAtlanIntegration_UnexpectedTypename(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"deleteAtlanIntegrationSettings": map[string]any{
				"__typename": "PythonError",
				"message":    "error",
			},
		})
	}))
	defer srv.Close()

	err := newClient(srv).DeleteAtlanIntegration(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "DeleteAtlanIntegration") {
		t.Errorf("error should include DeleteAtlanIntegration, got: %v", err)
	}
}

// --- SelectGithubInstallation ---

func TestSelectGithubInstallation_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "SelectGithubInstallation") {
			t.Errorf("expected SelectGithubInstallation mutation, got: %s", b.Query)
		}
		if b.Variables["accountName"] != "my-github-org" {
			t.Errorf("accountName = %v, want my-github-org", b.Variables["accountName"])
		}
		gqlOK(w, map[string]any{
			"selectInstallation": map[string]any{
				"__typename":  "GithubAppInstallation",
				"accountName": "my-github-org",
				"appId":       "app-123",
				"settingsUrl": "https://github.com/settings/installations/123",
				"repos":       []any{"repo-a"},
			},
		})
	}))
	defer srv.Close()

	gh, err := newClient(srv).SelectGithubInstallation(context.Background(), "my-github-org")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gh.AccountName != "my-github-org" {
		t.Errorf("AccountName = %q, want my-github-org", gh.AccountName)
	}
	if gh.AppID != "app-123" {
		t.Errorf("AppID = %q, want app-123", gh.AppID)
	}
}

func TestSelectGithubInstallation_UnexpectedTypename(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"selectInstallation": map[string]any{
				"__typename": "InstallationNotFoundError",
				"message":    "not found",
			},
		})
	}))
	defer srv.Close()

	_, err := newClient(srv).SelectGithubInstallation(context.Background(), "missing-org")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "SelectGithubInstallation") {
		t.Errorf("error should include SelectGithubInstallation, got: %v", err)
	}
}

// --- DeselectGithubInstallation ---

func TestDeselectGithubInstallation_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "DeselectGithubInstallation") {
			t.Errorf("expected DeselectGithubInstallation mutation, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"deselectInstallation": map[string]any{
				"__typename": "DeselectInstallationSuccess",
				"ok":         true,
			},
		})
	}))
	defer srv.Close()

	if err := newClient(srv).DeselectGithubInstallation(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeselectGithubInstallation_UnexpectedTypename(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"deselectInstallation": map[string]any{
				"__typename": "GitHubError",
				"message":    "error",
			},
		})
	}))
	defer srv.Close()

	err := newClient(srv).DeselectGithubInstallation(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "DeselectGithubInstallation") {
		t.Errorf("error should include DeselectGithubInstallation, got: %v", err)
	}
}
