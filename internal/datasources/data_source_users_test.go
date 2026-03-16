package datasources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func TestUsersDataSource_Schema(t *testing.T) {
	d := NewUsersDataSource().(*usersDataSource)
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

	// users: Computed list nested attribute
	usersRaw, ok := resp.Schema.Attributes["users"]
	if !ok {
		t.Fatal("missing 'users' attribute")
	}
	usersAttr, ok := usersRaw.(dsschema.ListNestedAttribute)
	if !ok {
		t.Fatalf("users should be ListNestedAttribute, got %T", usersRaw)
	}
	if !usersAttr.IsComputed() {
		t.Error("users should be Computed")
	}

	// All nested user fields should be Computed
	for _, attrName := range []string{"id", "email", "name", "role"} {
		a, ok := usersAttr.NestedObject.Attributes[attrName]
		if !ok {
			t.Errorf("users nested object missing attribute %q", attrName)
			continue
		}
		strA, ok := a.(dsschema.StringAttribute)
		if !ok {
			t.Errorf("users.%s should be StringAttribute, got %T", attrName, a)
			continue
		}
		if !strA.IsComputed() {
			t.Errorf("users.%s should be Computed", attrName)
		}
		if strA.IsRequired() {
			t.Errorf("users.%s should not be Required", attrName)
		}
	}
}
