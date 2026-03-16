package datasources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func TestTeamsDataSource_Schema(t *testing.T) {
	d := NewTeamsDataSource().(*teamsDataSource)
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

	// teams: Computed list nested attribute
	teamsRaw, ok := resp.Schema.Attributes["teams"]
	if !ok {
		t.Fatal("missing 'teams' attribute")
	}
	teamsAttr, ok := teamsRaw.(dsschema.ListNestedAttribute)
	if !ok {
		t.Fatalf("teams should be ListNestedAttribute, got %T", teamsRaw)
	}
	if !teamsAttr.IsComputed() {
		t.Error("teams should be Computed")
	}

	// Nested team fields: id and name should be Computed
	for _, attrName := range []string{"id", "name"} {
		a, ok := teamsAttr.NestedObject.Attributes[attrName]
		if !ok {
			t.Errorf("teams nested object missing attribute %q", attrName)
			continue
		}
		strA, ok := a.(dsschema.StringAttribute)
		if !ok {
			t.Errorf("teams.%s should be StringAttribute, got %T", attrName, a)
			continue
		}
		if !strA.IsComputed() {
			t.Errorf("teams.%s should be Computed", attrName)
		}
		if strA.IsRequired() {
			t.Errorf("teams.%s should not be Required", attrName)
		}
	}
}
