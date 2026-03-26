package api

import (
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	t.Parallel()

	code := "not_found"
	message := "resource not found"
	detail := map[string]interface{}{"key": "value"}

	err := &APIError{
		HttpCode: 404,
		Code:     &code,
		Message:  &message,
		Detail:   &detail,
	}

	result := err.Error()
	if result == "" {
		t.Error("expected non-empty error string")
	}
}

func TestAPIError_Error_nilFields(t *testing.T) {
	t.Parallel()

	err := &APIError{
		HttpCode: 500,
		Code:     nil,
		Message:  nil,
		Detail:   nil,
	}

	result := err.Error()
	if result == "" {
		t.Error("expected non-empty error string")
	}
}

func TestAPIError_IsCode_match(t *testing.T) {
	t.Parallel()

	code := "unexpected_status"
	err := &APIError{Code: &code}

	if !err.IsCode("unexpected_status") {
		t.Error("expected IsCode to return true")
	}
}

func TestAPIError_IsCode_noMatch(t *testing.T) {
	t.Parallel()

	code := "not_found"
	err := &APIError{Code: &code}

	if err.IsCode("unexpected_status") {
		t.Error("expected IsCode to return false")
	}
}

func TestAPIError_IsCode_nilCode(t *testing.T) {
	t.Parallel()

	err := &APIError{Code: nil}

	if err.IsCode("anything") {
		t.Error("expected IsCode to return false for nil code")
	}
}

func TestGetValue_nonNil(t *testing.T) {
	t.Parallel()

	val := "hello"
	result := getValue(&val)

	if result != "hello" {
		t.Errorf("expected hello, got %v", result)
	}
}

func TestGetValue_nil(t *testing.T) {
	t.Parallel()

	result := getValue[string](nil)

	if result != "<nil>" {
		t.Errorf("expected <nil>, got %v", result)
	}
}

func TestSetIfNotNil_withValue(t *testing.T) {
	t.Parallel()

	m := map[string]interface{}{}
	val := "test"
	setIfNotNil(m, "key", &val)

	if m["key"] != "test" {
		t.Errorf("expected test, got %v", m["key"])
	}
}

func TestSetIfNotNil_withNil(t *testing.T) {
	t.Parallel()

	m := map[string]interface{}{}
	setIfNotNil[string](m, "key", nil)

	if _, exists := m["key"]; exists {
		t.Error("expected key to not exist in map")
	}
}

func TestSetStrIfNotNil_withValue(t *testing.T) {
	t.Parallel()

	m := map[string]string{}
	val := "test"
	setStrIfNotNil(m, "key", &val)

	if m["key"] != "test" {
		t.Errorf("expected test, got %v", m["key"])
	}
}

func TestSetStrIfNotNil_withNil(t *testing.T) {
	t.Parallel()

	m := map[string]string{}
	setStrIfNotNil(m, "key", nil)

	if _, exists := m["key"]; exists {
		t.Error("expected key to not exist in map")
	}
}
