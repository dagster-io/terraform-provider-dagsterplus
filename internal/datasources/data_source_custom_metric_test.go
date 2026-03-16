package datasources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func TestCustomMetricDataSource_Schema(t *testing.T) {
	d := NewCustomMetricDataSource().(*customMetricDataSource)
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	// metadata_key: Required (lookup key)
	keyRaw, ok := resp.Schema.Attributes["metadata_key"]
	if !ok {
		t.Fatal("missing 'metadata_key' attribute")
	}
	keyAttr, ok := keyRaw.(dsschema.StringAttribute)
	if !ok {
		t.Fatalf("metadata_key should be StringAttribute, got %T", keyRaw)
	}
	if !keyAttr.IsRequired() {
		t.Error("metadata_key should be Required")
	}
	if keyAttr.IsComputed() {
		t.Error("metadata_key should not be Computed")
	}

	// id: Computed only
	idRaw, ok := resp.Schema.Attributes["id"]
	if !ok {
		t.Fatal("missing 'id' attribute")
	}
	idAttr, ok := idRaw.(dsschema.StringAttribute)
	if !ok {
		t.Fatalf("id should be StringAttribute, got %T", idRaw)
	}
	if !idAttr.IsComputed() {
		t.Error("id should be Computed")
	}
	if idAttr.IsRequired() {
		t.Error("id should not be Required")
	}

	// display_name: Computed only
	displayNameRaw, ok := resp.Schema.Attributes["display_name"]
	if !ok {
		t.Fatal("missing 'display_name' attribute")
	}
	displayNameAttr, ok := displayNameRaw.(dsschema.StringAttribute)
	if !ok {
		t.Fatalf("display_name should be StringAttribute, got %T", displayNameRaw)
	}
	if !displayNameAttr.IsComputed() {
		t.Error("display_name should be Computed")
	}

	// create_timestamp: Computed
	ctRaw, ok := resp.Schema.Attributes["create_timestamp"]
	if !ok {
		t.Fatal("missing 'create_timestamp' attribute")
	}
	ctAttr, ok := ctRaw.(dsschema.Float64Attribute)
	if !ok {
		t.Fatalf("create_timestamp should be Float64Attribute, got %T", ctRaw)
	}
	if !ctAttr.IsComputed() {
		t.Error("create_timestamp should be Computed")
	}

	// update_timestamp: Computed
	utRaw, ok := resp.Schema.Attributes["update_timestamp"]
	if !ok {
		t.Fatal("missing 'update_timestamp' attribute")
	}
	utAttr, ok := utRaw.(dsschema.Float64Attribute)
	if !ok {
		t.Fatalf("update_timestamp should be Float64Attribute, got %T", utRaw)
	}
	if !utAttr.IsComputed() {
		t.Error("update_timestamp should be Computed")
	}
}
