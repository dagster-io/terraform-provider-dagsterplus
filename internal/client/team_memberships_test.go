package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAddUserToTeam_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "AddMemberToTeam") {
			t.Errorf("expected AddMemberToTeam mutation, got: %s", b.Query)
		}
		if b.Variables["teamId"] != "team-abc" {
			t.Errorf("teamId = %v, want team-abc", b.Variables["teamId"])
		}
		if b.Variables["memberId"] != float64(42) {
			t.Errorf("memberId = %v, want 42", b.Variables["memberId"])
		}
		gqlOK(w, map[string]any{
			"addMemberToTeam": map[string]any{"__typename": "AddMemberToTeamSuccess"},
		})
	}))
	defer srv.Close()

	if err := newClient(srv).AddUserToTeam(context.Background(), "team-abc", "42"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddUserToTeam_InvalidUserID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("expected no request to be made for invalid user ID")
		gqlOK(w, nil)
	}))
	defer srv.Close()

	err := newClient(srv).AddUserToTeam(context.Background(), "team-abc", "not-an-int")
	if err == nil {
		t.Fatal("expected error for invalid user ID, got nil")
	}
}

func TestAddUserToTeam_GraphQLError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlErr(w, "user not found")
	}))
	defer srv.Close()

	err := newClient(srv).AddUserToTeam(context.Background(), "team-abc", "42")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "user not found") {
		t.Errorf("error should include API message, got: %v", err)
	}
}

func TestRemoveUserFromTeam_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "RemoveMemberFromTeam") {
			t.Errorf("expected RemoveMemberFromTeam mutation, got: %s", b.Query)
		}
		if b.Variables["teamId"] != "team-abc" {
			t.Errorf("teamId = %v, want team-abc", b.Variables["teamId"])
		}
		if b.Variables["memberId"] != float64(99) {
			t.Errorf("memberId = %v, want 99", b.Variables["memberId"])
		}
		gqlOK(w, map[string]any{
			"removeMemberFromTeam": map[string]any{"__typename": "RemoveMemberFromTeamSuccess"},
		})
	}))
	defer srv.Close()

	if err := newClient(srv).RemoveUserFromTeam(context.Background(), "team-abc", "99"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRemoveUserFromTeam_InvalidUserID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("expected no request to be made for invalid user ID")
		gqlOK(w, nil)
	}))
	defer srv.Close()

	err := newClient(srv).RemoveUserFromTeam(context.Background(), "team-abc", "not-an-int")
	if err == nil {
		t.Fatal("expected error for invalid user ID, got nil")
	}
}

func TestIsUserInTeam_Found(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "ListTeamMembers") {
			t.Errorf("expected ListTeamMembers query, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"teamPermissions": []any{
				map[string]any{
					"team": map[string]any{
						"id":      "team-abc",
						"members": []any{map[string]any{"userId": 42}},
					},
				},
			},
		})
	}))
	defer srv.Close()

	found, err := newClient(srv).IsUserInTeam(context.Background(), "team-abc", "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Error("expected user to be found in team")
	}
}

func TestIsUserInTeam_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"teamPermissions": []any{
				map[string]any{
					"team": map[string]any{
						"id":      "team-abc",
						"members": []any{},
					},
				},
			},
		})
	}))
	defer srv.Close()

	found, err := newClient(srv).IsUserInTeam(context.Background(), "team-abc", "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("expected user not to be found in team")
	}
}

func TestIsUserInTeam_TeamNotInResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{"teamPermissions": []any{}})
	}))
	defer srv.Close()

	found, err := newClient(srv).IsUserInTeam(context.Background(), "team-missing", "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("expected false for team not in response, got true")
	}
}
