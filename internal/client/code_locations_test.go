package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
)

func mockCodeLocationResponse() map[string]any {
	return map[string]any{
		"locationName":     "my-pipeline",
		"image":            "my-registry/my-image:latest",
		"codeSource":       map[string]any{"pythonFile": "repo.py"},
		"workingDirectory": "/app",
		"executablePath":   "/usr/bin/python3",
	}
}

func TestAddCodeLocation_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "AddCodeLocation") {
			t.Errorf("expected AddCodeLocation mutation, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"addOrUpdateCodeLocation": map[string]any{
				"codeLocationDeployData": mockCodeLocationResponse(),
			},
		})
	}))
	defer srv.Close()

	input := client.CodeLocationInput{
		Name:             "my-pipeline",
		Image:            "my-registry/my-image:latest",
		CodeSource:       client.CodeSource{PythonFile: "repo.py"},
		WorkingDirectory: "/app",
		ExecutablePath:   "/usr/bin/python3",
	}

	cl, err := newClient(srv).AddCodeLocation(context.Background(), "prod", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cl.Name != "my-pipeline" {
		t.Errorf("Name = %q, want my-pipeline", cl.Name)
	}
	if cl.Image != "my-registry/my-image:latest" {
		t.Errorf("Image = %q, unexpected", cl.Image)
	}
	if cl.ID != "prod/my-pipeline" {
		t.Errorf("ID = %q, want prod/my-pipeline", cl.ID)
	}
	if cl.CodeSource.PythonFile != "repo.py" {
		t.Errorf("CodeSource.PythonFile = %q, want repo.py", cl.CodeSource.PythonFile)
	}
}

func TestAddCodeLocation_NilResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"addOrUpdateCodeLocation": map[string]any{"codeLocationDeployData": nil},
		})
	}))
	defer srv.Close()

	_, err := newClient(srv).AddCodeLocation(context.Background(), "prod", client.CodeLocationInput{Name: "x"})
	if err == nil {
		t.Fatal("expected error for nil response, got nil")
	}
}

func TestListCodeLocations_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "ListCodeLocations") {
			t.Errorf("expected ListCodeLocations query, got: %s", b.Query)
		}
		if b.Variables["deploymentName"] != "prod" {
			t.Errorf("deploymentName = %v, want prod", b.Variables["deploymentName"])
		}
		gqlOK(w, map[string]any{
			"codeLocations": []any{mockCodeLocationResponse()},
		})
	}))
	defer srv.Close()

	locs, err := newClient(srv).ListCodeLocations(context.Background(), "prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(locs) != 1 {
		t.Fatalf("len = %d, want 1", len(locs))
	}
	if locs[0].Name != "my-pipeline" {
		t.Errorf("Name = %q, want my-pipeline", locs[0].Name)
	}
	if locs[0].ID != "prod/my-pipeline" {
		t.Errorf("ID = %q, want prod/my-pipeline", locs[0].ID)
	}
}

func TestGetCodeLocation_Found(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"codeLocations": []any{
				mockCodeLocationResponse(),
				map[string]any{
					"locationName": "other", "image": "other:latest",
					"codeSource": map[string]any{}, "workingDirectory": "", "executablePath": "",
				},
			},
		})
	}))
	defer srv.Close()

	cl, err := newClient(srv).GetCodeLocation(context.Background(), "prod", "my-pipeline")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cl.Name != "my-pipeline" {
		t.Errorf("Name = %q, want my-pipeline", cl.Name)
	}
}

func TestGetCodeLocation_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{"codeLocations": []any{}})
	}))
	defer srv.Close()

	_, err := newClient(srv).GetCodeLocation(context.Background(), "prod", "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention not found, got: %v", err)
	}
}

func TestDeleteCodeLocation_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "DeleteCodeLocation") {
			t.Errorf("expected DeleteCodeLocation mutation, got: %s", b.Query)
		}
		if b.Variables["deploymentName"] != "prod" {
			t.Errorf("deploymentName = %v, want prod", b.Variables["deploymentName"])
		}
		if b.Variables["locationName"] != "my-pipeline" {
			t.Errorf("locationName = %v, want my-pipeline", b.Variables["locationName"])
		}
		gqlOK(w, map[string]any{
			"deleteCodeLocation": map[string]any{"success": true},
		})
	}))
	defer srv.Close()

	if err := newClient(srv).DeleteCodeLocation(context.Background(), "prod", "my-pipeline"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateCodeLocation_Success(t *testing.T) {
	// UpdateCodeLocation delegates to AddCodeLocation (upsert).
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"addOrUpdateCodeLocation": map[string]any{
				"codeLocationDeployData": mockCodeLocationResponse(),
			},
		})
	}))
	defer srv.Close()

	input := client.CodeLocationInput{
		Name:  "my-pipeline",
		Image: "my-registry/my-image:v2",
	}
	cl, err := newClient(srv).UpdateCodeLocation(context.Background(), "prod", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cl.Name != "my-pipeline" {
		t.Errorf("Name = %q, want my-pipeline", cl.Name)
	}
}
