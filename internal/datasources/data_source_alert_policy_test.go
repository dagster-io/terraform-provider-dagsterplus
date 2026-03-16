package datasources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func TestAlertPolicyDataSource_Schema(t *testing.T) {
	d := NewAlertPolicyDataSource().(*alertPolicyDataSource)
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	// "deployment" and "name" and "policy_type" are Required lookup keys
	requiredAttrs := []string{"deployment", "name", "policy_type"}
	for _, attrName := range requiredAttrs {
		attrRaw, ok := resp.Schema.Attributes[attrName]
		if !ok {
			t.Errorf("missing required attribute %q", attrName)
			continue
		}
		strAttr, ok := attrRaw.(dsschema.StringAttribute)
		if !ok {
			t.Errorf("attribute %q should be StringAttribute, got %T", attrName, attrRaw)
			continue
		}
		if !strAttr.IsRequired() {
			t.Errorf("attribute %q should be Required", attrName)
		}
	}

	// "id", "description", "enabled", "event_types" must be Computed only
	t.Run("id is Computed", func(t *testing.T) {
		attrRaw, ok := resp.Schema.Attributes["id"]
		if !ok {
			t.Fatal("missing 'id' attribute")
		}
		strAttr, ok := attrRaw.(dsschema.StringAttribute)
		if !ok {
			t.Fatalf("id should be StringAttribute, got %T", attrRaw)
		}
		if !strAttr.IsComputed() {
			t.Error("id should be Computed")
		}
		if strAttr.IsRequired() {
			t.Error("id should not be Required")
		}
	})

	t.Run("description is Computed", func(t *testing.T) {
		attrRaw, ok := resp.Schema.Attributes["description"]
		if !ok {
			t.Fatal("missing 'description' attribute")
		}
		strAttr, ok := attrRaw.(dsschema.StringAttribute)
		if !ok {
			t.Fatalf("description should be StringAttribute, got %T", attrRaw)
		}
		if !strAttr.IsComputed() {
			t.Error("description should be Computed")
		}
		if strAttr.IsRequired() {
			t.Error("description should not be Required")
		}
	})

	t.Run("enabled is Computed", func(t *testing.T) {
		attrRaw, ok := resp.Schema.Attributes["enabled"]
		if !ok {
			t.Fatal("missing 'enabled' attribute")
		}
		boolAttr, ok := attrRaw.(dsschema.BoolAttribute)
		if !ok {
			t.Fatalf("enabled should be BoolAttribute, got %T", attrRaw)
		}
		if !boolAttr.IsComputed() {
			t.Error("enabled should be Computed")
		}
		if boolAttr.IsRequired() {
			t.Error("enabled should not be Required")
		}
	})

	t.Run("event_types is Computed list", func(t *testing.T) {
		attrRaw, ok := resp.Schema.Attributes["event_types"]
		if !ok {
			t.Fatal("missing 'event_types' attribute")
		}
		listAttr, ok := attrRaw.(dsschema.ListAttribute)
		if !ok {
			t.Fatalf("event_types should be ListAttribute, got %T", attrRaw)
		}
		if !listAttr.IsComputed() {
			t.Error("event_types should be Computed")
		}
		if listAttr.IsRequired() {
			t.Error("event_types should not be Required")
		}
	})

	// Nested blocks should exist
	nestedBlocks := []string{"asset", "run", "code_location", "automation", "budget", "insight_metric", "notification_service"}
	for _, blockName := range nestedBlocks {
		if _, ok := resp.Schema.Blocks[blockName]; !ok {
			t.Errorf("missing block %q in alert policy data source schema", blockName)
		}
	}
}
