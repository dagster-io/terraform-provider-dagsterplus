package datasources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func TestSecretDataSource_Schema(t *testing.T) {
	d := NewSecretDataSource().(*secretDataSource)
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	// "secret_name" is the lookup key — must be Required
	snAttrRaw, ok := resp.Schema.Attributes["secret_name"]
	if !ok {
		t.Fatal("missing 'secret_name' attribute")
	}
	snAttr, ok := snAttrRaw.(dsschema.StringAttribute)
	if !ok {
		t.Fatalf("secret_name should be StringAttribute, got %T", snAttrRaw)
	}
	if !snAttr.IsRequired() {
		t.Error("secret_name should be Required")
	}

	// "secret_value" must be Computed and Sensitive
	svAttrRaw, ok := resp.Schema.Attributes["secret_value"]
	if !ok {
		t.Fatal("missing 'secret_value' attribute")
	}
	svAttr, ok := svAttrRaw.(dsschema.StringAttribute)
	if !ok {
		t.Fatalf("secret_value should be StringAttribute, got %T", svAttrRaw)
	}
	if !svAttr.IsComputed() {
		t.Error("secret_value should be Computed")
	}
	if svAttr.IsRequired() {
		t.Error("secret_value should not be Required")
	}
	if !svAttr.IsSensitive() {
		t.Error("secret_value should be Sensitive")
	}

	// Remaining attributes should be Computed only
	computedAttrs := []string{
		"id",
		"full_deployment_scope",
		"all_branch_deployments_scope",
		"specific_branch_deployment_scope",
		"local_deployment_scope",
		"location_names",
	}
	for _, name := range computedAttrs {
		attr, ok := resp.Schema.Attributes[name]
		if !ok {
			t.Errorf("missing attribute %q", name)
			continue
		}
		switch a := attr.(type) {
		case dsschema.StringAttribute:
			if !a.IsComputed() {
				t.Errorf("attribute %q should be Computed", name)
			}
			if a.IsRequired() {
				t.Errorf("attribute %q should not be Required", name)
			}
		case dsschema.BoolAttribute:
			if !a.IsComputed() {
				t.Errorf("attribute %q should be Computed", name)
			}
			if a.IsRequired() {
				t.Errorf("attribute %q should not be Required", name)
			}
		case dsschema.ListAttribute:
			if !a.IsComputed() {
				t.Errorf("attribute %q should be Computed", name)
			}
			if a.IsRequired() {
				t.Errorf("attribute %q should not be Required", name)
			}
		default:
			t.Errorf("unexpected attribute type for %q: %T", name, attr)
		}
	}
}
