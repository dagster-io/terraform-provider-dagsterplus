package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// CodeSource describes where the code lives within a code location.
type CodeSource struct {
	PythonFile  string `json:"pythonFile,omitempty"`
	PackageName string `json:"packageName,omitempty"`
	ModuleName  string `json:"moduleName,omitempty"`
}

// CodeLocation represents a Dagster+ code location.
type CodeLocation struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	Image            string     `json:"image"`
	CodeSource       CodeSource `json:"codeSource"`
	WorkingDirectory string     `json:"workingDirectory"`
	ExecutablePath   string     `json:"executablePath"`
}

// CodeLocationInput is the input for creating or updating a code location.
type CodeLocationInput struct {
	Name             string     `json:"name"`
	Image            string     `json:"image"`
	CodeSource       CodeSource `json:"codeSource"`
	WorkingDirectory string     `json:"workingDirectory,omitempty"`
	ExecutablePath   string     `json:"executablePath,omitempty"`
}

// AddCodeLocation adds a code location to a deployment.
func (c *Client) AddCodeLocation(ctx context.Context, deployment string, loc CodeLocationInput) (*CodeLocation, error) {
	const mutation = `
mutation AddCodeLocation($deploymentName: String!, $serializedCodeLocationDeployData: String!) {
  addOrUpdateCodeLocation(
    deploymentName: $deploymentName
    serializedCodeLocationDeployData: $serializedCodeLocationDeployData
  ) {
    codeLocationDeployData {
      locationName
      image
      codeSource {
        pythonFile
        packageName
        moduleName
      }
      workingDirectory
      executablePath
    }
  }
}`

	type codeSourceInput struct {
		PythonFile  *string `json:"pythonFile,omitempty"`
		PackageName *string `json:"packageName,omitempty"`
		ModuleName  *string `json:"moduleName,omitempty"`
	}

	type deployData struct {
		LocationName     string          `json:"locationName"`
		Image            string          `json:"image"`
		CodeSource       codeSourceInput `json:"codeSource"`
		WorkingDirectory string          `json:"workingDirectory,omitempty"`
		ExecutablePath   string          `json:"executablePath,omitempty"`
	}

	cs := codeSourceInput{}
	if loc.CodeSource.PythonFile != "" {
		cs.PythonFile = &loc.CodeSource.PythonFile
	}
	if loc.CodeSource.PackageName != "" {
		cs.PackageName = &loc.CodeSource.PackageName
	}
	if loc.CodeSource.ModuleName != "" {
		cs.ModuleName = &loc.CodeSource.ModuleName
	}

	data := deployData{
		LocationName:     loc.Name,
		Image:            loc.Image,
		CodeSource:       cs,
		WorkingDirectory: loc.WorkingDirectory,
		ExecutablePath:   loc.ExecutablePath,
	}

	import_json, err := marshalJSON(data)
	if err != nil {
		return nil, fmt.Errorf("AddCodeLocation: marshalling deploy data: %w", err)
	}

	var result struct {
		AddOrUpdateCodeLocation struct {
			CodeLocationDeployData *CodeLocation `json:"codeLocationDeployData"`
		} `json:"addOrUpdateCodeLocation"`
	}

	err = c.doGraphQL(ctx, "", mutation, map[string]any{
		"deploymentName":                  deployment,
		"serializedCodeLocationDeployData": import_json,
	}, &result)
	if err != nil {
		return nil, fmt.Errorf("AddCodeLocation: %w", err)
	}

	if result.AddOrUpdateCodeLocation.CodeLocationDeployData == nil {
		return nil, fmt.Errorf("AddCodeLocation: API returned nil code location")
	}

	cl := result.AddOrUpdateCodeLocation.CodeLocationDeployData
	cl.ID = fmt.Sprintf("%s/%s", deployment, loc.Name)
	cl.Name = loc.Name // API returns "locationName"; struct tag is "name" — set explicitly
	return cl, nil
}

// GetCodeLocation retrieves a code location from a deployment by name.
func (c *Client) GetCodeLocation(ctx context.Context, deployment, name string) (*CodeLocation, error) {
	locations, err := c.ListCodeLocations(ctx, deployment)
	if err != nil {
		return nil, err
	}
	for i := range locations {
		if locations[i].Name == name {
			return &locations[i], nil
		}
	}
	return nil, fmt.Errorf("GetCodeLocation: code location %q not found in deployment %q", name, deployment)
}

// ListCodeLocations returns all code locations for a deployment.
func (c *Client) ListCodeLocations(ctx context.Context, deployment string) ([]CodeLocation, error) {
	const query = `
query ListCodeLocations($deploymentName: String!) {
  codeLocations(deploymentName: $deploymentName) {
    locationName
    image
    codeSource {
      pythonFile
      packageName
      moduleName
    }
    workingDirectory
    executablePath
  }
}`

	var result struct {
		CodeLocations []struct {
			LocationName     string     `json:"locationName"`
			Image            string     `json:"image"`
			CodeSource       CodeSource `json:"codeSource"`
			WorkingDirectory string     `json:"workingDirectory"`
			ExecutablePath   string     `json:"executablePath"`
		} `json:"codeLocations"`
	}

	if err := c.doGraphQL(ctx, "", query, map[string]any{"deploymentName": deployment}, &result); err != nil {
		return nil, fmt.Errorf("ListCodeLocations: %w", err)
	}

	locations := make([]CodeLocation, len(result.CodeLocations))
	for i, l := range result.CodeLocations {
		locations[i] = CodeLocation{
			ID:               fmt.Sprintf("%s/%s", deployment, l.LocationName),
			Name:             l.LocationName,
			Image:            l.Image,
			CodeSource:       l.CodeSource,
			WorkingDirectory: l.WorkingDirectory,
			ExecutablePath:   l.ExecutablePath,
		}
	}

	return locations, nil
}

// UpdateCodeLocation updates an existing code location (same API as add).
func (c *Client) UpdateCodeLocation(ctx context.Context, deployment string, loc CodeLocationInput) (*CodeLocation, error) {
	return c.AddCodeLocation(ctx, deployment, loc)
}

// DeleteCodeLocation removes a code location from a deployment.
func (c *Client) DeleteCodeLocation(ctx context.Context, deployment, name string) error {
	const mutation = `
mutation DeleteCodeLocation($deploymentName: String!, $locationName: String!) {
  deleteCodeLocation(deploymentName: $deploymentName, locationName: $locationName) {
    success
  }
}`

	err := c.doGraphQL(ctx, "", mutation, map[string]any{
		"deploymentName": deployment,
		"locationName":   name,
	}, nil)
	if err != nil {
		return fmt.Errorf("DeleteCodeLocation: %w", err)
	}

	return nil
}

// marshalJSON serialises v to a JSON string.
func marshalJSON(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
