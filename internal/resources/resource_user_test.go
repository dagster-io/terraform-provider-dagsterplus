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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestUserResource_Schema(t *testing.T) {
	r := NewUserResource().(*userResource)
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

	// email: Required (immutable)
	emailRaw, ok := resp.Schema.Attributes["email"]
	if !ok {
		t.Fatal("missing 'email' attribute")
	}
	emailAttr, ok := emailRaw.(rsschema.StringAttribute)
	if !ok {
		t.Fatalf("email should be StringAttribute, got %T", emailRaw)
	}
	if !emailAttr.IsRequired() {
		t.Error("email should be Required")
	}
	if emailAttr.IsComputed() {
		t.Error("email should not be Computed")
	}

	// name: Computed only (set by Dagster+ after user accepts invite)
	nameRaw, ok := resp.Schema.Attributes["name"]
	if !ok {
		t.Fatal("missing 'name' attribute")
	}
	nameAttr, ok := nameRaw.(rsschema.StringAttribute)
	if !ok {
		t.Fatalf("name should be StringAttribute, got %T", nameRaw)
	}
	if !nameAttr.IsComputed() {
		t.Error("name should be Computed")
	}
	if nameAttr.IsRequired() {
		t.Error("name should not be Required")
	}

	// role: Computed only (API sets it)
	roleRaw, ok := resp.Schema.Attributes["role"]
	if !ok {
		t.Fatal("missing 'role' attribute")
	}
	roleAttr, ok := roleRaw.(rsschema.StringAttribute)
	if !ok {
		t.Fatalf("role should be StringAttribute, got %T", roleRaw)
	}
	if !roleAttr.IsComputed() {
		t.Error("role should be Computed")
	}
	if roleAttr.IsRequired() {
		t.Error("role should not be Required")
	}
}

// newInviteServer mocks a Dagster+ API whose addUserToOrganization answers with
// a PythonError, creating the user first when createsUser is set.
func newInviteServer(t *testing.T, createsUser bool) *httptest.Server {
	added := false
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Query string `json:"query"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)

		users := []any{}
		orgUsers := []any{}
		if added {
			users = append(users, map[string]any{
				"user":                        map[string]any{"userId": 42, "email": "alice@example.com", "name": "Alice"},
				"organizationPermissionGrant": map[string]any{"grant": "VIEWER"},
			})
			orgUsers = append(orgUsers, map[string]any{"userId": 42, "email": "alice@example.com", "name": "Alice"})
		}

		var data map[string]any
		switch {
		case strings.Contains(body.Query, "AddUser"):
			added = createsUser
			data = map[string]any{"addUserToOrganization": map[string]any{
				"__typename": "PythonError",
				"message":    "failed to send invite email",
			}}
		case strings.Contains(body.Query, "ListAllOrgUsers"):
			data = map[string]any{"organization": map[string]any{"usersOrError": map[string]any{
				"__typename": "DagsterCloudUserList",
				"users":      orgUsers,
			}}}
		case strings.Contains(body.Query, "ListUsers"):
			data = map[string]any{"usersOrError": map[string]any{
				"__typename": "DagsterCloudUsersWithScopedPermissionGrants",
				"users":      users,
			}}
		default:
			t.Errorf("unexpected query: %s", body.Query)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"data": data})
	}))
}

func createUser(t *testing.T, srv *httptest.Server) (*resource.CreateResponse, userResourceModel) {
	ctx := context.Background()
	r := NewUserResource().(*userResource)
	r.client = client.New("acc-org", "acc-token", srv.URL)

	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	sch := schemaResp.Schema

	plan := tfsdk.Plan{Schema: sch, Raw: tftypes.NewValue(sch.Type().TerraformType(ctx), nil)}
	diags := plan.Set(ctx, userResourceModel{
		ID:      types.StringUnknown(),
		Email:   types.StringValue("alice@example.com"),
		Name:    types.StringUnknown(),
		Role:    types.StringUnknown(),
		Picture: types.StringUnknown(),
	})
	if diags.HasError() {
		t.Fatalf("building plan: %v", diags)
	}

	resp := &resource.CreateResponse{State: tfsdk.State{Schema: sch, Raw: tftypes.NewValue(sch.Type().TerraformType(ctx), nil)}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, resp)

	var state userResourceModel
	if !resp.State.Raw.IsNull() {
		resp.Diagnostics.Append(resp.State.Get(ctx, &state)...)
	}
	return resp, state
}

func TestUserResource_Create_AdoptsUserCreatedDespiteError(t *testing.T) {
	srv := newInviteServer(t, true)
	defer srv.Close()

	resp, state := createUser(t, srv)

	if resp.Diagnostics.HasError() {
		t.Fatalf("expected the created user to be adopted, got errors: %v", resp.Diagnostics)
	}
	if resp.Diagnostics.WarningsCount() != 1 {
		t.Errorf("warnings = %d, want 1 carrying the invite error", resp.Diagnostics.WarningsCount())
	} else if detail := resp.Diagnostics.Warnings()[0].Detail(); !strings.Contains(detail, "failed to send invite email") {
		t.Errorf("warning detail = %q, want it to carry the API message", detail)
	}
	if state.ID.ValueString() != "42" {
		t.Errorf("state ID = %q, want 42", state.ID.ValueString())
	}
}

func TestUserResource_Create_ReportsErrorWhenUserMissing(t *testing.T) {
	srv := newInviteServer(t, false)
	defer srv.Close()

	resp, state := createUser(t, srv)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected Create to report an error when the user was not created")
	}
	if detail := resp.Diagnostics.Errors()[0].Detail(); !strings.Contains(detail, "failed to send invite email") {
		t.Errorf("error detail = %q, want it to carry the API message", detail)
	}
	if !state.ID.IsNull() && state.ID.ValueString() != "" {
		t.Errorf("state ID = %q, want no user in state", state.ID.ValueString())
	}
}
