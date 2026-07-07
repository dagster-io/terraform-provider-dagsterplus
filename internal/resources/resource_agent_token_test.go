package resources

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestAgentTokenResource_Schema(t *testing.T) {
	r := NewAgentTokenResource().(*agentTokenResource)
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

	// name: Required, not Computed
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

	// token: Computed, Sensitive
	tokenRaw, ok := resp.Schema.Attributes["token"]
	if !ok {
		t.Fatal("missing 'token' attribute")
	}
	tokenAttr, ok := tokenRaw.(rsschema.StringAttribute)
	if !ok {
		t.Fatalf("token should be StringAttribute, got %T", tokenRaw)
	}
	if !tokenAttr.IsComputed() {
		t.Error("token should be Computed")
	}
	if !tokenAttr.IsSensitive() {
		t.Error("token should be Sensitive")
	}
	if tokenAttr.IsRequired() {
		t.Error("token should not be Required")
	}
}

// TestAgentTokenResource_Create_GrantFailurePersistsToken covers the partial-create
// path described in issue #44: when the token is created remotely but a subsequent
// grant call fails, Create must still record the token in Terraform state so a later
// apply reconciles the existing token instead of leaking a new one.
func TestAgentTokenResource_Create_GrantFailurePersistsToken(t *testing.T) {
	ctx := context.Background()

	// Server: CreateAgentToken succeeds; the grant mutation fails with a
	// PythonError so applyGrants returns an error mid-Create.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Query string `json:"query"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)

		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(body.Query, "CreateAgentToken"):
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{
				"createAgentToken": map[string]any{
					"__typename":  "DagsterCloudAgentToken",
					"id":          123,
					"description": "acc-tf-issue44",
					"token":       "secret-value",
				},
			}})
		case strings.Contains(body.Query, "SetAgentGrant"):
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{
				"createOrUpdateAgentPermissions": map[string]any{
					"__typename": "PythonError",
				},
			}})
		default:
			t.Errorf("unexpected query: %s", body.Query)
		}
	}))
	defer srv.Close()

	r := NewAgentTokenResource().(*agentTokenResource)
	r.client = client.New("acc-org", "acc-token", srv.URL)

	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	sch := schemaResp.Schema

	planRaw := tftypes.NewValue(sch.Type().TerraformType(ctx), map[string]tftypes.Value{
		"id":                        tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"name":                      tftypes.NewValue(tftypes.String, "acc-tf-issue44"),
		"token":                     tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"organization":              tftypes.NewValue(tftypes.Bool, true),
		"all_branch_deployments":    tftypes.NewValue(tftypes.Bool, false),
		"deployment_grants":         tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
		"branch_deployments_grants": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
	})

	req := resource.CreateRequest{Plan: tfsdk.Plan{Schema: sch, Raw: planRaw}}
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: sch}}
	r.Create(ctx, req, resp)

	// The grant failure must surface as an error...
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected Create to report an error when the grant call fails")
	}

	// ...but the token must still be recorded in state so it is not leaked.
	var state agentTokenResourceModel
	resp.Diagnostics.Append(resp.State.Get(ctx, &state)...)
	if state.ID.ValueString() != "123" {
		t.Errorf("state ID = %q, want 123 (token must be persisted despite grant failure)", state.ID.ValueString())
	}
	if state.Token.ValueString() != "secret-value" {
		t.Errorf("state Token = %q, want secret-value", state.Token.ValueString())
	}
}
