package datasources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func TestRoleDataSource_Schema(t *testing.T) {
	d := NewRoleDataSource().(*roleDataSource)
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	// "name" is the lookup key — must be Required
	nameAttrRaw, ok := resp.Schema.Attributes["name"]
	if !ok {
		t.Fatal("missing 'name' attribute")
	}
	nameAttr, ok := nameAttrRaw.(dsschema.StringAttribute)
	if !ok {
		t.Fatalf("name should be StringAttribute, got %T", nameAttrRaw)
	}
	if !nameAttr.IsRequired() {
		t.Error("name should be Required")
	}
	if nameAttr.IsComputed() {
		t.Error("name should not be Computed when it is the lookup key")
	}

	// Computed-only string attributes
	computedStringAttrs := []string{"id", "description", "icon", "role_type"}
	for _, attrName := range computedStringAttrs {
		attrRaw, ok := resp.Schema.Attributes[attrName]
		if !ok {
			t.Errorf("missing attribute %q", attrName)
			continue
		}
		strAttr, ok := attrRaw.(dsschema.StringAttribute)
		if !ok {
			t.Errorf("attribute %q should be StringAttribute, got %T", attrName, attrRaw)
			continue
		}
		if !strAttr.IsComputed() {
			t.Errorf("attribute %q should be Computed", attrName)
		}
		if strAttr.IsRequired() {
			t.Errorf("attribute %q should not be Required", attrName)
		}
	}

	// "permissions" is a Computed set
	permsAttrRaw, ok := resp.Schema.Attributes["permissions"]
	if !ok {
		t.Fatal("missing 'permissions' attribute")
	}
	permsAttr, ok := permsAttrRaw.(dsschema.SetAttribute)
	if !ok {
		t.Fatalf("permissions should be SetAttribute, got %T", permsAttrRaw)
	}
	if !permsAttr.IsComputed() {
		t.Error("permissions should be Computed")
	}
	if permsAttr.IsRequired() {
		t.Error("permissions should not be Required")
	}
}
