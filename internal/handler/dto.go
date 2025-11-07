package handler

import (
	"bet/internal/domain"
)

type BetDTO struct {
	ID         string  `json:"id"`
	UserID     int64   `json:"user_id"`
	Amount     float64 `json:"amount"`
	CrashPoint float64 `json:"crash_point"`
	CreatedAt  string  `json:"created_at"`
}

func BetDTOFromDomain(bet *domain.Bet) BetDTO {
	return BetDTO{
		ID:         bet.ID,
		UserID:     bet.UserID,
		Amount:     bet.Amount,
		CrashPoint: bet.CrashPoint,
		CreatedAt:  bet.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

type ListBetsResponseDTO struct {
	Bets  []BetDTO `json:"bets"`
	Total int      `json:"total"`
	Page  int      `json:"page"`
	Limit int      `json:"limit"`
}

func ListBetsResponseDTOFromDomain(resp domain.ListBetsResponse) ListBetsResponseDTO {
	bets := make([]BetDTO, len(resp.Bets))
	for i, bet := range resp.Bets {
		bets[i] = BetDTOFromDomain(&bet)
	}

	return ListBetsResponseDTO{
		Bets:  bets,
		Total: resp.Total,
		Page:  resp.Page,
		Limit: resp.Limit,
	}
}
