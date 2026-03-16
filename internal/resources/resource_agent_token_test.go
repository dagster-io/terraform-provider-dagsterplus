package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestAgentTokenResource_Schema(t *testing.T) {
	r := NewAgentTokenResource().(*agentTokenResource)
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

	// name: Required, not Computed
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
