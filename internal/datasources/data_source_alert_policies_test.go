package datasources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func TestAlertPoliciesDataSource_Schema(t *testing.T) {
	d := NewAlertPoliciesDataSource().(*alertPoliciesDataSource)
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	// deployment: Required (lookup key)
	deploymentRaw, ok := resp.Schema.Attributes["deployment"]
	if !ok {
		t.Fatal("missing 'deployment' attribute")
	}
	deploymentAttr, ok := deploymentRaw.(dsschema.StringAttribute)
	if !ok {
		t.Fatalf("deployment should be StringAttribute, got %T", deploymentRaw)
	}
	if !deploymentAttr.IsRequired() {
		t.Error("deployment should be Required")
	}
	if deploymentAttr.IsComputed() {
		t.Error("deployment should not be Computed (it is the lookup key)")
	}

	// id: Computed only
	idRaw, ok := resp.Schema.Attributes["id"]
	if !ok {
		t.Fatal("missing 'id' attribute")
	}
	idAttr, ok := idRaw.(dsschema.StringAttribute)
	if !ok {
		t.Fatalf("id should be StringAttribute, got %T", idRaw)
	}
	if !idAttr.IsComputed() {
		t.Error("id should be Computed")
	}

	// alert_policies: Computed list nested attribute
	apRaw, ok := resp.Schema.Attributes["alert_policies"]
	if !ok {
		t.Fatal("missing 'alert_policies' attribute")
	}
	apAttr, ok := apRaw.(dsschema.ListNestedAttribute)
	if !ok {
		t.Fatalf("alert_policies should be ListNestedAttribute, got %T", apRaw)
	}
	if !apAttr.IsComputed() {
		t.Error("alert_policies should be Computed")
	}

	// notification_service nested attribute inside alert_policies — verify it is Computed
	nsRaw, ok := apAttr.NestedObject.Attributes["notification_service"]
	if !ok {
		t.Fatal("alert_policies missing 'notification_service' nested attribute")
	}
	nsAttr, ok := nsRaw.(dsschema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("notification_service should be SingleNestedAttribute, got %T", nsRaw)
	}
	if !nsAttr.IsComputed() {
		t.Error("notification_service should be Computed")
	}

	// webhook_url inside notification_service: Computed, Sensitive
	webhookRaw, ok := nsAttr.Attributes["webhook_url"]
	if !ok {
		t.Fatal("notification_service missing 'webhook_url' attribute")
	}
	webhookAttr, ok := webhookRaw.(dsschema.StringAttribute)
	if !ok {
		t.Fatalf("webhook_url should be StringAttribute, got %T", webhookRaw)
	}
	if !webhookAttr.IsComputed() {
		t.Error("webhook_url should be Computed")
	}
	if !webhookAttr.IsSensitive() {
		t.Error("webhook_url should be Sensitive")
	}
}
