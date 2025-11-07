package domain

import (
	"time"
	"github.com/google/uuid"
)

type Bet struct {
	ID         string
	UserID     int64
	Amount     float64
	CrashPoint float64
	CreatedAt  time.Time
}

func NewBet(userID int64, amount, crashPoint float64) *Bet {
	return &Bet{
		ID:         uuid.New().String(),
		UserID:     userID,
		Amount:     amount,
		CrashPoint: crashPoint,
		CreatedAt:  time.Now(),
	}
}

type BetFilters struct {
	UserID    *int64
	MinAmount *float64
	MaxAmount *float64
}

type PaginationParams struct {
	Page  int
	Limit int
}

type SortParams struct {
	SortBy string
	Order  string
}

type ListBetsRequest struct {
	Filters   BetFilters
	Pagination PaginationParams
	Sort      SortParams
}

type ListBetsResponse struct {
	Bets  []Bet
	Total int
	Page  int
	Limit int
}

