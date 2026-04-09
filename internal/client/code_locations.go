package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client/schema"
)

// locationDocument is the snake_case document format expected by addOrUpdateLocationFromDocument.
type locationDocument struct {
	LocationName     string              `json:"location_name"`
	Image            string              `json:"image,omitempty"`
	CodeSource       *documentCodeSource `json:"code_source,omitempty"`
	WorkingDirectory string              `json:"working_directory,omitempty"`
	ExecutablePath   string              `json:"executable_path,omitempty"`
	Attribute        string              `json:"attribute,omitempty"`
	AgentQueue       string              `json:"agent_queue,omitempty"`
	Git              *documentGit        `json:"git,omitempty"`
	ContainerContext map[string]any      `json:"container_context,omitempty"`
}

// documentCodeSource is the code_source block in the document format.
type documentCodeSource struct {
	PythonFile  string `json:"python_file,omitempty"`
	PackageName string `json:"package_name,omitempty"`
	ModuleName  string `json:"module_name,omitempty"`
}

// documentGit is the git block in the document format.
type documentGit struct {
	CommitHash string `json:"commit_hash,omitempty"`
	URL        string `json:"url,omitempty"`
}

// userDocument is the camelCase document format used by the dagsterplus_code_location_from_document
// resource. This is the user-facing format and must not change to avoid breaking existing configs.
type userDocument struct {
	LocationName     string         `json:"locationName"`
	Image            string         `json:"image,omitempty"`
	CodeSource       userCodeSource `json:"codeSource"`
	WorkingDirectory string         `json:"workingDirectory,omitempty"`
	ExecutablePath   string         `json:"executablePath,omitempty"`
	Attribute        string         `json:"attribute,omitempty"`
	AgentQueue       string         `json:"agentQueue,omitempty"`
	CommitHash       string         `json:"commitHash,omitempty"`
	URL              string         `json:"url,omitempty"`
	ContainerContext map[string]any `json:"containerContext,omitempty"`
}

type userCodeSource struct {
	PythonFile  string `json:"pythonFile,omitempty"`
	PackageName string `json:"packageName,omitempty"`
	ModuleName  string `json:"moduleName,omitempty"`
}

// CodeSource describes where the code lives within a code location.
type CodeSource struct {
	PythonFile  string `json:"python_file,omitempty"`
	PackageName string `json:"package_name,omitempty"`
	ModuleName  string `json:"module_name,omitempty"`
}

// CodeLocation represents a Dagster+ code location.
type CodeLocation struct {
	ID               string
	Name             string
	Image            string
	CodeSource       CodeSource
	WorkingDirectory string
	ExecutablePath   string
	Attribute        string
	AgentQueue       string
	CommitHash       string
	URL              string
	ContainerContext map[string]any
}

// CodeLocationInput is the input for creating or updating a code location.
type CodeLocationInput struct {
	Name             string
	Image            string
	CodeSource       CodeSource
	WorkingDirectory string
	ExecutablePath   string
	Attribute        string
	AgentQueue       string
	CommitHash       string
	URL              string
	ContainerContext map[string]any
}

func inputToDocument(loc CodeLocationInput) locationDocument {
	doc := locationDocument{
		LocationName:     loc.Name,
		Image:            loc.Image,
		WorkingDirectory: loc.WorkingDirectory,
		ExecutablePath:   loc.ExecutablePath,
		Attribute:        loc.Attribute,
		AgentQueue:       loc.AgentQueue,
		ContainerContext: loc.ContainerContext,
	}
	if loc.CodeSource.PythonFile != "" || loc.CodeSource.PackageName != "" || loc.CodeSource.ModuleName != "" {
		doc.CodeSource = &documentCodeSource{
			PythonFile:  loc.CodeSource.PythonFile,
			PackageName: loc.CodeSource.PackageName,
			ModuleName:  loc.CodeSource.ModuleName,
		}
	}
	if loc.CommitHash != "" || loc.URL != "" {
		doc.Git = &documentGit{
			CommitHash: loc.CommitHash,
			URL:        loc.URL,
		}
	}
	return doc
}

// ParseCodeLocationDocument parses a user-facing JSON document (camelCase keys) into a
// CodeLocationInput and returns the location name. Used by dagsterplus_code_location_from_document.
func ParseCodeLocationDocument(docJSON string) (CodeLocationInput, string, error) {
	var doc userDocument
	if err := json.Unmarshal([]byte(docJSON), &doc); err != nil {
		return CodeLocationInput{}, "", fmt.Errorf("ParseCodeLocationDocument: %w", err)
	}
	if doc.LocationName == "" {
		return CodeLocationInput{}, "", fmt.Errorf("ParseCodeLocationDocument: document must include 'locationName'")
	}
	input := CodeLocationInput{
		Name:  doc.LocationName,
		Image: doc.Image,
		CodeSource: CodeSource{
			PythonFile:  doc.CodeSource.PythonFile,
			PackageName: doc.CodeSource.PackageName,
			ModuleName:  doc.CodeSource.ModuleName,
		},
		WorkingDirectory: doc.WorkingDirectory,
		ExecutablePath:   doc.ExecutablePath,
		Attribute:        doc.Attribute,
		AgentQueue:       doc.AgentQueue,
		CommitHash:       doc.CommitHash,
		URL:              doc.URL,
		ContainerContext: doc.ContainerContext,
	}
	return input, doc.LocationName, nil
}

