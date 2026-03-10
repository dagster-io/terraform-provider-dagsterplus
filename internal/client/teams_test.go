package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateTeam_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "CreateOrUpdateTeam") {
			t.Errorf("expected CreateOrUpdateTeam mutation, got: %s", b.Query)
		}
		if b.Variables["name"] != "data-engineering" {
			t.Errorf("name = %v, want data-engineering", b.Variables["name"])
		}
		gqlOK(w, map[string]any{
			"createOrUpdateTeam": map[string]any{
				"__typename": "CreateOrUpdateTeamSuccess",
				"team":       map[string]any{"id": "team-abc", "name": "data-engineering"},
			},
		})
	}))
	defer srv.Close()

	team, err := newClient(srv).CreateTeam(context.Background(), "data-engineering")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if team.ID != "team-abc" {
		t.Errorf("ID = %q, want team-abc", team.ID)
	}
	if team.Name != "data-engineering" {
		t.Errorf("Name = %q, want data-engineering", team.Name)
	}
}

func TestCreateTeam_UnexpectedTypename(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"createOrUpdateTeam": map[string]any{
				"__typename": "CreateOrUpdateTeamFailure",
				"team":       nil,
			},
		})
	}))
	defer srv.Close()

	_, err := newClient(srv).CreateTeam(context.Background(), "data-engineering")
	if err == nil {
		t.Fatal("expected error for unexpected typename, got nil")
	}
}

func TestListTeams_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "ListTeams") {
			t.Errorf("expected ListTeams query, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"teamPermissions": []any{
				map[string]any{"team": map[string]any{"id": "team-abc", "name": "data-engineering"}},
				map[string]any{"team": map[string]any{"id": "team-def", "name": "platform"}},
			},
		})
	}))
	defer srv.Close()

	teams, err := newClient(srv).ListTeams(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(teams) != 2 {
		t.Fatalf("len = %d, want 2", len(teams))
	}
	if teams[0].ID != "team-abc" {
		t.Errorf("teams[0].ID = %q, want team-abc", teams[0].ID)
	}
}

func TestGetTeam_Found(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"teamPermissions": []any{
				map[string]any{"team": map[string]any{"id": "team-abc", "name": "data-engineering"}},
				map[string]any{"team": map[string]any{"id": "team-def", "name": "platform"}},
			},
		})
	}))
	defer srv.Close()

	team, err := newClient(srv).GetTeam(context.Background(), "team-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if team.Name != "data-engineering" {
		t.Errorf("Name = %q, want data-engineering", team.Name)
	}
}

func TestGetTeam_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{"teamPermissions": []any{}})
	}))
	defer srv.Close()

	_, err := newClient(srv).GetTeam(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention not found, got: %v", err)
	}
}

func TestDeleteTeam_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "DeleteTeam") {
			t.Errorf("expected DeleteTeam mutation, got: %s", b.Query)
		}
		if b.Variables["teamId"] != "team-abc" {
			t.Errorf("teamId = %v, want team-abc", b.Variables["teamId"])
		}
		gqlOK(w, map[string]any{
			"deleteTeam": map[string]any{"__typename": "DeleteTeamSuccess"},
		})
	}))
	defer srv.Close()

	if err := newClient(srv).DeleteTeam(context.Background(), "team-abc"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
