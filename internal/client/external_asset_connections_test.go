package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
)

func mockExternalAssetConnection() map[string]any {
	return map[string]any{
		"id":               "conn-123",
		"name":             "my-connection",
		"description":      "A test connection",
		"connectionType":   "SNOWFLAKE",
		"sourceConfigYaml": "account: my_account",
		"cronString":       "0 * * * *",
		"timezone":         "UTC",
		"scheduleStatus":   "RUNNING",
	}
}

func testExternalAssetConnection() client.ExternalAssetConnection {
	return client.ExternalAssetConnection{
		Name:             "my-connection",
		Description:      "A test connection",
		ConnectionType:   "SNOWFLAKE",
		SourceConfigYaml: "account: my_account",
		CronString:       "0 * * * *",
		Timezone:         "UTC",
	}
}

func TestListExternalAssetConnections_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "ListExternalAssetConnections") {
			t.Errorf("expected ListExternalAssetConnections query, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"externalAssetConnections": []any{mockExternalAssetConnection()},
		})
	}))
	defer srv.Close()

	conns, err := newClient(srv).ListExternalAssetConnections(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(conns) != 1 {
		t.Fatalf("len = %d, want 1", len(conns))
	}
	if conns[0].Name != "my-connection" {
		t.Errorf("Name = %q, want my-connection", conns[0].Name)
	}
	if conns[0].ConnectionType != "SNOWFLAKE" {
		t.Errorf("ConnectionType = %q, want SNOWFLAKE", conns[0].ConnectionType)
	}
}

func TestCreateExternalAssetConnection_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "CreateExternalAssetConnection") {
			t.Errorf("expected CreateExternalAssetConnection mutation, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"createExternalAssetConnection": map[string]any{
				"__typename":       "ExternalAssetConnection",
				"id":               "conn-123",
				"name":             "my-connection",
				"description":      "A test connection",
				"connectionType":   "SNOWFLAKE",
				"sourceConfigYaml": "account: my_account",
				"cronString":       "0 * * * *",
				"timezone":         "UTC",
				"scheduleStatus":   "RUNNING",
			},
		})
	}))
	defer srv.Close()

	conn, err := newClient(srv).CreateExternalAssetConnection(context.Background(), testExternalAssetConnection())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conn.ID != "conn-123" {
		t.Errorf("ID = %q, want conn-123", conn.ID)
	}
	if conn.ScheduleStatus != "RUNNING" {
		t.Errorf("ScheduleStatus = %q, want RUNNING", conn.ScheduleStatus)
	}
}

func TestCreateExternalAssetConnection_UnexpectedType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"createExternalAssetConnection": map[string]any{
				"__typename": "UnauthorizedError",
			},
		})
	}))
	defer srv.Close()

	_, err := newClient(srv).CreateExternalAssetConnection(context.Background(), testExternalAssetConnection())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateExternalAssetConnection_GraphQLError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlErr(w, "permission denied")
	}))
	defer srv.Close()

	_, err := newClient(srv).CreateExternalAssetConnection(context.Background(), testExternalAssetConnection())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUpdateExternalAssetConnection_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "UpdateExternalAssetConnection") {
			t.Errorf("expected UpdateExternalAssetConnection mutation, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"updateExternalAssetConnection": map[string]any{
				"__typename":       "ExternalAssetConnection",
				"id":               "conn-123",
				"name":             "my-connection-updated",
				"description":      "Updated",
				"connectionType":   "SNOWFLAKE",
				"sourceConfigYaml": "account: my_account",
				"cronString":       "0 * * * *",
				"timezone":         "UTC",
				"scheduleStatus":   "RUNNING",
			},
		})
	}))
	defer srv.Close()

	conn := testExternalAssetConnection()
	conn.ID = "conn-123"
	conn.Name = "my-connection-updated"
	result, err := newClient(srv).UpdateExternalAssetConnection(context.Background(), conn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "my-connection-updated" {
		t.Errorf("Name = %q, want my-connection-updated", result.Name)
	}
}

func TestDeleteExternalAssetConnection_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "DeleteExternalAssetConnection") {
			t.Errorf("expected DeleteExternalAssetConnection mutation, got: %s", b.Query)
		}
		gqlOK(w, map[string]any{
			"deleteExternalAssetConnection": map[string]any{
				"__typename": "DeleteExternalAssetConnectionSuccess",
				"id":         "conn-123",
			},
		})
	}))
	defer srv.Close()

	if err := newClient(srv).DeleteExternalAssetConnection(context.Background(), "conn-123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteExternalAssetConnection_UnexpectedType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"deleteExternalAssetConnection": map[string]any{
				"__typename": "UnauthorizedError",
			},
		})
	}))
	defer srv.Close()

	err := newClient(srv).DeleteExternalAssetConnection(context.Background(), "conn-123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
