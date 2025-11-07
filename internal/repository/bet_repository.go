package repository

import (
	"bet/internal/domain"
	"context"
)

type BetRepository interface {
	Create(ctx context.Context, bet *domain.Bet) error
	GetByID(ctx context.Context, id string) (*domain.Bet, error)
	List(ctx context.Context, req domain.ListBetsRequest) (domain.ListBetsResponse, error)
	HealthCheck(ctx context.Context) error
}
