package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestAtlanIntegrationResource_Schema(t *testing.T) {
	r := NewAtlanIntegrationResource().(*atlanIntegrationResource)
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

	// token: Required, Sensitive
	tokenRaw, ok := resp.Schema.Attributes["token"]
	if !ok {
		t.Fatal("missing 'token' attribute")
	}
	tokenAttr, ok := tokenRaw.(rsschema.StringAttribute)
	if !ok {
		t.Fatalf("token should be StringAttribute, got %T", tokenRaw)
	}
	if !tokenAttr.IsRequired() {
		t.Error("token should be Required")
	}
	if !tokenAttr.IsSensitive() {
		t.Error("token should be Sensitive")
	}
	if tokenAttr.IsComputed() {
		t.Error("token should not be Computed")
	}

	// domain: Required, not Sensitive
	domainRaw, ok := resp.Schema.Attributes["domain"]
	if !ok {
		t.Fatal("missing 'domain' attribute")
	}
	domainAttr, ok := domainRaw.(rsschema.StringAttribute)
	if !ok {
		t.Fatalf("domain should be StringAttribute, got %T", domainRaw)
	}
	if !domainAttr.IsRequired() {
		t.Error("domain should be Required")
	}
	if domainAttr.IsSensitive() {
		t.Error("domain should not be Sensitive")
	}
}
