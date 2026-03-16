package datasources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func TestCodeLocationDataSource_Schema(t *testing.T) {
	d := NewCodeLocationDataSource().(*codeLocationDataSource)
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
		t.Error("deployment should not be Computed")
	}

	// name: Required (lookup key)
	nameRaw, ok := resp.Schema.Attributes["name"]
	if !ok {
		t.Fatal("missing 'name' attribute")
	}
	nameAttr, ok := nameRaw.(dsschema.StringAttribute)
	if !ok {
		t.Fatalf("name should be StringAttribute, got %T", nameRaw)
	}
	if !nameAttr.IsRequired() {
		t.Error("name should be Required")
	}
	if nameAttr.IsComputed() {
		t.Error("name should not be Computed")
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
	if idAttr.IsRequired() {
		t.Error("id should not be Required")
	}

	// image: Computed only
	imageRaw, ok := resp.Schema.Attributes["image"]
	if !ok {
		t.Fatal("missing 'image' attribute")
	}
	imageAttr, ok := imageRaw.(dsschema.StringAttribute)
	if !ok {
		t.Fatalf("image should be StringAttribute, got %T", imageRaw)
	}
	if !imageAttr.IsComputed() {
		t.Error("image should be Computed")
	}

	// code_source block must be present
	codeSourceBlock, ok := resp.Schema.Blocks["code_source"]
	if !ok {
		t.Fatal("missing 'code_source' block")
	}
	if _, ok := codeSourceBlock.(dsschema.SingleNestedBlock); !ok {
		t.Errorf("code_source should be SingleNestedBlock, got %T", codeSourceBlock)
	}
}
