package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestTeamResource_Schema(t *testing.T) {
	r := NewTeamResource().(*teamResource)
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

	// name: Required
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

	// blocks: organization_grant, all_branch_deployments_grant, deployment_grant, member
	for _, blockName := range []string{"organization_grant", "all_branch_deployments_grant", "deployment_grant", "member"} {
		if _, ok := resp.Schema.Blocks[blockName]; !ok {
			t.Errorf("missing block %q", blockName)
		}
	}

	// organization_grant is a ListNestedBlock
	orgGrantBlock, ok := resp.Schema.Blocks["organization_grant"]
	if !ok {
		t.Fatal("missing organization_grant block")
	}
	if _, ok := orgGrantBlock.(rsschema.ListNestedBlock); !ok {
		t.Errorf("organization_grant should be ListNestedBlock, got %T", orgGrantBlock)
	}

	// deployment_grant block has deployment, grant, custom_role_id attributes
	deployGrantBlock, ok := resp.Schema.Blocks["deployment_grant"]
	if !ok {
		t.Fatal("missing deployment_grant block")
	}
	deployGrantNested, ok := deployGrantBlock.(rsschema.ListNestedBlock)
	if !ok {
		t.Fatalf("deployment_grant should be ListNestedBlock, got %T", deployGrantBlock)
	}
	for _, attr := range []string{"deployment", "grant", "custom_role_id"} {
		if _, ok := deployGrantNested.NestedObject.Attributes[attr]; !ok {
			t.Errorf("deployment_grant block missing attribute %q", attr)
		}
	}
}
