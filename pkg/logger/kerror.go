package logger

import "fmt"

type Code int

const (
	CodeUnknown Code = iota

	// User
	CodeInvalidArgument
	CodeConfiguration

	// Runtime
	CodeSystem
	CodePermission
	CodeNetwork

	//Kubernetes
	CodeKubernetes

	// infrastructure
	CodeContainer
	CodeSSH
)

func (c Code) String() string {
	switch c {
	case CodeInvalidArgument:
		return "invalid argument"
	case CodeConfiguration:
		return "configuration error"
	case CodeSystem:
		return "system error"
	case CodePermission:
		return "permission denied"
	case CodeNetwork:
		return "nerwork error"
	case CodeKubernetes:
		return "kubernetes error"
	case CodeContainer:
		return "contaoner error"
	case CodeSSH:
		return "ssh error"
	default:
		return "unknown error"
	}
}

// Error represents an application-level error
type Error struct {
	Code    Code
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err == nil {
		return e.Message
	}
	return fmt.Sprintf("%s: %v Exit Code [%v] ", e.Message, e.Err, e.Code)
}
func (e *Error) Unwrap() error {
	return e.Err
}

// New creates a new application error
func New(code Code, message string) error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// Wrap wraps an existing error with an application error code
func Wrap(code Code, message string, err error) error {
	return &Error{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
