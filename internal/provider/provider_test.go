package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
)

func TestProvider_metadata(t *testing.T) {
	t.Parallel()

	p := New("test", false)()
	resp := &provider.MetadataResponse{}
	p.Metadata(context.Background(), provider.MetadataRequest{}, resp)

	if resp.TypeName != "eci" {
		t.Errorf("expected type name 'eci', got %s", resp.TypeName)
	}
	if resp.Version != "test" {
		t.Errorf("expected version 'test', got %s", resp.Version)
	}
}

func TestProvider_schema(t *testing.T) {
	t.Parallel()

	p := New("test", false)()
	resp := &provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}

	requiredAttrs := []string{"api_access_token", "api_endpoint", "zone_id"}
	for _, attr := range requiredAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected schema to contain attribute %s", attr)
		}
	}
}
