package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
)

func TestCreateTeam_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "CreateTeam") {
			t.Errorf("expected CreateTeam mutation, got: %s", b.Query)
		}
		if b.Variables["teamName"] != "data-engineering" {
			t.Errorf("teamName = %v, want data-engineering", b.Variables["teamName"])
		}
		gqlOK(w, map[string]any{
			"createOrganizationMemberTeam": map[string]any{
				"team": map[string]any{"teamId": "team-abc", "teamName": "data-engineering"},
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

func TestCreateTeam_NilTeam(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"createOrganizationMemberTeam": map[string]any{"team": nil},
		})
	}))
	defer srv.Close()

	_, err := newClient(srv).CreateTeam(context.Background(), "data-engineering")
	if err == nil {
		t.Fatal("expected error for nil team, got nil")
	}
}

func TestListTeams_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "ListTeams") {
			t.Errorf("expected ListTeams query, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"organizationMemberTeams": []any{
				map[string]any{"teamId": "team-abc", "teamName": "data-engineering"},
				map[string]any{"teamId": "team-def", "teamName": "platform"},
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
			"organizationMemberTeams": []any{
				map[string]any{"teamId": "team-abc", "teamName": "data-engineering"},
				map[string]any{"teamId": "team-def", "teamName": "platform"},
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
		gqlOK(w, map[string]any{"organizationMemberTeams": []any{}})
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

func TestGetTeamPermissions_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "GetTeamPermissions") {
			t.Errorf("expected GetTeamPermissions query, got: %s", b.Query)
		}
		if b.Variables["teamId"] != "team-abc" {
			t.Errorf("teamId = %v, want team-abc", b.Variables["teamId"])
		}
		gqlOK(w, map[string]any{
			"organizationMemberTeam": map[string]any{
				"deploymentPermissions": []any{
					map[string]any{"deploymentName": "prod", "deploymentRole": "EDITOR"},
					map[string]any{"deploymentName": "staging", "deploymentRole": "ADMIN"},
				},
			},
		})
	}))
	defer srv.Close()

	perms, err := newClient(srv).GetTeamPermissions(context.Background(), "team-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(perms) != 2 {
		t.Fatalf("len = %d, want 2", len(perms))
	}
	if perms[0].DeploymentName != "prod" {
		t.Errorf("perms[0].DeploymentName = %q, want prod", perms[0].DeploymentName)
	}
	if perms[0].Role != "EDITOR" {
		t.Errorf("perms[0].Role = %q, want EDITOR", perms[0].Role)
	}
}

func TestUpdateTeamPermissions_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "UpdateTeamPermissions") {
			t.Errorf("expected UpdateTeamPermissions mutation, got: %s", b.Query)
		}
		if b.Variables["teamId"] != "team-abc" {
			t.Errorf("teamId = %v, want team-abc", b.Variables["teamId"])
		}
		gqlOK(w, map[string]any{
			"updateOrganizationMemberTeamPermissions": map[string]any{"success": true},
		})
	}))
	defer srv.Close()

	perms := []client.Permission{
		{DeploymentName: "prod", Role: "EDITOR"},
	}
	if err := newClient(srv).UpdateTeamPermissions(context.Background(), "team-abc", perms); err != nil {
		t.Fatalf("unexpected error: %v", err)
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
			"deleteOrganizationMemberTeam": map[string]any{"success": true},
		})
	}))
	defer srv.Close()

	if err := newClient(srv).DeleteTeam(context.Background(), "team-abc"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
