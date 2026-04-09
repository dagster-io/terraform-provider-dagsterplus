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

// serializedMeta returns a serialized JSON string for code location metadata
// using snake_case keys (matching the dagster serdes format).
func serializedMeta(image, pythonFile, workDir, execPath string) string {
	b, _ := json.Marshal(map[string]any{
		"image":             image,
		"python_file":       pythonFile,
		"working_directory": workDir,
		"executable_path":   execPath,
	})
	return string(b)
}

func serializedMetaWithContainerContext(image, pythonFile string, containerContext map[string]any) string {
	m := map[string]any{
		"image":       image,
		"python_file": pythonFile,
	}
	if containerContext != nil {
		m["container_context"] = containerContext
	}
	b, _ := json.Marshal(m)
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
		// Verify the document uses snake_case format.
		doc, ok := b.Variables["document"].(map[string]any)
		if !ok {
			t.Fatal("expected document variable to be a map")
		}
		if doc["location_name"] != "my-pipeline" {
			t.Errorf("location_name = %v, want my-pipeline", doc["location_name"])
		}
		cs, ok := doc["code_source"].(map[string]any)
		if !ok {
			t.Fatal("expected code_source to be a map")
		}
		if cs["python_file"] != "repo.py" {
			t.Errorf("code_source.python_file = %v, want repo.py", cs["python_file"])
		}

		gqlOK(w, map[string]any{
			"addOrUpdateLocationFromDocument": map[string]any{
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

func TestAddCodeLocation_WithContainerContext(t *testing.T) {
	containerCtx := map[string]any{
		"k8s": map[string]any{
			"namespace": "dagster",
			"env_vars":  []any{"FOO=bar"},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		doc, ok := b.Variables["document"].(map[string]any)
		if !ok {
			t.Fatal("expected document variable to be a map")
		}
		cc, ok := doc["container_context"].(map[string]any)
		if !ok {
			t.Fatal("expected container_context to be a map")
		}
		k8s, ok := cc["k8s"].(map[string]any)
		if !ok {
			t.Fatal("expected k8s to be a map")
		}
		if k8s["namespace"] != "dagster" {
			t.Errorf("k8s.namespace = %v, want dagster", k8s["namespace"])
		}

		gqlOK(w, map[string]any{
			"addOrUpdateLocationFromDocument": map[string]any{
				"__typename":                   "WorkspaceEntry",
				"locationName":                 "my-pipeline",
				"serializedDeploymentMetadata": serializedMetaWithContainerContext("my-image:latest", "repo.py", containerCtx),
			},
		})
	}))
	defer srv.Close()

	input := client.CodeLocationInput{
		Name:             "my-pipeline",
		Image:            "my-image:latest",
		CodeSource:       client.CodeSource{PythonFile: "repo.py"},
		ContainerContext: containerCtx,
	}

	cl, err := newClient(srv).AddCodeLocation(context.Background(), "prod", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cl.ContainerContext == nil {
		t.Fatal("expected ContainerContext to be non-nil")
	}
	k8s, ok := cl.ContainerContext["k8s"].(map[string]any)
	if !ok {
		t.Fatal("expected k8s to be a map")
	}
	if k8s["namespace"] != "dagster" {
		t.Errorf("k8s.namespace = %v, want dagster", k8s["namespace"])
	}
}

func TestAddCodeLocation_ErrorResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"addOrUpdateLocationFromDocument": map[string]any{
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
	if locs[0].Image != "my-registry/my-image:latest" {
		t.Errorf("Image = %q, want my-registry/my-image:latest", locs[0].Image)
	}
	if locs[0].WorkingDirectory != "/app" {
		t.Errorf("WorkingDirectory = %q, want /app", locs[0].WorkingDirectory)
	}
}

func TestListCodeLocations_WithContainerContext(t *testing.T) {
	containerCtx := map[string]any{
		"ecs": map[string]any{
			"env_vars":         []any{"DB_HOST=prod"},
			"server_resources": map[string]any{"cpu": "256", "memory": "512"},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, workspaceResponse([]any{
			map[string]any{
				"locationName":                 "my-pipeline",
				"serializedDeploymentMetadata": serializedMetaWithContainerContext("my-image:latest", "repo.py", containerCtx),
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
	if locs[0].ContainerContext == nil {
		t.Fatal("expected ContainerContext to be non-nil")
	}
	ecs, ok := locs[0].ContainerContext["ecs"].(map[string]any)
	if !ok {
		t.Fatal("expected ecs to be a map")
	}
	sr, ok := ecs["server_resources"].(map[string]any)
	if !ok {
		t.Fatal("expected server_resources to be a map")
	}
	if sr["cpu"] != "256" {
		t.Errorf("ecs.server_resources.cpu = %v, want 256", sr["cpu"])
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
			"addOrUpdateLocationFromDocument": map[string]any{
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

func TestAddCodeLocation_WithGit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		doc, ok := b.Variables["document"].(map[string]any)
		if !ok {
			t.Fatal("expected document variable to be a map")
		}
		git, ok := doc["git"].(map[string]any)
		if !ok {
			t.Fatal("expected git to be a map")
		}
		if git["commit_hash"] != "abc123" {
			t.Errorf("git.commit_hash = %v, want abc123", git["commit_hash"])
		}
		if git["url"] != "https://github.com/example/repo" {
			t.Errorf("git.url = %v, want https://github.com/example/repo", git["url"])
		}
		// Verify image is not set when using git.
		if _, hasImage := doc["image"]; hasImage && doc["image"] != "" {
			t.Errorf("image should be empty when using git, got: %v", doc["image"])
		}

		gqlOK(w, map[string]any{
			"addOrUpdateLocationFromDocument": map[string]any{
				"__typename":   "WorkspaceEntry",
				"locationName": "my-pipeline",
				"serializedDeploymentMetadata": func() string {
					b, _ := json.Marshal(map[string]any{
						"git_metadata": map[string]any{
							"commit_hash": "abc123",
							"url":         "https://github.com/example/repo",
						},
						"python_file": "repo.py",
					})
					return string(b)
				}(),
			},
		})
	}))
	defer srv.Close()

	input := client.CodeLocationInput{
		Name:       "my-pipeline",
		CommitHash: "abc123",
		URL:        "https://github.com/example/repo",
		CodeSource: client.CodeSource{PythonFile: "repo.py"},
	}
	cl, err := newClient(srv).AddCodeLocation(context.Background(), "prod", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cl.CommitHash != "abc123" {
		t.Errorf("CommitHash = %q, want abc123", cl.CommitHash)
	}
}
