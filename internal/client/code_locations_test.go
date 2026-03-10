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

// serializedMeta returns a serialized JSON string for code location metadata.
func serializedMeta(image, pythonFile, workDir, execPath string) string {
	b, _ := json.Marshal(map[string]any{
		"image":            image,
		"codeSource":       map[string]any{"pythonFile": pythonFile},
		"workingDirectory": workDir,
		"executablePath":   execPath,
	})
	return string(b)
}

func workspaceResponse(entries []any) map[string]any {
	return map[string]any{
		"workspace": map[string]any{
			"workspaceEntries": entries,
		},
	}
}

func TestAddCodeLocation_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "AddOrUpdateCodeLocation") {
			t.Errorf("expected AddOrUpdateCodeLocation mutation, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"addOrUpdateLocation": map[string]any{
				"__typename":                   "WorkspaceEntry",
				"locationName":                 "my-pipeline",
				"serializedDeploymentMetadata": serializedMeta("my-registry/my-image:latest", "repo.py", "/app", "/usr/bin/python3"),
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

func TestAddCodeLocation_ErrorResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"addOrUpdateLocation": map[string]any{
				"__typename": "InvalidLocationError",
			},
		})
	}))
	defer srv.Close()

	_, err := newClient(srv).AddCodeLocation(context.Background(), "prod", client.CodeLocationInput{Name: "x"})
	if err == nil {
		t.Fatal("expected error for invalid location response, got nil")
	}
}

func TestListCodeLocations_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "ListCodeLocations") {
			t.Errorf("expected ListCodeLocations query, got: %s", b.Query)
		}
		gqlOK(w, workspaceResponse([]any{
			map[string]any{
				"locationName":                 "my-pipeline",
				"serializedDeploymentMetadata": serializedMeta("my-registry/my-image:latest", "repo.py", "/app", ""),
			},
		}))
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
		gqlOK(w, workspaceResponse([]any{
			map[string]any{
				"locationName":                 "my-pipeline",
				"serializedDeploymentMetadata": serializedMeta("my-registry/my-image:latest", "repo.py", "", ""),
			},
			map[string]any{
				"locationName":                 "other",
				"serializedDeploymentMetadata": serializedMeta("other:latest", "", "", ""),
			},
		}))
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
		gqlOK(w, workspaceResponse([]any{}))
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
		if b.Variables["name"] != "my-pipeline" {
			t.Errorf("name = %v, want my-pipeline", b.Variables["name"])
		}
		gqlOK(w, map[string]any{
			"deleteLocation": map[string]any{"__typename": "DeleteLocationSuccess"},
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
			"addOrUpdateLocation": map[string]any{
				"__typename":                   "WorkspaceEntry",
				"locationName":                 "my-pipeline",
				"serializedDeploymentMetadata": serializedMeta("my-registry/my-image:v2", "", "", ""),
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
