package utils

import (
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestStringOrNull_nil(t *testing.T) {
	t.Parallel()

	var ptr *uuid.UUID
	result := StringOrNull(ptr)

	if !result.IsNull() {
		t.Errorf("expected null, got %s", result.ValueString())
	}
}

func TestStringOrNull_nonNil(t *testing.T) {
	t.Parallel()

	id := uuid.MustParse("12345678-1234-1234-1234-123456789012")
	result := StringOrNull(&id)

	expected := types.StringValue("12345678-1234-1234-1234-123456789012")
	if !result.Equal(expected) {
		t.Errorf("expected %s, got %s", expected.ValueString(), result.ValueString())
	}
}

func TestStringValOrNull_nil(t *testing.T) {
	t.Parallel()

	result := StringValOrNull(nil)

	if !result.IsNull() {
		t.Errorf("expected null, got %s", result.ValueString())
	}
}

func TestStringValOrNull_nonNil(t *testing.T) {
	t.Parallel()

	val := "hello"
	result := StringValOrNull(&val)

	expected := types.StringValue("hello")
	if !result.Equal(expected) {
		t.Errorf("expected %s, got %s", expected.ValueString(), result.ValueString())
	}
}
