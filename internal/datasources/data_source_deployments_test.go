package datasources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func TestDeploymentsDataSource_Schema(t *testing.T) {
	d := NewDeploymentsDataSource().(*deploymentsDataSource)
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

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

	// deployments: Computed list nested attribute
	deploymentsRaw, ok := resp.Schema.Attributes["deployments"]
	if !ok {
		t.Fatal("missing 'deployments' attribute")
	}
	deploymentsAttr, ok := deploymentsRaw.(dsschema.ListNestedAttribute)
	if !ok {
		t.Fatalf("deployments should be ListNestedAttribute, got %T", deploymentsRaw)
	}
	if !deploymentsAttr.IsComputed() {
		t.Error("deployments should be Computed")
	}

	// All nested fields should be Computed
	for _, attrName := range []string{"id", "name", "type"} {
		a, ok := deploymentsAttr.NestedObject.Attributes[attrName]
		if !ok {
			t.Errorf("deployments nested object missing attribute %q", attrName)
			continue
		}
		strA, ok := a.(dsschema.StringAttribute)
		if !ok {
			t.Errorf("deployments.%s should be StringAttribute, got %T", attrName, a)
			continue
		}
		if !strA.IsComputed() {
			t.Errorf("deployments.%s should be Computed", attrName)
		}
		if strA.IsRequired() {
			t.Errorf("deployments.%s should not be Required", attrName)
		}
	}
}
