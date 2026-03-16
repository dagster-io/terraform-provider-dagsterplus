package resources

import (
	"context"
	"testing"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestGithubIntegrationResource_Schema(t *testing.T) {
	r := NewGithubIntegrationResource().(*githubIntegrationResource)
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	// id: Computed only
	idRaw, ok := resp.Schema.Attributes["id"]
	if !ok {
		t.Fatal("missing 'id' attribute")
	}
	idAttr, ok := idRaw.(rsschema.StringAttribute)
	if !ok {
		t.Fatalf("id should be StringAttribute, got %T", idRaw)
	}
	if !idAttr.IsComputed() {
		t.Error("id should be Computed")
	}

	// account_name: Required
	acctRaw, ok := resp.Schema.Attributes["account_name"]
	if !ok {
		t.Fatal("missing 'account_name' attribute")
	}
	acctAttr, ok := acctRaw.(rsschema.StringAttribute)
	if !ok {
		t.Fatalf("account_name should be StringAttribute, got %T", acctRaw)
	}
	if !acctAttr.IsRequired() {
		t.Error("account_name should be Required")
	}

	// app_id: Computed only
	appIDRaw, ok := resp.Schema.Attributes["app_id"]
	if !ok {
		t.Fatal("missing 'app_id' attribute")
	}
	appIDAttr, ok := appIDRaw.(rsschema.StringAttribute)
	if !ok {
		t.Fatalf("app_id should be StringAttribute, got %T", appIDRaw)
	}
	if !appIDAttr.IsComputed() {
		t.Error("app_id should be Computed")
	}
	if appIDAttr.IsRequired() {
		t.Error("app_id should not be Required")
	}

	// settings_url: Computed, Sensitive
	urlRaw, ok := resp.Schema.Attributes["settings_url"]
	if !ok {
		t.Fatal("missing 'settings_url' attribute")
	}
	urlAttr, ok := urlRaw.(rsschema.StringAttribute)
	if !ok {
		t.Fatalf("settings_url should be StringAttribute, got %T", urlRaw)
	}
	if !urlAttr.IsComputed() {
		t.Error("settings_url should be Computed")
	}
	if !urlAttr.IsSensitive() {
		t.Error("settings_url should be Sensitive")
	}

	// repos: Computed list
	reposRaw, ok := resp.Schema.Attributes["repos"]
	if !ok {
		t.Fatal("missing 'repos' attribute")
	}
	reposAttr, ok := reposRaw.(rsschema.ListAttribute)
	if !ok {
		t.Fatalf("repos should be ListAttribute, got %T", reposRaw)
	}
	if !reposAttr.IsComputed() {
		t.Error("repos should be Computed")
	}
}

func TestGithubInstallationToModel_Basic(t *testing.T) {
	gh := &client.GitHubInstallation{
		AccountName: "my-github-org",
		AppID:       "app-123",
		SettingsURL: "https://github.com/organizations/my-github-org/settings/installations/123",
		Repos:       []string{"repo-a", "repo-b"},
	}

	m, diags := githubInstallationToModel("my-dagster-org", gh)
	if diags.HasError() {
		t.Fatalf("githubInstallationToModel returned diagnostics: %v", diags)
	}

	if m.ID.ValueString() != "my-dagster-org" {
		t.Errorf("expected ID my-dagster-org, got %q", m.ID.ValueString())
	}
	if m.AccountName.ValueString() != "my-github-org" {
		t.Errorf("expected AccountName my-github-org, got %q", m.AccountName.ValueString())
	}
	if m.AppID.ValueString() != "app-123" {
		t.Errorf("expected AppID app-123, got %q", m.AppID.ValueString())
	}
	if m.SettingsURL.ValueString() != "https://github.com/organizations/my-github-org/settings/installations/123" {
		t.Errorf("unexpected SettingsURL: %q", m.SettingsURL.ValueString())
	}
	if m.Repos.IsNull() || m.Repos.IsUnknown() {
		t.Fatal("Repos should not be null/unknown")
	}
	var repos []string
	m.Repos.ElementsAs(context.Background(), &repos, false)
	if len(repos) != 2 || repos[0] != "repo-a" || repos[1] != "repo-b" {
		t.Errorf("unexpected Repos: %v", repos)
	}
}

func TestGithubInstallationToModel_EmptyRepos(t *testing.T) {
	gh := &client.GitHubInstallation{
		AccountName: "acme",
		AppID:       "42",
		SettingsURL: "https://example.com",
		Repos:       []string{},
	}

	m, diags := githubInstallationToModel("acme-org", gh)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if m.Repos.IsNull() {
		t.Error("Repos should not be null for empty slice")
	}
}
