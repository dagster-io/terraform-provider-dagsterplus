package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
)

// --- shared test helpers ---

// newClient creates a Client pointing at srv with a fixed test token.
func newClient(srv *httptest.Server) *client.Client {
	return client.New("test-org", "test-token", srv.URL)
}

// gqlBody is the decoded GraphQL request body received by a mock server.
type gqlBody struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

// parseBody decodes a GraphQL request body; fails the test on error.
func parseBody(t *testing.T, r *http.Request) gqlBody {
	t.Helper()
	var b gqlBody
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
	return b
}

// gqlOK writes a GraphQL success envelope wrapping data.
func gqlOK(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"data": data})
}

// gqlErr writes a GraphQL error envelope.
func gqlErr(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"errors": []map[string]any{{"message": msg}},
	})
}

// --- base client behaviour tests ---

func TestClient_AuthHeader(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("Dagster-Cloud-Api-Token")
		gqlOK(w, map[string]any{"deployments": []any{}})
	}))
	defer srv.Close()

	if _, err := newClient(srv).ListDeployments(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "test-token" {
		t.Errorf("auth header = %q, want %q", got, "test-token")
	}
}

func TestClient_OrgLevelEndpoint(t *testing.T) {
	var path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		gqlOK(w, map[string]any{"deployments": []any{}})
	}))
	defer srv.Close()

	if _, err := newClient(srv).ListDeployments(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "/graphql" {
		t.Errorf("path = %q, want /graphql", path)
	}
}

func TestClient_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer srv.Close()

	_, err := newClient(srv).ListDeployments(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("error should mention status code, got: %v", err)
	}
}

func TestClient_GraphQLError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlErr(w, "permission denied")
	}))
	defer srv.Close()

	_, err := newClient(srv).ListDeployments(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("error should include GQL message, got: %v", err)
	}
}

func TestClient_TrailingSlashBaseURL(t *testing.T) {
	var path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		gqlOK(w, map[string]any{"deployments": []any{}})
	}))
	defer srv.Close()

	c := client.New("test-org", "test-token", srv.URL+"/")
	if _, err := c.ListDeployments(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "/graphql" {
		t.Errorf("path = %q, want /graphql", path)
	}
}
