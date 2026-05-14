package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestDeploymentResource_Schema(t *testing.T) {
	r := NewDeploymentResource().(*deploymentResource)
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	// id: Computed only
	idRaw, ok := resp.Schema.Attributes["id"]
	if !ok {
		t.Fatal("missing 'id' attribute")
	}
	idAttr, ok := idRaw.(rsschema.StringAttribute)
	if !ok {
		t.Fatalf("id should be StringAttribute, got %T", idRaw)
	}
	if !idAttr.IsComputed() {
		t.Error("id should be Computed")
	}
	if idAttr.IsRequired() {
		t.Error("id should not be Required")
	}

	// name: Required (immutable — triggers replacement)
	nameRaw, ok := resp.Schema.Attributes["name"]
	if !ok {
		t.Fatal("missing 'name' attribute")
	}
	nameAttr, ok := nameRaw.(rsschema.StringAttribute)
	if !ok {
		t.Fatalf("name should be StringAttribute, got %T", nameRaw)
	}
	if !nameAttr.IsRequired() {
		t.Error("name should be Required")
	}
	if nameAttr.IsComputed() {
		t.Error("name should not be Computed")
	}

	// deployment_id: Computed only, UseStateForUnknown
	deplIDRaw, ok := resp.Schema.Attributes["deployment_id"]
	if !ok {
		t.Fatal("missing 'deployment_id' attribute")
	}
	deplIDAttr, ok := deplIDRaw.(rsschema.StringAttribute)
	if !ok {
		t.Fatalf("deployment_id should be StringAttribute, got %T", deplIDRaw)
	}
	if !deplIDAttr.IsComputed() {
		t.Error("deployment_id should be Computed")
	}
	if deplIDAttr.IsRequired() || deplIDAttr.IsOptional() {
		t.Error("deployment_id should not be Required or Optional")
	}
}
