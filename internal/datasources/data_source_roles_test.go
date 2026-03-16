package datasources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func TestRolesDataSource_Schema(t *testing.T) {
	d := NewRolesDataSource().(*rolesDataSource)
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

	// roles: Computed list nested attribute
	rolesRaw, ok := resp.Schema.Attributes["roles"]
	if !ok {
		t.Fatal("missing 'roles' attribute")
	}
	rolesAttr, ok := rolesRaw.(dsschema.ListNestedAttribute)
	if !ok {
		t.Fatalf("roles should be ListNestedAttribute, got %T", rolesRaw)
	}
	if !rolesAttr.IsComputed() {
		t.Error("roles should be Computed")
	}

	// All nested role fields should be Computed
	for _, attrName := range []string{"id", "name", "description", "icon", "role_type"} {
		a, ok := rolesAttr.NestedObject.Attributes[attrName]
		if !ok {
			t.Errorf("roles nested object missing attribute %q", attrName)
			continue
		}
		strA, ok := a.(dsschema.StringAttribute)
		if !ok {
			t.Errorf("roles.%s should be StringAttribute, got %T", attrName, a)
			continue
		}
		if !strA.IsComputed() {
			t.Errorf("roles.%s should be Computed", attrName)
		}
	}

	// permissions: Computed set
	permRaw, ok := rolesAttr.NestedObject.Attributes["permissions"]
	if !ok {
		t.Fatal("roles nested object missing 'permissions' attribute")
	}
	permAttr, ok := permRaw.(dsschema.SetAttribute)
	if !ok {
		t.Fatalf("permissions should be SetAttribute, got %T", permRaw)
	}
	if !permAttr.IsComputed() {
		t.Error("permissions should be Computed")
	}
}
