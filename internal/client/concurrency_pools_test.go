package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetConcurrencyPool_Set(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "GetConcurrencyLimit") {
			t.Errorf("expected GetConcurrencyLimit query, got: %s", b.Query)
		}
		if r.URL.Path != "/prod/graphql" {
			t.Errorf("path = %q, want /prod/graphql", r.URL.Path)
		}
		if b.Variables["concurrencyKey"] != "database" {
			t.Errorf("concurrencyKey = %v, want database", b.Variables["concurrencyKey"])
		}
		gqlOK(w, map[string]any{
			"instance": map[string]any{
				"concurrencyLimit": map[string]any{
					"concurrencyKey":    "database",
					"limit":             5,
					"usingDefaultLimit": false,
				},
			},
		})
	}))
	defer srv.Close()

	pool, err := newClient(srv).GetConcurrencyPool(context.Background(), "prod", "database")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !pool.IsSet {
		t.Error("expected IsSet true")
	}
	if pool.Limit != 5 {
		t.Errorf("limit = %d, want 5", pool.Limit)
	}
	if pool.Name != "database" {
		t.Errorf("name = %q, want database", pool.Name)
	}
}

func TestGetConcurrencyPool_UsingDefault(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"instance": map[string]any{
				"concurrencyLimit": map[string]any{
					"concurrencyKey":    "database",
					"limit":             0,
					"usingDefaultLimit": true,
				},
			},
		})
	}))
	defer srv.Close()

	pool, err := newClient(srv).GetConcurrencyPool(context.Background(), "prod", "database")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pool.IsSet {
		t.Error("expected IsSet false when using default limit")
	}
}

func TestGetConcurrencyPool_GraphQLError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlErr(w, "permission denied")
	}))
	defer srv.Close()

	_, err := newClient(srv).GetConcurrencyPool(context.Background(), "prod", "database")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSetConcurrencyPool_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "SetConcurrencyLimit") {
			t.Errorf("expected SetConcurrencyLimit mutation, got: %s", b.Query)
		}
		if b.Variables["limit"].(float64) != 1 {
			t.Errorf("limit = %v, want 1", b.Variables["limit"])
		}
		gqlOK(w, map[string]any{"setConcurrencyLimit": true})
	}))
	defer srv.Close()

	if err := newClient(srv).SetConcurrencyPool(context.Background(), "prod", "database", 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetConcurrencyPool_Failure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{"setConcurrencyLimit": false})
	}))
	defer srv.Close()

	err := newClient(srv).SetConcurrencyPool(context.Background(), "prod", "database", 1)
	if err == nil {
		t.Fatal("expected error when API returns false, got nil")
	}
}

func TestSetConcurrencyPool_GraphQLError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlErr(w, "unauthorized")
	}))
	defer srv.Close()

	err := newClient(srv).SetConcurrencyPool(context.Background(), "prod", "database", 1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDeleteConcurrencyPool_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "DeleteConcurrencyLimit") {
			t.Errorf("expected DeleteConcurrencyLimit mutation, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{"deleteConcurrencyLimit": true})
	}))
	defer srv.Close()

	if err := newClient(srv).DeleteConcurrencyPool(context.Background(), "prod", "database"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteConcurrencyPool_Failure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{"deleteConcurrencyLimit": false})
	}))
	defer srv.Close()

	err := newClient(srv).DeleteConcurrencyPool(context.Background(), "prod", "database")
	if err == nil {
		t.Fatal("expected error when API returns false, got nil")
	}
}
