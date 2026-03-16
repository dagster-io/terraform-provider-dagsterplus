package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestServiceTokenResource_Schema(t *testing.T) {
	r := NewServiceTokenResource().(*serviceTokenResource)
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

	// service_user_id: Required (immutable)
	suidRaw, ok := resp.Schema.Attributes["service_user_id"]
	if !ok {
		t.Fatal("missing 'service_user_id' attribute")
	}
	suidAttr, ok := suidRaw.(rsschema.StringAttribute)
	if !ok {
		t.Fatalf("service_user_id should be StringAttribute, got %T", suidRaw)
	}
	if !suidAttr.IsRequired() {
		t.Error("service_user_id should be Required")
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
