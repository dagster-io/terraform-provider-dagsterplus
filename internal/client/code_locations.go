package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client/schema"
)

// codeLocationDocument is the canonical JSON shape for a code location document.
type codeLocationDocument struct {
	LocationName     string     `json:"locationName"`
	Image            string     `json:"image,omitempty"`
	CodeSource       CodeSource `json:"codeSource"`
	WorkingDirectory string     `json:"workingDirectory,omitempty"`
	ExecutablePath   string     `json:"executablePath,omitempty"`
	Attribute        string     `json:"attribute,omitempty"`
	AgentQueue       string     `json:"agentQueue,omitempty"`
	CommitHash       string     `json:"commitHash,omitempty"`
	URL              string     `json:"url,omitempty"`
}

// ParseCodeLocationDocument parses a JSON document into a CodeLocationInput and returns the location name.
func ParseCodeLocationDocument(docJSON string) (CodeLocationInput, string, error) {
	var doc codeLocationDocument
	if err := json.Unmarshal([]byte(docJSON), &doc); err != nil {
		return CodeLocationInput{}, "", fmt.Errorf("ParseCodeLocationDocument: %w", err)
	}
	if doc.LocationName == "" {
		return CodeLocationInput{}, "", fmt.Errorf("ParseCodeLocationDocument: document must include 'locationName'")
	}
	input := CodeLocationInput{
		Name:             doc.LocationName,
		Image:            doc.Image,
		CodeSource:       doc.CodeSource,
		WorkingDirectory: doc.WorkingDirectory,
		ExecutablePath:   doc.ExecutablePath,
		Attribute:        doc.Attribute,
		AgentQueue:       doc.AgentQueue,
		CommitHash:       doc.CommitHash,
		URL:              doc.URL,
	}
	return input, doc.LocationName, nil
}

// CodeLocationToDocument serializes a CodeLocation to the canonical JSON document format.
func CodeLocationToDocument(loc *CodeLocation) (string, error) {
	doc := codeLocationDocument{
		LocationName:     loc.Name,
		Image:            loc.Image,
		CodeSource:       loc.CodeSource,
		WorkingDirectory: loc.WorkingDirectory,
		ExecutablePath:   loc.ExecutablePath,
		Attribute:        loc.Attribute,
		AgentQueue:       loc.AgentQueue,
		CommitHash:       loc.CommitHash,
		URL:              loc.URL,
	}
	b, err := json.Marshal(doc)
	if err != nil {
		return "", fmt.Errorf("CodeLocationToDocument: %w", err)
	}
	return string(b), nil
}

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
	Attribute        string     `json:"attribute"`
	AgentQueue       string     `json:"agentQueue"`
	CommitHash       string     `json:"commitHash"`
	URL              string     `json:"url"`
}

// CodeLocationInput is the input for creating or updating a code location.
type CodeLocationInput struct {
	Name             string     `json:"name"`
	Image            string     `json:"image"`
	CodeSource       CodeSource `json:"codeSource"`
	WorkingDirectory string     `json:"workingDirectory,omitempty"`
	ExecutablePath   string     `json:"executablePath,omitempty"`
	Attribute        string     `json:"attribute,omitempty"`
	AgentQueue       string     `json:"agentQueue,omitempty"`
	CommitHash       string     `json:"commitHash,omitempty"`
	URL              string     `json:"url,omitempty"`
}

// serializedMetadata mirrors the JSON shape of WorkspaceEntry.serializedDeploymentMetadata.
type serializedMetadata struct {
	LocationName     string     `json:"locationName"`
	Image            string     `json:"image"`
	CodeSource       CodeSource `json:"codeSource"`
	WorkingDirectory string     `json:"workingDirectory"`
	ExecutablePath   string     `json:"executablePath"`
	Attribute        string     `json:"attribute"`
	AgentQueue       string     `json:"agentQueue"`
	CommitHash       string     `json:"commitHash"`
	URL              string     `json:"url"`
}

// AddCodeLocation adds a code location to a deployment.
func (c *Client) AddCodeLocation(ctx context.Context, deployment string, loc CodeLocationInput) (*CodeLocation, error) {
	selector := schema.LocationSelector{
		Name:             loc.Name,
		Image:            loc.Image,
		PythonFile:       loc.CodeSource.PythonFile,
		PackageName:      loc.CodeSource.PackageName,
		ModuleName:       loc.CodeSource.ModuleName,
		WorkingDirectory: loc.WorkingDirectory,
		ExecutablePath:   loc.ExecutablePath,
		Attribute:        loc.Attribute,
		AgentQueue:       loc.AgentQueue,
		CommitHash:       loc.CommitHash,
		Url:              loc.URL,
	}

	resp, err := schema.AddOrUpdateCodeLocation(ctx, c.gqlClient(deployment), selector)
	if err != nil {
		return nil, fmt.Errorf("AddCodeLocation: %w", err)
	}

	switch resp.AddOrUpdateLocation.(type) {
	case *schema.AddOrUpdateCodeLocationAddOrUpdateLocationWorkspaceEntry:
		cl := &CodeLocation{
			ID:               fmt.Sprintf("%s/%s", deployment, loc.Name),
			Name:             loc.Name,
			Image:            loc.Image,
			CodeSource:       loc.CodeSource,
			WorkingDirectory: loc.WorkingDirectory,
			ExecutablePath:   loc.ExecutablePath,
			Attribute:        loc.Attribute,
			AgentQueue:       loc.AgentQueue,
			CommitHash:       loc.CommitHash,
			URL:              loc.URL,
		}
		return cl, nil
	default:
		return nil, fmt.Errorf("AddCodeLocation: unexpected result type %T", resp.AddOrUpdateLocation)
	}
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
	resp, err := schema.ListCodeLocations(ctx, c.gqlClient(deployment))
	if err != nil {
		return nil, fmt.Errorf("ListCodeLocations: %w", err)
	}

	entries := resp.Workspace.WorkspaceEntries
	locations := make([]CodeLocation, 0, len(entries))
	for _, e := range entries {
		var meta serializedMetadata
		if err := json.Unmarshal([]byte(e.WorkspaceEntryFields.SerializedDeploymentMetadata), &meta); err != nil {
			// Skip entries with unparseable metadata rather than failing the whole list.
			continue
		}
		locations = append(locations, CodeLocation{
			ID:               fmt.Sprintf("%s/%s", deployment, e.WorkspaceEntryFields.LocationName),
			Name:             e.WorkspaceEntryFields.LocationName,
			Image:            meta.Image,
			CodeSource:       meta.CodeSource,
			WorkingDirectory: meta.WorkingDirectory,
			ExecutablePath:   meta.ExecutablePath,
			Attribute:        meta.Attribute,
			AgentQueue:       meta.AgentQueue,
			CommitHash:       meta.CommitHash,
			URL:              meta.URL,
		})
	}
	return locations, nil
}

// UpdateCodeLocation updates an existing code location (same API as add).
func (c *Client) UpdateCodeLocation(ctx context.Context, deployment string, loc CodeLocationInput) (*CodeLocation, error) {
	return c.AddCodeLocation(ctx, deployment, loc)
}

// DeleteCodeLocation removes a code location from a deployment.
func (c *Client) DeleteCodeLocation(ctx context.Context, deployment, name string) error {
	_, err := schema.DeleteCodeLocation(ctx, c.gqlClient(deployment), name)
	if err != nil {
		return fmt.Errorf("DeleteCodeLocation: %w", err)
	}
	return nil
}
