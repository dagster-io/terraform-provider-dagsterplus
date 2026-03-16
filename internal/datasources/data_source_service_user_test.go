package datasources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func TestServiceUserDataSource_Schema(t *testing.T) {
	d := NewServiceUserDataSource().(*serviceUserDataSource)
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	// name: Required (lookup key)
	nameRaw, ok := resp.Schema.Attributes["name"]
	if !ok {
		t.Fatal("missing 'name' attribute")
	}
	nameAttr, ok := nameRaw.(dsschema.StringAttribute)
	if !ok {
		t.Fatalf("name should be StringAttribute, got %T", nameRaw)
	}
	if !nameAttr.IsRequired() {
		t.Error("name should be Required")
	}
	if nameAttr.IsComputed() {
		t.Error("name should not be Computed")
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
	if idAttr.IsRequired() {
		t.Error("id should not be Required")
	}

	// description: Computed only
	descRaw, ok := resp.Schema.Attributes["description"]
	if !ok {
		t.Fatal("missing 'description' attribute")
	}
	descAttr, ok := descRaw.(dsschema.StringAttribute)
	if !ok {
		t.Fatalf("description should be StringAttribute, got %T", descRaw)
	}
	if !descAttr.IsComputed() {
		t.Error("description should be Computed")
	}
	if descAttr.IsRequired() {
		t.Error("description should not be Required")
	}
}
