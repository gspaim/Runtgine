package result

import "encoding/json"

type Result struct {
	OK         bool            `json:"ok"`
	StepID     string          `json:"step_id"`
	Capability string          `json:"capability"`
	Player     string          `json:"player"`
	Output     json.RawMessage `json:"output,omitempty"`
	DurationMS int64           `json:"duration_ms"`
	Attempts   int             `json:"attempts,omitempty"`
}

type Error struct {
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	Retryable bool           `json:"retryable"`
	Details   map[string]any `json:"details,omitempty"`
}

func (e Error) Error() string {
	return e.Code + ": " + e.Message
}

func Validation(code, msg string, details map[string]any) Error {
	return Error{Code: code, Message: msg, Retryable: false, Details: details}
}

func Runtime(code, msg string, retryable bool, details map[string]any) Error {
	return Error{Code: code, Message: msg, Retryable: retryable, Details: details}
}

const (
	CodeSchema               = "validation.schema"
	CodeUnknownCapability    = "validation.unknown_capability"
	CodeInvalidInput         = "validation.invalid_input"
	CodePlayerError          = "runtime.player_error"
	CodeTimeout              = "runtime.timeout"
	CodeCancelled            = "runtime.cancelled"
	CodeInternal             = "runtime.internal"
	CodePolicyDenied         = "policy.denied"
	CodePolicyApprovalDenied = "policy.approval_denied"
	CodePolicyNotWaiting     = "policy.not_waiting"
	CodeClaimConflict        = "claim.conflict"
)
