package service

import (
	"bet/internal/domain"
	"bet/internal/repository"
	"context"
	"fmt"
)

type BetServiceUseCase interface {
	CreateBet(ctx context.Context, userID int64, amount, crashPoint float64) (*domain.Bet, error)
	GetBetByID(ctx context.Context, id string) (*domain.Bet, error)
	ListBets(ctx context.Context, req domain.ListBetsRequest) (domain.ListBetsResponse, error)
}

type BetService struct {
	repo repository.BetRepository
}

func NewBetService(repo repository.BetRepository) *BetService {
	return &BetService{
		repo: repo,
	}
}

func (s *BetService) CreateBet(ctx context.Context, userID int64, amount, crashPoint float64) (*domain.Bet, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	bet := domain.NewBet(userID, amount, crashPoint)

	if err := s.repo.Create(ctx, bet); err != nil {
		return nil, domain.NewRepositoryError("CreateBet", "failed to create bet", err)
	}

	return bet, nil
}

func (s *BetService) GetBetByID(ctx context.Context, id string) (*domain.Bet, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	bet, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, domain.NewRepositoryError("GetBetByID", fmt.Sprintf("failed to get bet by id %s", id), err)
	}

	return bet, nil
}

func (s *BetService) ListBets(ctx context.Context, req domain.ListBetsRequest) (domain.ListBetsResponse, error) {
	if ctx.Err() != nil {
		return domain.ListBetsResponse{}, ctx.Err()
	}

	if req.Sort.SortBy == "" {
		req.Sort.SortBy = "created_at"
	}
	if req.Sort.Order == "" {
		req.Sort.Order = "desc"
	}

	response, err := s.repo.List(ctx, req)
	if err != nil {
		return domain.ListBetsResponse{}, domain.NewRepositoryError("ListBets", "failed to list bets", err)
	}

	return response, nil
}
