package errors

import "fmt"

type ErrorCode string

const (
	ErrCodeValidation     ErrorCode = "VALIDATION_ERROR"
	ErrCodeConfigLoad     ErrorCode = "CONFIG_LOAD_ERROR"
	ErrCodeTemplateRender ErrorCode = "TEMPLATE_RENDER_ERROR"
	ErrCodeProxyRequest   ErrorCode = "PROXY_REQUEST_ERROR"
	ErrCodeServerStart    ErrorCode = "SERVER_START_ERROR"
	ErrCodeBehaviorInit   ErrorCode = "BEHAVIOR_INIT_ERROR"
)

type AppError struct {
	Code    ErrorCode
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewValidationError(message string, err error) *AppError {
	return &AppError{
		Code:    ErrCodeValidation,
		Message: message,
		Err:     err,
	}
}

func NewConfigLoadError(message string, err error) *AppError {
	return &AppError{
		Code:    ErrCodeConfigLoad,
		Message: message,
		Err:     err,
	}
}

func NewTemplateRenderError(message string, err error) *AppError {
	return &AppError{
		Code:    ErrCodeTemplateRender,
		Message: message,
		Err:     err,
	}
}

func NewProxyRequestError(message string, err error) *AppError {
	return &AppError{
		Code:    ErrCodeProxyRequest,
		Message: message,
		Err:     err,
	}
}

func NewServerStartError(message string, err error) *AppError {
	return &AppError{
		Code:    ErrCodeServerStart,
		Message: message,
		Err:     err,
	}
}

func NewBehaviorInitError(message string, err error) *AppError {
	return &AppError{
		Code:    ErrCodeBehaviorInit,
		Message: message,
		Err:     err,
	}
}
