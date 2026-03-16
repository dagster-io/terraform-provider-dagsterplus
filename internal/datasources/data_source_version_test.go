package datasources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func TestVersionDataSource_Schema(t *testing.T) {
	d := NewVersionDataSource().(*versionDataSource)
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	// deployment: Required (lookup key)
	deploymentRaw, ok := resp.Schema.Attributes["deployment"]
	if !ok {
		t.Fatal("missing 'deployment' attribute")
	}
	deploymentAttr, ok := deploymentRaw.(dsschema.StringAttribute)
	if !ok {
		t.Fatalf("deployment should be StringAttribute, got %T", deploymentRaw)
	}
	if !deploymentAttr.IsRequired() {
		t.Error("deployment should be Required")
	}
	if deploymentAttr.IsComputed() {
		t.Error("deployment should not be Computed")
	}

	// version: Computed only
	versionRaw, ok := resp.Schema.Attributes["version"]
	if !ok {
		t.Fatal("missing 'version' attribute")
	}
	versionAttr, ok := versionRaw.(dsschema.StringAttribute)
	if !ok {
		t.Fatalf("version should be StringAttribute, got %T", versionRaw)
	}
	if !versionAttr.IsComputed() {
		t.Error("version should be Computed")
	}
	if versionAttr.IsRequired() {
		t.Error("version should not be Required")
	}
}
