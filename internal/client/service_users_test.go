package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func mockServiceUserResponse() map[string]any {
	return map[string]any{
		"id": "su-abc",
		"serviceUser": map[string]any{
			"id":          "su-abc",
			"name":        "my-service-user",
			"description": "A test service user",
		},
	}
}

func TestListServiceUsers_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "ListServiceUsers") {
			t.Errorf("expected ListServiceUsers query, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"serviceUsersOrError": map[string]any{
				"__typename":   "ServiceUsersWithScopedPermissionGrants",
				"serviceUsers": []any{mockServiceUserResponse()},
			},
		})
	}))
	defer srv.Close()

	users, err := newClient(srv).ListServiceUsers(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("len = %d, want 1", len(users))
	}
	if users[0].Name != "my-service-user" {
		t.Errorf("Name = %q, want my-service-user", users[0].Name)
	}
}

func TestListServiceUsers_UnexpectedType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"serviceUsersOrError": map[string]any{
				"__typename": "UnauthorizedError",
			},
		})
	}))
	defer srv.Close()

	_, err := newClient(srv).ListServiceUsers(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateServiceUser_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "CreateServiceUser") {
			t.Errorf("expected CreateServiceUser mutation, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"createServiceUser": map[string]any{
				"__typename": "ServiceUserWithScopedPermissionGrants",
				"id":         "su-abc",
				"serviceUser": map[string]any{
					"id":          "su-abc",
					"name":        "my-service-user",
					"description": "A test service user",
				},
			},
		})
	}))
	defer srv.Close()

	u, err := newClient(srv).CreateServiceUser(context.Background(), "my-service-user", "A test service user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.ID != "su-abc" {
		t.Errorf("ID = %q, want su-abc", u.ID)
	}
	if u.Name != "my-service-user" {
		t.Errorf("Name = %q, want my-service-user", u.Name)
	}
}

func TestCreateServiceUser_NameConflict(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"createServiceUser": map[string]any{
				"__typename": "ServiceUserNameConflictError",
				"message":    "name already in use",
			},
		})
	}))
	defer srv.Close()

	_, err := newClient(srv).CreateServiceUser(context.Background(), "my-service-user", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "name already in use") {
		t.Errorf("error should mention name conflict, got: %v", err)
	}
}

func TestDeleteServiceUser_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "DeleteServiceUser") {
			t.Errorf("expected DeleteServiceUser mutation, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"deleteServiceUser": map[string]any{
				"__typename": "DeleteServiceUserSuccess",
				"success":    true,
			},
		})
	}))
	defer srv.Close()

	if err := newClient(srv).DeleteServiceUser(context.Background(), "su-abc"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetServiceUser_Found(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "GetServiceUser") {
			t.Errorf("expected GetServiceUser query, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"serviceUser": map[string]any{
				"__typename": "ServiceUserWithScopedPermissionGrants",
				"id":         "su-abc",
				"serviceUser": map[string]any{
					"id":          "su-abc",
					"name":        "my-service-user",
					"description": "desc",
				},
			},
		})
	}))
	defer srv.Close()

	u, err := newClient(srv).GetServiceUser(context.Background(), "su-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.ID != "su-abc" {
		t.Errorf("ID = %q, want su-abc", u.ID)
	}
}

func TestGetServiceUser_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"serviceUser": map[string]any{
				"__typename": "ServiceUserNotFoundError",
				"message":    "not found",
			},
		})
	}))
	defer srv.Close()

	_, err := newClient(srv).GetServiceUser(context.Background(), "su-missing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention not found, got: %v", err)
	}
}

func TestCreateServiceToken_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "CreateServiceToken") {
			t.Errorf("expected CreateServiceToken mutation, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"createServiceToken": map[string]any{
				"__typename":  "ServiceToken",
				"id":          "tok-xyz",
				"token":       "secret-token-value",
				"description": "my token",
				"revoked":     false,
			},
		})
	}))
	defer srv.Close()

	tok, err := newClient(srv).CreateServiceToken(context.Background(), "su-abc", "my token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok.ID != "tok-xyz" {
		t.Errorf("ID = %q, want tok-xyz", tok.ID)
	}
	if tok.Token != "secret-token-value" {
		t.Errorf("Token = %q, want secret-token-value", tok.Token)
	}
}

func TestListServiceTokens_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"serviceTokensOrError": map[string]any{
				"__typename": "ServiceTokens",
				"tokens": []any{
					map[string]any{
						"id":          "tok-xyz",
						"token":       "",
						"description": "my token",
						"revoked":     false,
						"serviceUser": map[string]any{"id": "su-abc"},
					},
				},
			},
		})
	}))
	defer srv.Close()

	tokens, err := newClient(srv).ListServiceTokens(context.Background(), "su-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tokens) != 1 {
		t.Fatalf("len = %d, want 1", len(tokens))
	}
	if tokens[0].ID != "tok-xyz" {
		t.Errorf("ID = %q, want tok-xyz", tokens[0].ID)
	}
}

func TestListServiceTokens_UnexpectedType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"serviceTokensOrError": map[string]any{
				"__typename": "UnauthorizedError",
			},
		})
	}))
	defer srv.Close()

	_, err := newClient(srv).ListServiceTokens(context.Background(), "su-abc")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
