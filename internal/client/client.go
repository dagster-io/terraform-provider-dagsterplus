package client

import (
	"fmt"
	"net/http"
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
	return t.inner.Do(req)
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
