package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
"time"
)

// Client is the Dagster+ API client.
type Client struct {
	Organization string
	BaseURL      string
	HTTPClient   *http.Client
	Token        string
}

// New creates a new Dagster+ API client.
func New(organization, token, baseURL string) *Client {
	if baseURL == "" {
		baseURL = fmt.Sprintf("https://%s.dagster.cloud", organization)
	}
	return &Client{
		Organization: organization,
		BaseURL:      baseURL,
		Token:        token,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// graphQLRequest is the body sent to the GraphQL endpoint.
type graphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

// graphQLResponse is the envelope returned by the GraphQL endpoint.
type graphQLResponse struct {
	Data   json.RawMessage  `json:"data"`
	Errors []graphQLError   `json:"errors"`
}

type graphQLError struct {
	Message string `json:"message"`
}

// doGraphQL executes a GraphQL request against the deployment-scoped endpoint.
// Pass deployment="" to target the org-level endpoint.
func (c *Client) doGraphQL(ctx context.Context, deployment, query string, variables map[string]any, out any) error {
	var endpoint string
	if deployment == "" {
		endpoint = fmt.Sprintf("%s/graphql", c.BaseURL)
	} else {
		endpoint = fmt.Sprintf("%s/%s/graphql", c.BaseURL, deployment)
	}

	body, err := json.Marshal(graphQLRequest{Query: query, Variables: variables})
	if err != nil {
		return fmt.Errorf("marshalling GraphQL request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("building HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Dagster-Cloud-Api-Token", c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing HTTP request: %w", err)
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(rawBody))
	}

	var gqlResp graphQLResponse
	if err := json.Unmarshal(rawBody, &gqlResp); err != nil {
		return fmt.Errorf("decoding GraphQL response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		return fmt.Errorf("GraphQL error: %s", gqlResp.Errors[0].Message)
	}

	if out != nil {
		if err := json.Unmarshal(gqlResp.Data, out); err != nil {
			return fmt.Errorf("decoding GraphQL data: %w", err)
		}
	}

	return nil
}
