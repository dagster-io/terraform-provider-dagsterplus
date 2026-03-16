package resources

import (
	"testing"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestExternalAssetConnectionFromModel_Basic(t *testing.T) {
	m := externalAssetConnectionResourceModel{
		ID:               types.StringValue("conn-id-1"),
		Name:             types.StringValue("snowflake-conn"),
		Description:      types.StringValue("Snowflake connection"),
		ConnectionType:   types.StringValue("SNOWFLAKE"),
		SourceConfigYaml: types.StringValue("host: myhost\nuser: myuser"),
		CronString:       types.StringValue("0 * * * *"),
		Timezone:         types.StringValue("UTC"),
		ScheduleStatus:   types.StringValue("RUNNING"),
	}

	conn := externalAssetConnectionFromModel(m)

	if conn.Name != "snowflake-conn" {
		t.Errorf("expected Name snowflake-conn, got %q", conn.Name)
	}
	if conn.Description != "Snowflake connection" {
		t.Errorf("expected Description 'Snowflake connection', got %q", conn.Description)
	}
	if conn.ConnectionType != "SNOWFLAKE" {
		t.Errorf("expected ConnectionType SNOWFLAKE, got %q", conn.ConnectionType)
	}
	if conn.SourceConfigYaml != "host: myhost\nuser: myuser" {
		t.Errorf("unexpected SourceConfigYaml: %q", conn.SourceConfigYaml)
	}
	if conn.CronString != "0 * * * *" {
		t.Errorf("expected CronString '0 * * * *', got %q", conn.CronString)
	}
	if conn.Timezone != "UTC" {
		t.Errorf("expected Timezone UTC, got %q", conn.Timezone)
	}
	// ID is not carried by externalAssetConnectionFromModel
	if conn.ID != "" {
		t.Errorf("expected empty ID from fromModel, got %q", conn.ID)
	}
	// ScheduleStatus is not carried (computed-only field)
	if conn.ScheduleStatus != "" {
		t.Errorf("expected empty ScheduleStatus from fromModel, got %q", conn.ScheduleStatus)
	}
}

func TestExternalAssetConnectionToModel_Basic(t *testing.T) {
	conn := &client.ExternalAssetConnection{
		ID:               "conn-abc",
		Name:             "bigquery-conn",
		Description:      "BigQuery connection",
		ConnectionType:   "BIGQUERY",
		SourceConfigYaml: "project: my-project",
		CronString:       "30 6 * * *",
		Timezone:         "America/New_York",
		ScheduleStatus:   "STOPPED",
	}

	model := externalAssetConnectionToModel(conn)

	if model.ID.ValueString() != "conn-abc" {
		t.Errorf("expected ID conn-abc, got %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "bigquery-conn" {
		t.Errorf("expected Name bigquery-conn, got %q", model.Name.ValueString())
	}
	if model.Description.ValueString() != "BigQuery connection" {
		t.Errorf("expected Description 'BigQuery connection', got %q", model.Description.ValueString())
	}
	if model.ConnectionType.ValueString() != "BIGQUERY" {
		t.Errorf("expected ConnectionType BIGQUERY, got %q", model.ConnectionType.ValueString())
	}
	if model.SourceConfigYaml.ValueString() != "project: my-project" {
		t.Errorf("unexpected SourceConfigYaml: %q", model.SourceConfigYaml.ValueString())
	}
	if model.CronString.ValueString() != "30 6 * * *" {
		t.Errorf("expected CronString '30 6 * * *', got %q", model.CronString.ValueString())
	}
	if model.Timezone.ValueString() != "America/New_York" {
		t.Errorf("expected Timezone America/New_York, got %q", model.Timezone.ValueString())
	}
	if model.ScheduleStatus.ValueString() != "STOPPED" {
		t.Errorf("expected ScheduleStatus STOPPED, got %q", model.ScheduleStatus.ValueString())
	}
}

func TestExternalAssetConnectionRoundTrip(t *testing.T) {
	original := &client.ExternalAssetConnection{
		ID:               "rt-id",
		Name:             "postgres-conn",
		Description:      "Postgres",
		ConnectionType:   "POSTGRES",
		SourceConfigYaml: "host: localhost",
		CronString:       "*/5 * * * *",
		Timezone:         "Europe/Berlin",
		ScheduleStatus:   "RUNNING",
	}

	model := externalAssetConnectionToModel(original)
	conn := externalAssetConnectionFromModel(model)
	conn.ID = original.ID // ID is not round-tripped by fromModel

	if conn.ID != original.ID {
		t.Errorf("ID mismatch: want %q got %q", original.ID, conn.ID)
	}
	if conn.Name != original.Name {
		t.Errorf("Name mismatch: want %q got %q", original.Name, conn.Name)
	}
	if conn.Description != original.Description {
		t.Errorf("Description mismatch: want %q got %q", original.Description, conn.Description)
	}
	if conn.ConnectionType != original.ConnectionType {
		t.Errorf("ConnectionType mismatch: want %q got %q", original.ConnectionType, conn.ConnectionType)
	}
	if conn.SourceConfigYaml != original.SourceConfigYaml {
		t.Errorf("SourceConfigYaml mismatch: want %q got %q", original.SourceConfigYaml, conn.SourceConfigYaml)
	}
	if conn.CronString != original.CronString {
		t.Errorf("CronString mismatch: want %q got %q", original.CronString, conn.CronString)
	}
	if conn.Timezone != original.Timezone {
		t.Errorf("Timezone mismatch: want %q got %q", original.Timezone, conn.Timezone)
	}
}
