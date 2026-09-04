package result_test

import (
	"testing"

	"github.com/gspaim/Runtgine/internal/core/result"
)

func TestErrorMessageFormat(t *testing.T) {
	e := result.Validation(result.CodeInvalidInput, "bad input", nil)
	if e.Error() != "validation.invalid_input: bad input" {
		t.Fatalf("err=%q", e.Error())
	}
}

func TestValidationIsNotRetryable(t *testing.T) {
	e := result.Validation(result.CodeUnknownCapability, "cap", nil)
	if e.Retryable {
		t.Fatal("validation errors must not be retryable")
	}
	if e.Code != result.CodeUnknownCapability {
		t.Fatalf("code=%s", e.Code)
	}
}

func TestRuntimeKeepsRetryableFlag(t *testing.T) {
	retryable := result.Runtime(result.CodeTimeout, "timeout", true, nil)
	if !retryable.Retryable {
		t.Fatal("timeout must be retryable")
	}
	fatal := result.Runtime(result.CodePlayerError, "boom", false, map[string]any{"exit_code": 1})
	if fatal.Retryable {
		t.Fatal("player error must not be retryable")
	}
	if fatal.Details["exit_code"] != 1 {
		t.Fatalf("details=%v", fatal.Details)
	}
}

func TestErrorCodesAreDistinct(t *testing.T) {
	codes := []string{
		result.CodeSchema,
		result.CodeUnknownCapability,
		result.CodeInvalidInput,
		result.CodePlayerError,
		result.CodeTimeout,
		result.CodeCancelled,
		result.CodeInternal,
		result.CodePolicyDenied,
		result.CodePolicyApprovalDenied,
		result.CodePolicyNotWaiting,
		result.CodeClaimConflict,
		result.CodeNotFound,
		result.CodeUnauthorized,
		result.CodeInputLimit,
	}
	seen := map[string]bool{}
	for _, c := range codes {
		if c == "" {
			t.Fatal("empty code")
		}
		if seen[c] {
			t.Fatalf("duplicate code %q", c)
		}
		seen[c] = true
	}
}
