package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestUserResource_Schema(t *testing.T) {
	r := NewUserResource().(*userResource)
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

	// email: Required (immutable)
	emailRaw, ok := resp.Schema.Attributes["email"]
	if !ok {
		t.Fatal("missing 'email' attribute")
	}
	emailAttr, ok := emailRaw.(rsschema.StringAttribute)
	if !ok {
		t.Fatalf("email should be StringAttribute, got %T", emailRaw)
	}
	if !emailAttr.IsRequired() {
		t.Error("email should be Required")
	}
	if emailAttr.IsComputed() {
		t.Error("email should not be Computed")
	}

	// name: Computed only (set by Dagster+ after user accepts invite)
	nameRaw, ok := resp.Schema.Attributes["name"]
	if !ok {
		t.Fatal("missing 'name' attribute")
	}
	nameAttr, ok := nameRaw.(rsschema.StringAttribute)
	if !ok {
		t.Fatalf("name should be StringAttribute, got %T", nameRaw)
	}
	if !nameAttr.IsComputed() {
		t.Error("name should be Computed")
	}
	if nameAttr.IsRequired() {
		t.Error("name should not be Required")
	}

	// role: Computed only (API sets it)
	roleRaw, ok := resp.Schema.Attributes["role"]
	if !ok {
		t.Fatal("missing 'role' attribute")
	}
	roleAttr, ok := roleRaw.(rsschema.StringAttribute)
	if !ok {
		t.Fatalf("role should be StringAttribute, got %T", roleRaw)
	}
	if !roleAttr.IsComputed() {
		t.Error("role should be Computed")
	}
	if roleAttr.IsRequired() {
		t.Error("role should not be Required")
	}
}
