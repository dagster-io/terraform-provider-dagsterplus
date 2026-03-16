package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestDeploymentSettingsResource_Schema(t *testing.T) {
	r := NewDeploymentSettingsResource().(*deploymentSettingsResource)
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

	// deployment: Required (immutable)
	deploymentRaw, ok := resp.Schema.Attributes["deployment"]
	if !ok {
		t.Fatal("missing 'deployment' attribute")
	}
	deploymentAttr, ok := deploymentRaw.(rsschema.StringAttribute)
	if !ok {
		t.Fatalf("deployment should be StringAttribute, got %T", deploymentRaw)
	}
	if !deploymentAttr.IsRequired() {
		t.Error("deployment should be Required")
	}

	// settings_json: Required
	settingsRaw, ok := resp.Schema.Attributes["settings_json"]
	if !ok {
		t.Fatal("missing 'settings_json' attribute")
	}
	settingsAttr, ok := settingsRaw.(rsschema.StringAttribute)
	if !ok {
		t.Fatalf("settings_json should be StringAttribute, got %T", settingsRaw)
	}
	if !settingsAttr.IsRequired() {
		t.Error("settings_json should be Required")
	}
	if settingsAttr.IsComputed() {
		t.Error("settings_json should not be Computed")
	}
}
