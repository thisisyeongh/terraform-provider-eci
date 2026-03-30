package resource

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"terraform-provider-eci/internal/api"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func TestAddResourceError_withResourceId(t *testing.T) {
	t.Parallel()

	diags := diag.Diagnostics{}
	addResourceError(&diags, "failed to create", "abc-123", fmt.Errorf("timeout"))

	if !diags.HasError() {
		t.Fatal("expected error diagnostic")
	}
	detail := diags.Errors()[0].Detail()
	if !strings.Contains(detail, "abc-123") {
		t.Errorf("expected detail to contain resource id, got %s", detail)
	}
}

func TestAddResourceError_withoutResourceId(t *testing.T) {
	t.Parallel()

	diags := diag.Diagnostics{}
	addResourceError(&diags, "failed to create", "", fmt.Errorf("timeout"))

	if !diags.HasError() {
		t.Fatal("expected error diagnostic")
	}
	detail := diags.Errors()[0].Detail()
	if strings.Contains(detail, "resource id") {
		t.Errorf("expected detail without resource id, got %s", detail)
	}
}

func TestIsResourceDeleted_nilError(t *testing.T) {
	t.Parallel()

	msg, err := isResourceDeleted(nil, "resource_virtual_network", "deleted")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if msg != "successfully deleted" {
		t.Errorf("expected 'successfully deleted', got %s", msg)
	}
}

func TestIsResourceDeleted_404(t *testing.T) {
	t.Parallel()

	code := "not_found"
	apiErr := &api.APIError{
		HttpCode: 404,
		Code:     &code,
	}

	msg, err := isResourceDeleted(apiErr, "resource_virtual_network", "deleted")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if msg != "resource does not exist" {
		t.Errorf("expected 'resource does not exist', got %s", msg)
	}
}

func TestIsResourceDeleted_409_deleted(t *testing.T) {
	t.Parallel()

	code := "unexpected_status"
	detail := map[string]interface{}{
		"resource_virtual_network": map[string]interface{}{
			"status": "deleted",
		},
	}
	apiErr := &api.APIError{
		HttpCode: 409,
		Code:     &code,
		Detail:   &detail,
	}

	msg, err := isResourceDeleted(apiErr, "resource_virtual_network", "deleted")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if msg != "resource is already deleted" {
		t.Errorf("expected 'resource is already deleted', got %s", msg)
	}
}

func TestIsResourceDeleted_409_otherStatus(t *testing.T) {
	t.Parallel()

	code := "unexpected_status"
	detail := map[string]interface{}{
		"resource_virtual_network": map[string]interface{}{
			"status": "active",
		},
	}
	apiErr := &api.APIError{
		HttpCode: 409,
		Code:     &code,
		Detail:   &detail,
	}

	_, err := isResourceDeleted(apiErr, "resource_virtual_network", "deleted")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestIsResourceDeleted_nonAPIError(t *testing.T) {
	t.Parallel()

	_, err := isResourceDeleted(fmt.Errorf("network error"), "resource_virtual_network", "deleted")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestWaitStatus_immediateMatch(t *testing.T) {
	t.Parallel()

	status := "active"
	getStatus := func() (*string, error) {
		return &status, nil
	}

	result, diags := waitStatus(getStatus, []string{"active"}, 1)

	if diags.HasError() {
		t.Fatalf("expected no error, got %v", diags)
	}
	if result == nil || *result != "active" {
		t.Error("expected active status")
	}
}

func TestWaitStatus_maxRetryExceeded(t *testing.T) {
	t.Parallel()

	status := "pending"
	getStatus := func() (*string, error) {
		return &status, nil
	}

	result, diags := waitStatus(getStatus, []string{"active"}, 1)

	if !diags.HasError() {
		t.Error("expected error diagnostic for max retry")
	}
	if result != nil {
		t.Error("expected nil result")
	}
}

func TestWaitStatus_errorThenSuccess(t *testing.T) {
	t.Parallel()

	callCount := 0
	getStatus := func() (*string, error) {
		callCount++
		if callCount == 1 {
			return nil, errors.New("temporary error")
		}
		status := "active"
		return &status, nil
	}

	result, diags := waitStatus(getStatus, []string{"active"}, 3)

	if diags.HasError() {
		t.Fatalf("expected no error, got %v", diags)
	}
	if result == nil || *result != "active" {
		t.Error("expected active status")
	}
}
