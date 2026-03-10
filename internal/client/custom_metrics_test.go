package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func mockCustomMetric() map[string]any {
	return map[string]any{
		"id":          "cm-123",
		"metadataKey": "my_metric",
		"displayName": "My Metric",
		"description": "A test metric",
	}
}

func TestListCustomMetrics_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "ListCustomMetrics") {
			t.Errorf("expected ListCustomMetrics query, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"customMetrics": []any{mockCustomMetric()},
		})
	}))
	defer srv.Close()

	metrics, err := newClient(srv).ListCustomMetrics(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(metrics) != 1 {
		t.Fatalf("len = %d, want 1", len(metrics))
	}
	if metrics[0].MetadataKey != "my_metric" {
		t.Errorf("MetadataKey = %q, want my_metric", metrics[0].MetadataKey)
	}
}

func TestCreateCustomMetric_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "CreateCustomMetric") {
			t.Errorf("expected CreateCustomMetric mutation, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"createCustomMetric": map[string]any{
				"__typename":   "CreateCustomMetricSuccess",
				"customMetric": mockCustomMetric(),
			},
		})
	}))
	defer srv.Close()

	m, err := newClient(srv).CreateCustomMetric(context.Background(), "my_metric", "My Metric", "A test metric")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.ID != "cm-123" {
		t.Errorf("ID = %q, want cm-123", m.ID)
	}
}

func TestCreateCustomMetric_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"createCustomMetric": map[string]any{
				"__typename": "CustomMetricError",
				"message":    "metric key already exists",
			},
		})
	}))
	defer srv.Close()

	_, err := newClient(srv).CreateCustomMetric(context.Background(), "my_metric", "", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "metric key already exists") {
		t.Errorf("error should mention metric error, got: %v", err)
	}
}

func TestCreateCustomMetric_GraphQLError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlErr(w, "permission denied")
	}))
	defer srv.Close()

	_, err := newClient(srv).CreateCustomMetric(context.Background(), "my_metric", "", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUpdateCustomMetric_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "UpdateCustomMetric") {
			t.Errorf("expected UpdateCustomMetric mutation, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"updateCustomMetric": map[string]any{
				"__typename":   "UpdateCustomMetricSuccess",
				"customMetric": mockCustomMetric(),
			},
		})
	}))
	defer srv.Close()

	m, err := newClient(srv).UpdateCustomMetric(context.Background(), "cm-123", "Updated Metric", "Updated desc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.ID != "cm-123" {
		t.Errorf("ID = %q, want cm-123", m.ID)
	}
}

func TestDeleteCustomMetric_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "DeleteCustomMetric") {
			t.Errorf("expected DeleteCustomMetric mutation, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"deleteCustomMetric": map[string]any{
				"__typename":     "DeleteCustomMetricSuccess",
				"customMetricId": "cm-123",
			},
		})
	}))
	defer srv.Close()

	if err := newClient(srv).DeleteCustomMetric(context.Background(), "cm-123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteCustomMetric_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"deleteCustomMetric": map[string]any{
				"__typename": "CustomMetricError",
				"message":    "metric not found",
			},
		})
	}))
	defer srv.Close()

	err := newClient(srv).DeleteCustomMetric(context.Background(), "cm-missing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetCustomMetric_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"customMetrics": []any{},
		})
	}))
	defer srv.Close()

	_, err := newClient(srv).GetCustomMetric(context.Background(), "cm-nonexistent")
	if err == nil {
		t.Fatal("expected error for missing metric, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention not found, got: %v", err)
	}
}
