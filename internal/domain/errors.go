package domain

import (
	"errors"
	"fmt"
)

var (
	ErrBetNotFound = &NotFoundError{Resource: "bet", Message: "bet not found"}
)

type RepositoryError struct {
	Message string
	Op      string
	Err     error
}

func (e *RepositoryError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Op, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Op, e.Message)
}

func (e *RepositoryError) Unwrap() error {
	return e.Err
}

func NewRepositoryError(op, message string, err error) *RepositoryError {
	return &RepositoryError{
		Op:      op,
		Message: message,
		Err:     err,
	}
}

func IsRepositoryError(err error) bool {
	var repoErr *RepositoryError
	return errors.As(err, &repoErr)
}

type NotFoundError struct {
	Resource string
	Message  string
}

func (e *NotFoundError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("%s not found", e.Resource)
}

func IsNotFoundError(err error) bool {
	var notFoundErr *NotFoundError
	return errors.As(err, &notFoundErr)
}

type InvalidInputError struct {
	Message string
}

func (e *InvalidInputError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "invalid input"
}

func IsInvalidInputError(err error) bool {
	var invalidInputErr *InvalidInputError
	return errors.As(err, &invalidInputErr)
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

func IsValidationError(err error) bool {
	var validationErr *ValidationError
	return errors.As(err, &validationErr)
}
