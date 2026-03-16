package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestCustomMetricResource_Schema(t *testing.T) {
	r := NewCustomMetricResource().(*customMetricResource)
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

	// metadata_key: Required (immutable)
	keyRaw, ok := resp.Schema.Attributes["metadata_key"]
	if !ok {
		t.Fatal("missing 'metadata_key' attribute")
	}
	keyAttr, ok := keyRaw.(rsschema.StringAttribute)
	if !ok {
		t.Fatalf("metadata_key should be StringAttribute, got %T", keyRaw)
	}
	if !keyAttr.IsRequired() {
		t.Error("metadata_key should be Required")
	}

	// display_name: Optional+Computed
	displayNameRaw, ok := resp.Schema.Attributes["display_name"]
	if !ok {
		t.Fatal("missing 'display_name' attribute")
	}
	displayNameAttr, ok := displayNameRaw.(rsschema.StringAttribute)
	if !ok {
		t.Fatalf("display_name should be StringAttribute, got %T", displayNameRaw)
	}
	if !displayNameAttr.IsOptional() {
		t.Error("display_name should be Optional")
	}
	if !displayNameAttr.IsComputed() {
		t.Error("display_name should be Computed")
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

	// create_timestamp: Computed only
	ctRaw, ok := resp.Schema.Attributes["create_timestamp"]
	if !ok {
		t.Fatal("missing 'create_timestamp' attribute")
	}
	ctAttr, ok := ctRaw.(rsschema.Float64Attribute)
	if !ok {
		t.Fatalf("create_timestamp should be Float64Attribute, got %T", ctRaw)
	}
	if !ctAttr.IsComputed() {
		t.Error("create_timestamp should be Computed")
	}
	if ctAttr.IsRequired() {
		t.Error("create_timestamp should not be Required")
	}

	// update_timestamp: Computed only
	utRaw, ok := resp.Schema.Attributes["update_timestamp"]
	if !ok {
		t.Fatal("missing 'update_timestamp' attribute")
	}
	utAttr, ok := utRaw.(rsschema.Float64Attribute)
	if !ok {
		t.Fatalf("update_timestamp should be Float64Attribute, got %T", utRaw)
	}
	if !utAttr.IsComputed() {
		t.Error("update_timestamp should be Computed")
	}
}