// CodeLocationToDocument serializes a CodeLocation to the user-facing JSON document format
// (camelCase keys). Used by dagsterplus_code_location_from_document.
func CodeLocationToDocument(loc *CodeLocation) (string, error) {
	doc := userDocument{
		LocationName: loc.Name,
		Image:        loc.Image,
		CodeSource: userCodeSource{
			PythonFile:  loc.CodeSource.PythonFile,
			PackageName: loc.CodeSource.PackageName,
			ModuleName:  loc.CodeSource.ModuleName,
		},
		WorkingDirectory: loc.WorkingDirectory,
		ExecutablePath:   loc.ExecutablePath,
		Attribute:        loc.Attribute,
		AgentQueue:       loc.AgentQueue,
		CommitHash:       loc.CommitHash,
		URL:              loc.URL,
		ContainerContext: loc.ContainerContext,
	}
	b, err := json.Marshal(doc)
	if err != nil {
		return "", fmt.Errorf("CodeLocationToDocument: %w", err)
	}
	return string(b), nil
}

// AddCodeLocation adds a code location to a deployment.
func (c *Client) AddCodeLocation(ctx context.Context, deployment string, loc CodeLocationInput) (*CodeLocation, error) {
	doc := inputToDocument(loc)
	docBytes, err := json.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("AddCodeLocation: marshalling document: %w", err)
	}

	resp, err := schema.AddOrUpdateCodeLocation(ctx, c.gqlClient(deployment), json.RawMessage(docBytes))
	if err != nil {
		return nil, fmt.Errorf("AddCodeLocation: %w", err)
	}

	switch resp.AddOrUpdateLocationFromDocument.(type) {
	case *schema.AddOrUpdateCodeLocationAddOrUpdateLocationFromDocumentWorkspaceEntry:
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
			ContainerContext: loc.ContainerContext,
		}
		return cl, nil
	default:
		return nil, fmt.Errorf("AddCodeLocation: unexpected result type %T", resp.AddOrUpdateLocationFromDocument)
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

// metadataString extracts a string value from a metadata map, returning "" if absent or wrong type.
func metadataString(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// parseSerializedMetadata parses the serializedDeploymentMetadata JSON into a CodeLocation.
// The metadata is produced by dagster serdes and uses snake_case field names.
func parseSerializedMetadata(deployment, raw string) (CodeLocation, error) {
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return CodeLocation{}, err
	}

	name := metadataString(m, "location_name")

	loc := CodeLocation{
		ID:               fmt.Sprintf("%s/%s", deployment, name),
		Name:             name,
		Image:            metadataString(m, "image"),
		WorkingDirectory: metadataString(m, "working_directory"),
		ExecutablePath:   metadataString(m, "executable_path"),
		Attribute:        metadataString(m, "attribute"),
		CodeSource: CodeSource{
			PythonFile:  metadataString(m, "python_file"),
			PackageName: metadataString(m, "package_name"),
			ModuleName:  metadataString(m, "module_name"),
		},
	}

	// agent_queue may be a string or an AgentQueue record object.
	if v, ok := m["agent_queue"]; ok {
		switch aq := v.(type) {
		case string:
			loc.AgentQueue = aq
		case map[string]any:
			loc.AgentQueue = metadataString(aq, "queue_name")
		}
	}

	// git_metadata is a nested GitMetadata record object.
	if v, ok := m["git_metadata"]; ok {
		if gm, ok := v.(map[string]any); ok {
			loc.CommitHash = metadataString(gm, "commit_hash")
			loc.URL = metadataString(gm, "url")
		}
	}

	if v, ok := m["container_context"]; ok {
		if cc, ok := v.(map[string]any); ok {
			loc.ContainerContext = cc
		}
	}

	return loc, nil
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
		loc, err := parseSerializedMetadata(deployment, e.WorkspaceEntryFields.SerializedDeploymentMetadata)
		if err != nil {
			// Skip entries with unparseable metadata rather than failing the whole list.
			continue
		}
		// Use the workspace entry's locationName as the authoritative name.
		loc.Name = e.WorkspaceEntryFields.LocationName
		loc.ID = fmt.Sprintf("%s/%s", deployment, loc.Name)
		locations = append(locations, loc)
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
