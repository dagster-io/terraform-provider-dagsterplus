package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestServiceUserResource_Schema(t *testing.T) {
	r := NewServiceUserResource().(*serviceUserResource)
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

	// name: Required
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

	// description: Optional+Computed
	descRaw, ok := resp.Schema.Attributes["description"]
	if !ok {
		t.Fatal("missing 'description' attribute")
	}
	descAttr, ok := descRaw.(rsschema.StringAttribute)
	if !ok {
		t.Fatalf("description should be StringAttribute, got %T", descRaw)
	}
	if !descAttr.IsOptional() {
		t.Error("description should be Optional")
	}
	if !descAttr.IsComputed() {
		t.Error("description should be Computed")
	}
}
