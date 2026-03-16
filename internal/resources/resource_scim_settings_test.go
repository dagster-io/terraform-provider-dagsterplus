package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestScimSettingsResource_Schema(t *testing.T) {
	r := NewScimSettingsResource().(*scimSettingsResource)
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

	// enabled: Required bool
	enabledRaw, ok := resp.Schema.Attributes["enabled"]
	if !ok {
		t.Fatal("missing 'enabled' attribute")
	}
	enabledAttr, ok := enabledRaw.(rsschema.BoolAttribute)
	if !ok {
		t.Fatalf("enabled should be BoolAttribute, got %T", enabledRaw)
	}
	if !enabledAttr.IsRequired() {
		t.Error("enabled should be Required")
	}
	if enabledAttr.IsComputed() {
		t.Error("enabled should not be Computed")
	}
}
