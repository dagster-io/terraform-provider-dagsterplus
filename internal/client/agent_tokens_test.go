package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func agentTokensResponse(tokens []any) map[string]any {
	return map[string]any{
		"agentTokensOrError": map[string]any{
			"__typename": "DagsterCloudAgentTokens",
			"tokens":     tokens,
		},
	}
}

func TestCreateAgentToken_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "CreateAgentToken") {
			t.Errorf("expected CreateAgentToken mutation, got: %s", b.Query)
		}
		if b.Variables["description"] != "my-token" {
			t.Errorf("description = %v, want my-token", b.Variables["description"])
		}
		gqlOK(w, map[string]any{
			"createAgentToken": map[string]any{
				"__typename":  "DagsterCloudAgentToken",
				"id":          123,
				"description": "my-token",
				"token":       "secret-value",
			},
		})
	}))
	defer srv.Close()

	tok, err := newClient(srv).CreateAgentToken(context.Background(), "my-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok.ID != "123" {
		t.Errorf("ID = %q, want 123", tok.ID)
	}
	if tok.Name != "my-token" {
		t.Errorf("Name = %q, want my-token", tok.Name)
	}
	if tok.Token != "secret-value" {
		t.Errorf("Token = %q, want secret-value", tok.Token)
	}
}

func TestCreateAgentToken_UnexpectedTypename(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"createAgentToken": map[string]any{
				"__typename": "DagsterCloudTokenNotFoundError",
				"token":      "",
			},
		})
	}))
	defer srv.Close()

	_, err := newClient(srv).CreateAgentToken(context.Background(), "my-token")
	if err == nil {
		t.Fatal("expected error for empty token, got nil")
	}
}

func TestListAgentTokens_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "ListAgentTokens") {
			t.Errorf("expected ListAgentTokens query, got: %s", b.Query)
		}
		gqlOK(w, agentTokensResponse([]any{
			map[string]any{"id": 1, "description": "token-one", "token": ""},
			map[string]any{"id": 2, "description": "token-two", "token": ""},
		}))
	}))
	defer srv.Close()

	tokens, err := newClient(srv).ListAgentTokens(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tokens) != 2 {
		t.Fatalf("len(tokens) = %d, want 2", len(tokens))
	}
	if tokens[0].ID != "1" {
		t.Errorf("tokens[0].ID = %q, want 1", tokens[0].ID)
	}
}

func TestListAgentTokens_Empty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, agentTokensResponse([]any{}))
	}))
	defer srv.Close()

	tokens, err := newClient(srv).ListAgentTokens(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tokens) != 0 {
		t.Errorf("len(tokens) = %d, want 0", len(tokens))
	}
}

func TestGetAgentToken_Found(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, agentTokensResponse([]any{
			map[string]any{"id": 99, "description": "my-token", "token": ""},
		}))
	}))
	defer srv.Close()

	tok, err := newClient(srv).GetAgentToken(context.Background(), "99")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok.Name != "my-token" {
		t.Errorf("Name = %q, want my-token", tok.Name)
	}
}

func TestGetAgentToken_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, agentTokensResponse([]any{}))
	}))
	defer srv.Close()

	_, err := newClient(srv).GetAgentToken(context.Background(), "tok-missing")
	if err == nil {
		t.Fatal("expected error for not-found token, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention not found, got: %v", err)
	}
}

func TestDeleteAgentToken_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "RevokeAgentToken") {
			t.Errorf("expected RevokeAgentToken mutation, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"revokeAgentToken": map[string]any{
				"__typename":  "DagsterCloudAgentToken",
				"id":          123,
				"description": "my-token",
				"token":       "",
			},
		})
	}))
	defer srv.Close()

	err := newClient(srv).DeleteAgentToken(context.Background(), "123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
