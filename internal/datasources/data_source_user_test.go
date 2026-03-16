package datasources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func TestUserDataSource_Schema(t *testing.T) {
	d := NewUserDataSource().(*userDataSource)
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	// "email" is the lookup key — must be Required
	emailAttrRaw, ok := resp.Schema.Attributes["email"]
	if !ok {
		t.Fatal("missing 'email' attribute")
	}
	emailAttr, ok := emailAttrRaw.(dsschema.StringAttribute)
	if !ok {
		t.Fatalf("email should be StringAttribute, got %T", emailAttrRaw)
	}
	if !emailAttr.IsRequired() {
		t.Error("email should be Required")
	}
	if emailAttr.IsComputed() {
		t.Error("email should not be Computed when it is the lookup key")
	}

	// "id", "name", "role" must be Computed only
	computedAttrs := []string{"id", "name", "role"}
	for _, name := range computedAttrs {
		attrRaw, ok := resp.Schema.Attributes[name]
		if !ok {
			t.Errorf("missing attribute %q", name)
			continue
		}
		strAttr, ok := attrRaw.(dsschema.StringAttribute)
		if !ok {
			t.Errorf("attribute %q should be StringAttribute, got %T", name, attrRaw)
			continue
		}
		if !strAttr.IsComputed() {
			t.Errorf("attribute %q should be Computed", name)
		}
		if strAttr.IsRequired() {
			t.Errorf("attribute %q should not be Required", name)
		}
	}
}
