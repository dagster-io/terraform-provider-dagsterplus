package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestTeamMembershipResource_Schema(t *testing.T) {
	r := NewTeamMembershipResource().(*teamMembershipResource)
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

	// team_id: Required (immutable)
	teamIDRaw, ok := resp.Schema.Attributes["team_id"]
	if !ok {
		t.Fatal("missing 'team_id' attribute")
	}
	teamIDAttr, ok := teamIDRaw.(rsschema.StringAttribute)
	if !ok {
		t.Fatalf("team_id should be StringAttribute, got %T", teamIDRaw)
	}
	if !teamIDAttr.IsRequired() {
		t.Error("team_id should be Required")
	}

	// user_id: Required (immutable)
	userIDRaw, ok := resp.Schema.Attributes["user_id"]
	if !ok {
		t.Fatal("missing 'user_id' attribute")
	}
	userIDAttr, ok := userIDRaw.(rsschema.StringAttribute)
	if !ok {
		t.Fatalf("user_id should be StringAttribute, got %T", userIDRaw)
	}
	if !userIDAttr.IsRequired() {
		t.Error("user_id should be Required")
	}
}
