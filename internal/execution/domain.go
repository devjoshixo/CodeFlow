package execution

import (
	"errors"
	"time"
)

type Execution struct {
	ID         string
	UserID     string
	Language   string
	Code       string
	Status     Status
	Output     *string
	Error      *string
	ExitCode   int
	DurationMs int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Status string

const (
	StatusPending   Status = "pending"
	StatusQueued    Status = "queued"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusTimeout   Status = "timeout"
)

var (
	ErrExecutionNotFound       = errors.New("execution not found")
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrInvalidLanguage         = errors.New("unsupported programming lanugage")
	ErrCodeTooLarge            = errors.New("code exceeds maximum size")
)

var SupportedLanguages = map[string]struct{}{
	"python":     {},
	"go":         {},
	"javascript": {},
}

func (s Status) CanTransition(to Status) bool {
	switch s {
	case StatusPending:
		return to == StatusQueued
	case StatusQueued:
		return to == StatusRunning
	case StatusRunning:
		return to == StatusCompleted || to == StatusFailed || to == StatusTimeout
	default:
		return false
	}
}

const MaxCodeSize = 10 * 1024

func (e *Execution) Validate() error {
	if e.UserID == "" {
		return errors.New("user_id required")
	}

	if _, ok := SupportedLanguages[e.Language]; !ok {
		return ErrInvalidLanguage
	}
	if len(e.Code) > MaxCodeSize {
		return ErrCodeTooLarge
	}
	return nil
}
