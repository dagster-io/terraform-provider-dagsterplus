package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestCodeLocationFromDocumentResource_Schema(t *testing.T) {
	r := NewCodeLocationFromDocumentResource().(*codeLocationFromDocumentResource)
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	// id: Computed only
	idRaw, ok := resp.Schema.Attributes["id"]
	if !ok {
		t.Fatal("missing 'id' attribute")
	}
	idAttr, ok := idRaw.(rsschema.StringAttribute)
	if !ok {
		t.Fatalf("id should be StringAttribute, got %T", idRaw)
	}
	if !idAttr.IsComputed() {
		t.Error("id should be Computed")
	}
	if idAttr.IsRequired() {
		t.Error("id should not be Required")
	}

	// deployment: Required
	deploymentRaw, ok := resp.Schema.Attributes["deployment"]
	if !ok {
		t.Fatal("missing 'deployment' attribute")
	}
	deploymentAttr, ok := deploymentRaw.(rsschema.StringAttribute)
	if !ok {
		t.Fatalf("deployment should be StringAttribute, got %T", deploymentRaw)
	}
	if !deploymentAttr.IsRequired() {
		t.Error("deployment should be Required")
	}

	// document: Required
	documentRaw, ok := resp.Schema.Attributes["document"]
	if !ok {
		t.Fatal("missing 'document' attribute")
	}
	documentAttr, ok := documentRaw.(rsschema.StringAttribute)
	if !ok {
		t.Fatalf("document should be StringAttribute, got %T", documentRaw)
	}
	if !documentAttr.IsRequired() {
		t.Error("document should be Required")
	}

	// name: Computed only (derived from document)
	nameRaw, ok := resp.Schema.Attributes["name"]
	if !ok {
		t.Fatal("missing 'name' attribute")
	}
	nameAttr, ok := nameRaw.(rsschema.StringAttribute)
	if !ok {
		t.Fatalf("name should be StringAttribute, got %T", nameRaw)
	}
	if !nameAttr.IsComputed() {
		t.Error("name should be Computed")
	}
	if nameAttr.IsRequired() {
		t.Error("name should not be Required")
	}
}
