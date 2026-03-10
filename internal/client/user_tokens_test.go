package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateUserToken_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "CreateUserToken") {
			t.Errorf("expected CreateUserToken mutation, got: %s", b.Query)
		}
		if b.Variables["description"] != "my-user-token" {
			t.Errorf("description = %v, want my-user-token", b.Variables["description"])
		}
		gqlOK(w, map[string]any{
			"createUserToken": map[string]any{
				"__typename":  "DagsterCloudUserToken",
				"id":          456,
				"description": "my-user-token",
				"token":       "user-secret-value",
				"user":        map[string]any{"userId": 789},
			},
		})
	}))
	defer srv.Close()

	tok, err := newClient(srv).CreateUserToken(context.Background(), "my-user-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok.ID != "456" {
		t.Errorf("ID = %q, want 456", tok.ID)
	}
	if tok.UserID != "789" {
		t.Errorf("UserID = %q, want 789", tok.UserID)
	}
	if tok.Name != "my-user-token" {
		t.Errorf("Name = %q, want my-user-token", tok.Name)
	}
	if tok.Token != "user-secret-value" {
		t.Errorf("Token = %q, want user-secret-value", tok.Token)
	}
}

func TestCreateUserToken_UnexpectedTypename(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"createUserToken": map[string]any{
				"__typename": "CreateUserTokenFailure",
			},
		})
	}))
	defer srv.Close()

	_, err := newClient(srv).CreateUserToken(context.Background(), "my-user-token")
	if err == nil {
		t.Fatal("expected error for unexpected typename, got nil")
	}
}

func TestListUserTokens_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "ListUserTokens") {
			t.Errorf("expected ListUserTokens query, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"userTokensOrError": map[string]any{
				"__typename": "DagsterCloudUserTokens",
				"tokens": []any{
					map[string]any{"id": 1, "description": "token-one", "token": "", "user": map[string]any{"userId": 42}},
					map[string]any{"id": 2, "description": "token-two", "token": "", "user": map[string]any{"userId": 42}},
				},
			},
		})
	}))
	defer srv.Close()

	tokens, err := newClient(srv).ListUserTokens(context.Background(), "42")
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

func TestListUserTokens_Empty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"userTokensOrError": map[string]any{
				"__typename": "DagsterCloudUserTokens",
				"tokens":     []any{},
			},
		})
	}))
	defer srv.Close()

	tokens, err := newClient(srv).ListUserTokens(context.Background(), "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tokens) != 0 {
		t.Errorf("len(tokens) = %d, want 0", len(tokens))
	}
}

func TestGetUserToken_Found(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"userTokensOrError": map[string]any{
				"__typename": "DagsterCloudUserTokens",
				"tokens": []any{
					map[string]any{"id": 99, "description": "my-user-token", "token": "", "user": map[string]any{"userId": 42}},
				},
			},
		})
	}))
	defer srv.Close()

	tok, err := newClient(srv).GetUserToken(context.Background(), "99", "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok.Name != "my-user-token" {
		t.Errorf("Name = %q, want my-user-token", tok.Name)
	}
}

func TestGetUserToken_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"userTokensOrError": map[string]any{
				"__typename": "DagsterCloudUserTokens",
				"tokens":     []any{},
			},
		})
	}))
	defer srv.Close()

	_, err := newClient(srv).GetUserToken(context.Background(), "99", "42")
	if err == nil {
		t.Fatal("expected error for not-found token, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention not found, got: %v", err)
	}
}

func TestDeleteUserToken_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "RevokeUserToken") {
			t.Errorf("expected RevokeUserToken mutation, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"revokeUserToken": map[string]any{
				"__typename":  "DagsterCloudUserToken",
				"id":          123,
				"description": "tok",
				"token":       "",
				"user":        map[string]any{"userId": 42},
			},
		})
	}))
	defer srv.Close()

	err := newClient(srv).DeleteUserToken(context.Background(), "123", "456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
