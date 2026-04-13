package resources

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestModelToInput_PythonFile(t *testing.T) {
	plan := codeLocationResourceModel{
		Name:             types.StringValue("my-location"),
		Image:            types.StringValue("my-org/my-image:latest"),
		WorkingDirectory: types.StringValue("/app"),
		ExecutablePath:   types.StringValue("/usr/bin/python3"),
		CodeSource: CodeSourceModel{
			PythonFile:  types.StringValue("repo.py"),
			PackageName: types.StringValue(""),
			ModuleName:  types.StringValue(""),
		},
		ContainerContext: jsontypes.NewNormalizedNull(),
	}

	input, err := modelToInput(plan)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if input.Name != "my-location" {
		t.Errorf("expected Name my-location, got %q", input.Name)
	}
	if input.Image != "my-org/my-image:latest" {
		t.Errorf("expected Image my-org/my-image:latest, got %q", input.Image)
	}
	if input.WorkingDirectory != "/app" {
		t.Errorf("expected WorkingDirectory /app, got %q", input.WorkingDirectory)
	}
	if input.ExecutablePath != "/usr/bin/python3" {
		t.Errorf("expected ExecutablePath /usr/bin/python3, got %q", input.ExecutablePath)
	}
	if input.CodeSource.PythonFile != "repo.py" {
		t.Errorf("expected CodeSource.PythonFile repo.py, got %q", input.CodeSource.PythonFile)
	}
	if input.CodeSource.PackageName != "" {
		t.Errorf("expected empty PackageName, got %q", input.CodeSource.PackageName)
	}
	if input.CodeSource.ModuleName != "" {
		t.Errorf("expected empty ModuleName, got %q", input.CodeSource.ModuleName)
	}
}

func TestModelToInput_PackageName(t *testing.T) {
	plan := codeLocationResourceModel{
		Name:             types.StringValue("pkg-location"),
		Image:            types.StringValue("my-image:1.0"),
		WorkingDirectory: types.StringValue(""),
		ExecutablePath:   types.StringValue(""),
		CodeSource: CodeSourceModel{
			PythonFile:  types.StringValue(""),
			PackageName: types.StringValue("my_dagster_package"),
			ModuleName:  types.StringValue(""),
		},
		ContainerContext: jsontypes.NewNormalizedNull(),
	}

	input, err := modelToInput(plan)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if input.CodeSource.PackageName != "my_dagster_package" {
		t.Errorf("expected PackageName my_dagster_package, got %q", input.CodeSource.PackageName)
	}
	if input.CodeSource.PythonFile != "" {
		t.Errorf("expected empty PythonFile, got %q", input.CodeSource.PythonFile)
	}
	if input.CodeSource.ModuleName != "" {
		t.Errorf("expected empty ModuleName, got %q", input.CodeSource.ModuleName)
	}
}

func TestModelToInput_ModuleName(t *testing.T) {
	plan := codeLocationResourceModel{
		Name:  types.StringValue("mod-location"),
		Image: types.StringValue("mod-image:2.0"),
		CodeSource: CodeSourceModel{
			PythonFile:  types.StringValue(""),
			PackageName: types.StringValue(""),
			ModuleName:  types.StringValue("my.dagster.module"),
		},
		WorkingDirectory: types.StringValue(""),
		ExecutablePath:   types.StringValue(""),
		ContainerContext: jsontypes.NewNormalizedNull(),
	}

	input, err := modelToInput(plan)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if input.CodeSource.ModuleName != "my.dagster.module" {
		t.Errorf("expected ModuleName my.dagster.module, got %q", input.CodeSource.ModuleName)
	}
}

func TestModelToInput_OptionalFieldsEmpty(t *testing.T) {
	plan := codeLocationResourceModel{
		Name:             types.StringValue("bare-location"),
		Image:            types.StringValue("bare-image:latest"),
		WorkingDirectory: types.StringValue(""),
		ExecutablePath:   types.StringValue(""),
		CodeSource: CodeSourceModel{
			PythonFile:  types.StringValue("bare.py"),
			PackageName: types.StringValue(""),
			ModuleName:  types.StringValue(""),
		},
		ContainerContext: jsontypes.NewNormalizedNull(),
	}

	input, err := modelToInput(plan)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if input.WorkingDirectory != "" {
		t.Errorf("expected empty WorkingDirectory, got %q", input.WorkingDirectory)
	}
	if input.ExecutablePath != "" {
		t.Errorf("expected empty ExecutablePath, got %q", input.ExecutablePath)
	}
}

func TestModelToInput_WithContainerContext(t *testing.T) {
	plan := codeLocationResourceModel{
		Name:  types.StringValue("my-location"),
		Image: types.StringValue("my-image:latest"),
		CodeSource: CodeSourceModel{
			PythonFile:  types.StringValue("repo.py"),
			PackageName: types.StringValue(""),
			ModuleName:  types.StringValue(""),
		},
		WorkingDirectory: types.StringValue(""),
		ExecutablePath:   types.StringValue(""),
		ContainerContext: jsontypes.NewNormalizedValue(`{"k8s":{"namespace":"dagster","env_vars":["FOO=bar"]}}`),
	}

	input, err := modelToInput(plan)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if input.ContainerContext == nil {
		t.Fatal("expected ContainerContext to be non-nil")
	}
	k8s, ok := input.ContainerContext["k8s"].(map[string]any)
	if !ok {
		t.Fatal("expected k8s to be a map")
	}
	if k8s["namespace"] != "dagster" {
		t.Errorf("expected namespace dagster, got %v", k8s["namespace"])
	}
}

func TestModelToInput_ContainerContextInvalidJSON(t *testing.T) {
	plan := codeLocationResourceModel{
		Name:  types.StringValue("my-location"),
		Image: types.StringValue("my-image:latest"),
		CodeSource: CodeSourceModel{
			PythonFile: types.StringValue("repo.py"),
		},
		ContainerContext: jsontypes.NewNormalizedValue(`{not valid json}`),
	}

	_, err := modelToInput(plan)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}
