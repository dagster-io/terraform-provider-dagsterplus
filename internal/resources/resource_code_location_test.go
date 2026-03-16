package resources

import (
	"testing"

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
	}

	input := modelToInput(plan)

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
	}

	input := modelToInput(plan)

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
	}

	input := modelToInput(plan)

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
	}

	input := modelToInput(plan)

	if input.WorkingDirectory != "" {
		t.Errorf("expected empty WorkingDirectory, got %q", input.WorkingDirectory)
	}
	if input.ExecutablePath != "" {
		t.Errorf("expected empty ExecutablePath, got %q", input.ExecutablePath)
	}
}
