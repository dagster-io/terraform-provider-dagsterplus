package datasources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func TestConfigurationDocumentDataSource_Schema(t *testing.T) {
	d := NewConfigurationDocumentDataSource().(*configurationDocumentDataSource)
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	// yaml_body: Required (input)
	yamlRaw, ok := resp.Schema.Attributes["yaml_body"]
	if !ok {
		t.Fatal("missing 'yaml_body' attribute")
	}
	yamlAttr, ok := yamlRaw.(dsschema.StringAttribute)
	if !ok {
		t.Fatalf("yaml_body should be StringAttribute, got %T", yamlRaw)
	}
	if !yamlAttr.IsRequired() {
		t.Error("yaml_body should be Required")
	}
	if yamlAttr.IsComputed() {
		t.Error("yaml_body should not be Computed")
	}

	// json: Computed only (derived from yaml_body)
	jsonRaw, ok := resp.Schema.Attributes["json"]
	if !ok {
		t.Fatal("missing 'json' attribute")
	}
	jsonAttr, ok := jsonRaw.(dsschema.StringAttribute)
	if !ok {
		t.Fatalf("json should be StringAttribute, got %T", jsonRaw)
	}
	if !jsonAttr.IsComputed() {
		t.Error("json should be Computed")
	}
	if jsonAttr.IsRequired() {
		t.Error("json should not be Required")
	}
}
