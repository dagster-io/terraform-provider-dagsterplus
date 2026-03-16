package datasources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func TestTeamDataSource_Schema(t *testing.T) {
	d := NewTeamDataSource().(*teamDataSource)
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

	// "id" must be Computed only
	idAttrRaw, ok := resp.Schema.Attributes["id"]
	if !ok {
		t.Fatal("missing 'id' attribute")
	}
	idAttr, ok := idAttrRaw.(dsschema.StringAttribute)
	if !ok {
		t.Fatalf("id should be StringAttribute, got %T", idAttrRaw)
	}
	if !idAttr.IsComputed() {
		t.Error("id should be Computed")
	}
	if idAttr.IsRequired() {
		t.Error("id should not be Required")
	}
}
