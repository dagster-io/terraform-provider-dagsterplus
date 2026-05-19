package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func mockUserWithGrants(id int, email, name, role string) map[string]any {
	return map[string]any{
		"user": map[string]any{
			"userId": id,
			"email":  email,
			"name":   name,
		},
		"organizationPermissionGrant": map[string]any{
			"grant": role,
		},
	}
}

func TestAddUser_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "AddUser") {
			t.Errorf("expected AddUser mutation, got: %s", b.Query)
		}
		if b.Variables["email"] != "alice@example.com" {
			t.Errorf("email = %v, want alice@example.com", b.Variables["email"])
		}
		gqlOK(w, map[string]any{
			"addUserToOrganization": map[string]any{
				"__typename":     "AddUserToOrganizationSuccess",
				"userWithGrants": mockUserWithGrants(42, "alice@example.com", "Alice", "VIEWER"),
			},
		})
	}))
	defer srv.Close()

	user, err := newClient(srv).AddUser(context.Background(), "alice@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != "42" {
		t.Errorf("ID = %q, want 42", user.ID)
	}
	if user.Email != "alice@example.com" {
		t.Errorf("Email = %q, want alice@example.com", user.Email)
	}
	if user.Role != "VIEWER" {
		t.Errorf("Role = %q, want VIEWER", user.Role)
	}
}

func TestAddUser_UnexpectedTypename(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"addUserToOrganization": map[string]any{
				"__typename":     "AddUserToOrganizationFailure",
				"userWithGrants": nil,
			},
		})
	}))
	defer srv.Close()

	_, err := newClient(srv).AddUser(context.Background(), "alice@example.com")
	if err == nil {
		t.Fatal("expected error for unexpected typename, got nil")
	}
}

func TestListUsers_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "ListUsers") {
			t.Errorf("expected ListUsers query, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"usersOrError": map[string]any{
				"__typename": "DagsterCloudUsersWithScopedPermissionGrants",
				"users": []any{
					mockUserWithGrants(1, "alice@example.com", "Alice", "ADMIN"),
					mockUserWithGrants(2, "bob@example.com", "Bob", "VIEWER"),
				},
			},
		})
	}))
	defer srv.Close()

	users, err := newClient(srv).ListUsers(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("len(users) = %d, want 2", len(users))
	}
	if users[0].Email != "alice@example.com" {
		t.Errorf("users[0].Email = %q, want alice@example.com", users[0].Email)
	}
}

func TestListUsers_UnexpectedTypename(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"usersOrError": map[string]any{
				"__typename": "UnauthorizedError",
				"users":      []any{},
			},
		})
	}))
	defer srv.Close()

	_, err := newClient(srv).ListUsers(context.Background())
	if err == nil {
		t.Fatal("expected error for unexpected typename, got nil")
	}
}

func TestGetUser_Found(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"usersOrError": map[string]any{
				"__typename": "DagsterCloudUsersWithScopedPermissionGrants",
				"users": []any{
					mockUserWithGrants(99, "carol@example.com", "Carol", "EDITOR"),
				},
			},
		})
	}))
	defer srv.Close()

	user, err := newClient(srv).GetUser(context.Background(), "99")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "carol@example.com" {
		t.Errorf("Email = %q, want carol@example.com", user.Email)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"usersOrError": map[string]any{
				"__typename": "DagsterCloudUsersWithScopedPermissionGrants",
				"users":      []any{},
			},
		})
	}))
	defer srv.Close()

	_, err := newClient(srv).GetUser(context.Background(), "999")
	if err == nil {
		t.Fatal("expected error for not-found user, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention not found, got: %v", err)
	}
}

func TestUpdateUserRole_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "UpdateUserPermissions") {
			t.Errorf("expected UpdateUserPermissions mutation, got: %s", b.Query)
		}
		up, _ := b.Variables["userPermission"].(map[string]any)
		if up["email"] != "alice@example.com" {
			t.Errorf("email = %v, want alice@example.com", up["email"])
		}
		if up["grant"] != "EDITOR" {
			t.Errorf("grant = %v, want EDITOR", up["grant"])
		}
		gqlOK(w, map[string]any{
			"createOrUpdateUserPermissions": map[string]any{"__typename": "UnauthorizedError"},
		})
	}))
	defer srv.Close()

	err := newClient(srv).UpdateUserRole(context.Background(), "alice@example.com", "EDITOR")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateUserRole_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer srv.Close()

	err := newClient(srv).UpdateUserRole(context.Background(), "alice@example.com", "EDITOR")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("error should mention status code, got: %v", err)
	}
}

func TestRemoveUser_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "RemoveUser") {
			t.Errorf("expected RemoveUser mutation, got: %s", b.Query)
		}
		if b.Variables["email"] != "alice@example.com" {
			t.Errorf("email = %v, want alice@example.com", b.Variables["email"])
		}
		gqlOK(w, map[string]any{
			"removeUserFromOrganization": map[string]any{
				"__typename": "RemoveUserFromOrganizationSuccess",
				"email":      "alice@example.com",
			},
		})
	}))
	defer srv.Close()

	err := newClient(srv).RemoveUser(context.Background(), "alice@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
