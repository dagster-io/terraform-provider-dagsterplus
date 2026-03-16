package datasources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func TestCodeLocationsDataSource_Schema(t *testing.T) {
	d := NewCodeLocationsDataSource().(*codeLocationsDataSource)
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

	// id: Computed only
	idRaw, ok := resp.Schema.Attributes["id"]
	if !ok {
		t.Fatal("missing 'id' attribute")
	}
	idAttr, ok := idRaw.(dsschema.StringAttribute)
	if !ok {
		t.Fatalf("id should be StringAttribute, got %T", idRaw)
	}
	if !idAttr.IsComputed() {
		t.Error("id should be Computed")
	}

	// code_locations: Computed list nested attribute
	clRaw, ok := resp.Schema.Attributes["code_locations"]
	if !ok {
		t.Fatal("missing 'code_locations' attribute")
	}
	clAttr, ok := clRaw.(dsschema.ListNestedAttribute)
	if !ok {
		t.Fatalf("code_locations should be ListNestedAttribute, got %T", clRaw)
	}
	if !clAttr.IsComputed() {
		t.Error("code_locations should be Computed")
	}

	// All fields inside code_locations items should be Computed
	for _, attrName := range []string{"id", "deployment", "name", "image", "working_directory", "executable_path"} {
		a, ok := clAttr.NestedObject.Attributes[attrName]
		if !ok {
			t.Errorf("code_locations nested object missing attribute %q", attrName)
			continue
		}
		strA, ok := a.(dsschema.StringAttribute)
		if !ok {
			t.Errorf("code_locations.%s should be StringAttribute, got %T", attrName, a)
			continue
		}
		if !strA.IsComputed() {
			t.Errorf("code_locations.%s should be Computed", attrName)
		}
	}
}
