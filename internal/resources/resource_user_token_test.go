package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestUserTokenResource_Schema(t *testing.T) {
	r := NewUserTokenResource().(*userTokenResource)
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

	// user_id: Computed only (populated after creation, not user-supplied)
	userIDRaw, ok := resp.Schema.Attributes["user_id"]
	if !ok {
		t.Fatal("missing 'user_id' attribute")
	}
	userIDAttr, ok := userIDRaw.(rsschema.StringAttribute)
	if !ok {
		t.Fatalf("user_id should be StringAttribute, got %T", userIDRaw)
	}
	if !userIDAttr.IsComputed() {
		t.Error("user_id should be Computed")
	}
	if userIDAttr.IsRequired() {
		t.Error("user_id should not be Required")
	}

	// name: Required (immutable)
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

	// token: Computed, Sensitive
	tokenRaw, ok := resp.Schema.Attributes["token"]
	if !ok {
		t.Fatal("missing 'token' attribute")
	}
	tokenAttr, ok := tokenRaw.(rsschema.StringAttribute)
	if !ok {
		t.Fatalf("token should be StringAttribute, got %T", tokenRaw)
	}
	if !tokenAttr.IsComputed() {
		t.Error("token should be Computed")
	}
	if !tokenAttr.IsSensitive() {
		t.Error("token should be Sensitive")
	}
	if tokenAttr.IsRequired() {
		t.Error("token should not be Required")
	}
}
