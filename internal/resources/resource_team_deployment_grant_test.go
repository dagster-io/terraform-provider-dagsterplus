package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestTeamDeploymentGrantResource_Schema(t *testing.T) {
	r := NewTeamDeploymentGrantResource().(*teamDeploymentGrantResource)
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

	// grant: Optional (conflicts with custom_role_id)
	grantRaw, ok := resp.Schema.Attributes["grant"]
	if !ok {
		t.Fatal("missing 'grant' attribute")
	}
	grantAttr, ok := grantRaw.(rsschema.StringAttribute)
	if !ok {
		t.Fatalf("grant should be StringAttribute, got %T", grantRaw)
	}
	if !grantAttr.IsOptional() {
		t.Error("grant should be Optional")
	}

	// custom_role_id: Optional (conflicts with grant)
	customRoleIDRaw, ok := resp.Schema.Attributes["custom_role_id"]
	if !ok {
		t.Fatal("missing 'custom_role_id' attribute")
	}
	customRoleIDAttr, ok := customRoleIDRaw.(rsschema.StringAttribute)
	if !ok {
		t.Fatalf("custom_role_id should be StringAttribute, got %T", customRoleIDRaw)
	}
	if !customRoleIDAttr.IsOptional() {
		t.Error("custom_role_id should be Optional")
	}
}

// TestGrantFieldsFromAPI verifies that CUSTOM grants are represented by custom_role_id only.
func TestGrantFieldsFromAPI(t *testing.T) {
	tests := []struct {
		name             string
		grant            string
		customRoleID     string
		expectGrant      string
		expectGrantNull  bool
		expectCustomID   string
		expectCustomNull bool
	}{
		{
			name:             "standard grant VIEWER",
			grant:            "VIEWER",
			customRoleID:     "",
			expectGrant:      "VIEWER",
			expectGrantNull:  false,
			expectCustomNull: true,
		},
		{
			name:             "standard grant ADMIN",
			grant:            "ADMIN",
			customRoleID:     "",
			expectGrant:      "ADMIN",
			expectGrantNull:  false,
			expectCustomNull: true,
		},
		{
			name:             "CUSTOM grant with role ID",
			grant:            "CUSTOM",
			customRoleID:     "role-abc",
			expectGrantNull:  true,
			expectCustomID:   "role-abc",
			expectCustomNull: false,
		},
		{
			name:             "CUSTOM grant with empty role ID",
			grant:            "CUSTOM",
			customRoleID:     "",
			expectGrantNull:  true,
			expectCustomNull: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			grantVal, customRoleIDVal := grantFieldsFromAPI(tc.grant, tc.customRoleID)

			if tc.expectGrantNull {
				if !grantVal.IsNull() {
					t.Errorf("grant should be null, got %q", grantVal.ValueString())
				}
			} else {
				if grantVal.IsNull() {
					t.Error("grant should not be null")
				}
				if grantVal.ValueString() != tc.expectGrant {
					t.Errorf("grant: want %q, got %q", tc.expectGrant, grantVal.ValueString())
				}
			}

			if tc.expectCustomNull {
				if !customRoleIDVal.IsNull() {
					t.Errorf("custom_role_id should be null, got %q", customRoleIDVal.ValueString())
				}
			} else {
				if customRoleIDVal.IsNull() {
					t.Error("custom_role_id should not be null")
				}
				if customRoleIDVal.ValueString() != tc.expectCustomID {
					t.Errorf("custom_role_id: want %q, got %q", tc.expectCustomID, customRoleIDVal.ValueString())
				}
			}
		})
	}
}

// TestResolveGrantFields verifies that custom_role_id takes precedence over grant.
func TestResolveGrantFields(t *testing.T) {
	tests := []struct {
		name           string
		grant          types.String
		customRoleID   types.String
		expectGrant    string
		expectCustomID string
	}{
		{
			name:           "standard grant",
			grant:          types.StringValue("VIEWER"),
			customRoleID:   types.StringNull(),
			expectGrant:    "VIEWER",
			expectCustomID: "",
		},
		{
			name:           "custom role overrides grant",
			grant:          types.StringNull(),
			customRoleID:   types.StringValue("role-xyz"),
			expectGrant:    "CUSTOM",
			expectCustomID: "role-xyz",
		},
		{
			name:           "empty custom_role_id falls back to grant",
			grant:          types.StringValue("EDITOR"),
			customRoleID:   types.StringValue(""),
			expectGrant:    "EDITOR",
			expectCustomID: "",
		},
		{
			name:           "both null",
			grant:          types.StringNull(),
			customRoleID:   types.StringNull(),
			expectGrant:    "",
			expectCustomID: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			grantStr, customRoleIDStr := resolveGrantFields(tc.grant, tc.customRoleID)
			if grantStr != tc.expectGrant {
				t.Errorf("grant: want %q, got %q", tc.expectGrant, grantStr)
			}
			if customRoleIDStr != tc.expectCustomID {
				t.Errorf("customRoleID: want %q, got %q", tc.expectCustomID, customRoleIDStr)
			}
		})
	}
}

// TestNullableString verifies that empty strings become null.
func TestNullableString(t *testing.T) {
	if !nullableString("").IsNull() {
		t.Error("empty string should produce null types.String")
	}
	v := nullableString("hello")
	if v.IsNull() {
		t.Error("non-empty string should not produce null types.String")
	}
	if v.ValueString() != "hello" {
		t.Errorf("expected hello, got %q", v.ValueString())
	}
}

// TestGrantToOrgModel verifies that CUSTOM grants are represented by custom_role_id only.
func TestGrantToOrgModel_StandardGrant(t *testing.T) {
	// We test indirectly via grantFieldsFromAPI since grantToOrgModel is a thin wrapper.
	grantVal, customRoleIDVal := grantFieldsFromAPI("EDITOR", "")
	if grantVal.ValueString() != "EDITOR" {
		t.Errorf("expected EDITOR, got %q", grantVal.ValueString())
	}
	if !customRoleIDVal.IsNull() {
		t.Error("custom_role_id should be null for standard grant")
	}
}

func TestGrantToOrgModel_CustomGrant(t *testing.T) {
	grantVal, customRoleIDVal := grantFieldsFromAPI("CUSTOM", "role-999")
	if !grantVal.IsNull() {
		t.Errorf("grant should be null for CUSTOM, got %q", grantVal.ValueString())
	}
	if customRoleIDVal.ValueString() != "role-999" {
		t.Errorf("expected role-999, got %q", customRoleIDVal.ValueString())
	}
}
