package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Khan/genqlient/graphql"
)

// Client is the Dagster+ API client.
type Client struct {
	Organization string
	BaseURL      string
	httpClient   *http.Client
	Token        string
}

// New creates a new Dagster+ API client.
func New(organization, token, baseURL string) *Client {
	if baseURL == "" {
		baseURL = fmt.Sprintf("https://%s.dagster.cloud", organization)
	}
	baseURL = strings.TrimRight(baseURL, "/")
	return &Client{
		Organization: organization,
		BaseURL:      baseURL,
		Token:        token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// tokenDoer injects the API token header into every request.
type tokenDoer struct {
	token string
	inner *http.Client
}

func (t *tokenDoer) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Dagster-Cloud-Api-Token", t.token)
	resp, err := t.inner.Do(req)
	if err != nil || resp.StatusCode != http.StatusInternalServerError {
		return resp, err
	}
	// Some Dagster Cloud API versions return HTTP 500 when a GraphQL mutation
	// resolves to a PythonError union member. Rewrite those responses to 200 so
	// genqlient can parse and surface the error type through the normal union
	// switch, rather than returning a raw HTTP error.
	body, readErr := io.ReadAll(resp.Body)
	resp.Body.Close()
	if readErr != nil {
		resp.Body = http.NoBody
		return resp, nil
	}
	var check struct {
		Data json.RawMessage `json:"data"`
	}
	if json.Unmarshal(body, &check) == nil && len(check.Data) > 0 {
		resp.StatusCode = http.StatusOK
		resp.Status = "200 OK"
	}
	resp.Body = io.NopCloser(bytes.NewReader(body))
	return resp, nil
}

// gqlClient returns a genqlient GraphQL client for the given deployment.
// Pass deployment="" to target the org-level endpoint.
func (c *Client) gqlClient(deployment string) graphql.Client {
	var url string
	if deployment == "" {
		url = fmt.Sprintf("%s/graphql", c.BaseURL)
	} else {
		url = fmt.Sprintf("%s/%s/graphql", c.BaseURL, deployment)
	}
	return graphql.NewClient(url, &tokenDoer{token: c.Token, inner: c.httpClient})
}
