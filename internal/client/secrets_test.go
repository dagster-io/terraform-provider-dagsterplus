package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
)

func mockSecretResponse() map[string]any {
	return map[string]any{
		"id":                            "secret-123",
		"secretName":                    "MY_SECRET",
		"secretValue":                   "s3cr3t",
		"fullDeploymentScope":           true,
		"allBranchDeploymentsScope":     false,
		"specificBranchDeploymentScope": "",
		"localDeploymentScope":          false,
		"locationNames":                 []string{"my-location"},
	}
}

func testSecret() client.Secret {
	return client.Secret{
		SecretName:          "MY_SECRET",
		SecretValue:         "s3cr3t",
		FullDeploymentScope: true,
		LocationNames:       []string{"my-location"},
	}
}

func TestListSecrets_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "ListSecrets") {
			t.Errorf("expected ListSecrets query, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"secretsOrError": map[string]any{
				"__typename": "Secrets",
				"secrets":    []any{mockSecretResponse()},
			},
		})
	}))
	defer srv.Close()

	secrets, err := newClient(srv).ListSecrets(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(secrets) != 1 {
		t.Fatalf("len = %d, want 1", len(secrets))
	}
	if secrets[0].SecretName != "MY_SECRET" {
		t.Errorf("SecretName = %q, want MY_SECRET", secrets[0].SecretName)
	}
	if secrets[0].SecretValue != "s3cr3t" {
		t.Errorf("SecretValue = %q, want s3cr3t", secrets[0].SecretValue)
	}
}

func TestListSecrets_UnexpectedType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"secretsOrError": map[string]any{
				"__typename": "UnauthorizedError",
			},
		})
	}))
	defer srv.Close()

	_, err := newClient(srv).ListSecrets(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateSecret_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "CreateSecret") {
			t.Errorf("expected CreateSecret mutation, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"createSecret": map[string]any{
				"__typename": "CreateOrUpdateSecretSuccess",
				"secret":     mockSecretResponse(),
			},
		})
	}))
	defer srv.Close()

	s, err := newClient(srv).CreateSecret(context.Background(), testSecret())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.ID != "secret-123" {
		t.Errorf("ID = %q, want secret-123", s.ID)
	}
}

func TestCreateSecret_TooMany(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"createSecret": map[string]any{
				"__typename": "TooManySecretsError",
				"message":    "too many secrets",
			},
		})
	}))
	defer srv.Close()

	_, err := newClient(srv).CreateSecret(context.Background(), testSecret())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "too many secrets") {
		t.Errorf("error should mention too many secrets, got: %v", err)
	}
}

func TestCreateSecret_GraphQLError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlErr(w, "server error")
	}))
	defer srv.Close()

	_, err := newClient(srv).CreateSecret(context.Background(), testSecret())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUpdateSecret_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "UpdateSecret") {
			t.Errorf("expected UpdateSecret mutation, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"updateSecret": map[string]any{
				"__typename": "CreateOrUpdateSecretSuccess",
				"secret":     mockSecretResponse(),
			},
		})
	}))
	defer srv.Close()

	s := testSecret()
	s.ID = "secret-123"
	result, err := newClient(srv).UpdateSecret(context.Background(), s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "secret-123" {
		t.Errorf("ID = %q, want secret-123", result.ID)
	}
}

func TestDeleteSecret_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "DeleteSecret") {
			t.Errorf("expected DeleteSecret mutation, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"deleteSecret": map[string]any{
				"__typename": "DeleteSecretSuccess",
				"secretId":   "secret-123",
			},
		})
	}))
	defer srv.Close()

	if err := newClient(srv).DeleteSecret(context.Background(), "secret-123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteSecret_UnexpectedType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"deleteSecret": map[string]any{
				"__typename": "UnauthorizedError",
			},
		})
	}))
	defer srv.Close()

	err := newClient(srv).DeleteSecret(context.Background(), "secret-123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetSecretByName_Found(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"secretsOrError": map[string]any{
				"__typename": "Secrets",
				"secrets":    []any{mockSecretResponse()},
			},
		})
	}))
	defer srv.Close()

	s, err := newClient(srv).GetSecretByName(context.Background(), "MY_SECRET")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.SecretName != "MY_SECRET" {
		t.Errorf("SecretName = %q, want MY_SECRET", s.SecretName)
	}
}

func TestGetSecretByName_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"secretsOrError": map[string]any{
				"__typename": "Secrets",
				"secrets":    []any{},
			},
		})
	}))
	defer srv.Close()

	_, err := newClient(srv).GetSecretByName(context.Background(), "NONEXISTENT")
	if err == nil {
		t.Fatal("expected error for missing secret, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention not found, got: %v", err)
	}
}
