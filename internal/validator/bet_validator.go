package validator

import (
	"bet/internal/domain"
	"fmt"
	"math"
	"regexp"
)

type BetValidator interface {
	ValidateCreateRequest(userID int64, amount, crashPoint float64) error
	ValidateUserID(userID int64) error
	ValidateAmount(amount float64) error
	ValidateCrashPoint(crashPoint float64) error
	ValidatePagination(page, limit int) error
	ValidateSort(sortBy, order string) error
	ValidateBetID(id string) error
	ValidateNumber(value float64) error
}

const (
	MinAmount      = 1.0
	MaxAmount      = 100000.0
	MinCrashPoint  = 1.0
	MaxCrashPoint  = 100.0
	MinUserID      = 1
	MaxUserID      = 999999999
	MaxBetIDLength = 36
)

var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type betValidator struct{}

func NewBetValidator() BetValidator {
	return &betValidator{}
}

func (v *betValidator) ValidateCreateRequest(userID int64, amount, crashPoint float64) error {
	if err := v.ValidateUserID(userID); err != nil {
		return err
	}

	if err := v.ValidateAmount(amount); err != nil {
		return err
	}

	if err := v.ValidateCrashPoint(crashPoint); err != nil {
		return err
	}

	return nil
}

func (v *betValidator) ValidateUserID(userID int64) error {
	if userID < MinUserID || userID > MaxUserID {
		return &domain.ValidationError{
			Field:   "user_id",
			Message: fmt.Sprintf("user_id must be between %d and %d", MinUserID, MaxUserID),
		}
	}
	return nil
}

func (v *betValidator) ValidateAmount(amount float64) error {
	if err := v.ValidateNumber(amount); err != nil {
		return &domain.ValidationError{
			Field:   "amount",
			Message: "amount must be a valid numeric value",
		}
	}

	if amount < MinAmount {
		return &domain.ValidationError{
			Field:   "amount",
			Message: fmt.Sprintf("amount must be at least %.2f", MinAmount),
		}
	}

	if amount > MaxAmount {
		return &domain.ValidationError{
			Field:   "amount",
			Message: fmt.Sprintf("amount must not exceed %.2f", MaxAmount),
		}
	}

	return nil
}

func (v *betValidator) ValidateCrashPoint(crashPoint float64) error {
	if err := v.ValidateNumber(crashPoint); err != nil {
		return &domain.ValidationError{
			Field:   "crash_point",
			Message: "crash_point must be a valid numeric value",
		}
	}

	if crashPoint < MinCrashPoint {
		return &domain.ValidationError{
			Field:   "crash_point",
			Message: fmt.Sprintf("crash_point must be at least %.2f", MinCrashPoint),
		}
	}

	if crashPoint > MaxCrashPoint {
		return &domain.ValidationError{
			Field:   "crash_point",
			Message: fmt.Sprintf("crash_point must not exceed %.2f", MaxCrashPoint),
		}
	}

	return nil
}

func (v *betValidator) ValidatePagination(page, limit int) error {
	if page < 1 {
		return &domain.ValidationError{
			Field:   "page",
			Message: "page must be at least 1",
		}
	}

	if limit < 1 {
		return &domain.ValidationError{
			Field:   "limit",
			Message: "limit must be at least 1",
		}
	}

	if limit > 100 {
		return &domain.ValidationError{
			Field:   "limit",
			Message: "limit must not exceed 100",
		}
	}

	return nil
}

func (v *betValidator) ValidateSort(sortBy, order string) error {
	if sortBy != "" && sortBy != "amount" && sortBy != "created_at" {
		return &domain.ValidationError{
			Field:   "sort_by",
			Message: "sort_by must be either 'amount' or 'created_at'",
		}
	}

	if order != "" && order != "asc" && order != "desc" {
		return &domain.ValidationError{
			Field:   "order",
			Message: "order must be either 'asc' or 'desc'",
		}
	}

	return nil
}

func (v *betValidator) ValidateBetID(id string) error {
	if id == "" {
		return &domain.ValidationError{
			Field:   "id",
			Message: "bet id is required",
		}
	}

	if len(id) > MaxBetIDLength {
		return &domain.ValidationError{
			Field:   "id",
			Message: fmt.Sprintf("bet id must not exceed %d characters", MaxBetIDLength),
		}
	}

	if !uuidRegex.MatchString(id) {
		return &domain.ValidationError{
			Field:   "id",
			Message: "bet id must be a valid UUID format",
		}
	}

	return nil
}

func (v *betValidator) ValidateNumber(value float64) error {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return &domain.ValidationError{
			Field:   "number",
			Message: "number must be a valid numeric value",
		}
	}
	return nil
}
