package datasources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func TestDeploymentDataSource_Schema(t *testing.T) {
	d := NewDeploymentDataSource().(*deploymentDataSource)
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	// "name" must be Required
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
		t.Error("name should not be Computed (it is the lookup key)")
	}

	// "type" must be Computed only
	typeAttrRaw, ok := resp.Schema.Attributes["type"]
	if !ok {
		t.Fatal("missing 'type' attribute")
	}
	typeAttr, ok := typeAttrRaw.(dsschema.StringAttribute)
	if !ok {
		t.Fatalf("type should be StringAttribute, got %T", typeAttrRaw)
	}
	if !typeAttr.IsComputed() {
		t.Error("type should be Computed")
	}
	if typeAttr.IsRequired() {
		t.Error("type should not be Required")
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

	// "deployment_id" must be Computed only
	deplIDRaw, ok := resp.Schema.Attributes["deployment_id"]
	if !ok {
		t.Fatal("missing 'deployment_id' attribute")
	}
	deplIDAttr, ok := deplIDRaw.(dsschema.StringAttribute)
	if !ok {
		t.Fatalf("deployment_id should be StringAttribute, got %T", deplIDRaw)
	}
	if !deplIDAttr.IsComputed() {
		t.Error("deployment_id should be Computed")
	}
	if deplIDAttr.IsRequired() {
		t.Error("deployment_id should not be Required")
	}
}
