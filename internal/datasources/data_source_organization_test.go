package datasources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func TestOrganizationDataSource_Schema(t *testing.T) {
	d := NewOrganizationDataSource().(*organizationDataSource)
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	// All fields on organization data source are Computed (no Required lookup keys).
	computedFields := []string{
		"id", "public_id", "name", "status",
		"github_account_name", "github_app_id", "github_settings_url",
		"slack_team_id", "slack_team_name",
	}
	for _, fieldName := range computedFields {
		raw, ok := resp.Schema.Attributes[fieldName]
		if !ok {
			t.Errorf("missing attribute %q", fieldName)
			continue
		}
		strAttr, ok := raw.(dsschema.StringAttribute)
		if !ok {
			t.Errorf("%s should be StringAttribute, got %T", fieldName, raw)
			continue
		}
		if !strAttr.IsComputed() {
			t.Errorf("%s should be Computed", fieldName)
		}
		if strAttr.IsRequired() {
			t.Errorf("%s should not be Required", fieldName)
		}
	}

	// github_settings_url should be Sensitive
	urlRaw, ok := resp.Schema.Attributes["github_settings_url"]
	if !ok {
		t.Fatal("missing 'github_settings_url' attribute")
	}
	urlAttr, ok := urlRaw.(dsschema.StringAttribute)
	if !ok {
		t.Fatalf("github_settings_url should be StringAttribute, got %T", urlRaw)
	}
	if !urlAttr.IsSensitive() {
		t.Error("github_settings_url should be Sensitive")
	}

	// github_repos: Computed list
	reposRaw, ok := resp.Schema.Attributes["github_repos"]
	if !ok {
		t.Fatal("missing 'github_repos' attribute")
	}
	reposAttr, ok := reposRaw.(dsschema.ListAttribute)
	if !ok {
		t.Fatalf("github_repos should be ListAttribute, got %T", reposRaw)
	}
	if !reposAttr.IsComputed() {
		t.Error("github_repos should be Computed")
	}
}
