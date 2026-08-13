package resources

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestDeploymentResource_Schema(t *testing.T) {
	r := NewDeploymentResource().(*deploymentResource)
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

	// name: Required (immutable — triggers replacement)
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

	// agent_type: Optional + Computed (the API always reports one), enum-validated
	agentTypeRaw, ok := resp.Schema.Attributes["agent_type"]
	if !ok {
		t.Fatal("missing 'agent_type' attribute")
	}
	agentTypeAttr, ok := agentTypeRaw.(rsschema.StringAttribute)
	if !ok {
		t.Fatalf("agent_type should be StringAttribute, got %T", agentTypeRaw)
	}
	if !agentTypeAttr.IsOptional() {
		t.Error("agent_type should be Optional")
	}
	if !agentTypeAttr.IsComputed() {
		t.Error("agent_type should be Computed")
	}
	if len(agentTypeAttr.Validators) == 0 {
		t.Error("agent_type should have a OneOf validator")
	}
	for _, v := range deploymentAgentTypes {
		if !strings.Contains(agentTypeAttr.Description, v) {
			t.Errorf("agent_type description should list %q, got: %s", v, agentTypeAttr.Description)
		}
	}

	// deployment_id: Computed only, UseStateForUnknown
	deplIDRaw, ok := resp.Schema.Attributes["deployment_id"]
	if !ok {
		t.Fatal("missing 'deployment_id' attribute")
	}
	deplIDAttr, ok := deplIDRaw.(rsschema.StringAttribute)
	if !ok {
		t.Fatalf("deployment_id should be StringAttribute, got %T", deplIDRaw)
	}
	if !deplIDAttr.IsComputed() {
		t.Error("deployment_id should be Computed")
	}
	if deplIDAttr.IsRequired() || deplIDAttr.IsOptional() {
		t.Error("deployment_id should not be Required or Optional")
	}
}
