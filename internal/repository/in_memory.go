package repository

import (
	"bet/internal/domain"
	"context"
	"sort"
	"sync"
)

type inMemoryBetRepository struct {
	bets        map[string]*domain.Bet
	mu          sync.RWMutex
	userIDIndex map[int64][]string
	muIndex     sync.RWMutex
}

func NewInMemoryBetRepository() BetRepository {
	return &inMemoryBetRepository{
		bets:        make(map[string]*domain.Bet),
		userIDIndex: make(map[int64][]string),
	}
}

func (r *inMemoryBetRepository) Create(ctx context.Context, bet *domain.Bet) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.bets[bet.ID] = bet

	r.muIndex.Lock()
	r.userIDIndex[bet.UserID] = append(r.userIDIndex[bet.UserID], bet.ID)
	r.muIndex.Unlock()

	return nil
}

func (r *inMemoryBetRepository) GetByID(ctx context.Context, id string) (*domain.Bet, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	bet, exists := r.bets[id]
	if !exists {
		return nil, domain.ErrBetNotFound
	}

	betCopy := *bet
	return &betCopy, nil
}

func (r *inMemoryBetRepository) List(ctx context.Context, req domain.ListBetsRequest) (domain.ListBetsResponse, error) {
	if ctx.Err() != nil {
		return domain.ListBetsResponse{}, ctx.Err()
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	var betsToCheck []*domain.Bet

	if req.Filters.UserID != nil {
		r.muIndex.RLock()
		betIDs, exists := r.userIDIndex[*req.Filters.UserID]
		r.muIndex.RUnlock()

		if !exists {
			return domain.ListBetsResponse{
				Bets:  []domain.Bet{},
				Total: 0,
				Page:  req.Pagination.Page,
				Limit: req.Pagination.Limit,
			}, nil
		}

		betsToCheck = make([]*domain.Bet, 0, len(betIDs))
		for _, betID := range betIDs {
			if bet, exists := r.bets[betID]; exists {
				betsToCheck = append(betsToCheck, bet)
			}
		}
	} else {
		betsToCheck = make([]*domain.Bet, 0, len(r.bets))
		for _, bet := range r.bets {
			betsToCheck = append(betsToCheck, bet)
		}
	}

	estimatedCapacity := len(betsToCheck) / 2
	if estimatedCapacity < 10 {
		estimatedCapacity = 10
	}
	filteredBets := make([]domain.Bet, 0, estimatedCapacity)
	for _, bet := range betsToCheck {
		if r.matchesFilter(*bet, req.Filters) {
			filteredBets = append(filteredBets, *bet)
		}
	}

	sortedBets := r.applySorting(filteredBets, req.Sort)

	total := len(sortedBets)
	paginatedBets := r.applyPagination(sortedBets, req.Pagination)

	return domain.ListBetsResponse{
		Bets:  paginatedBets,
		Total: total,
		Page:  req.Pagination.Page,
		Limit: req.Pagination.Limit,
	}, nil
}

func (r *inMemoryBetRepository) matchesFilter(bet domain.Bet, filters domain.BetFilters) bool {
	if filters.UserID == nil && filters.MinAmount == nil && filters.MaxAmount == nil {
		return true
	}

	if filters.UserID != nil && bet.UserID != *filters.UserID {
		return false
	}

	if filters.MinAmount != nil && bet.Amount < *filters.MinAmount {
		return false
	}

	if filters.MaxAmount != nil && bet.Amount > *filters.MaxAmount {
		return false
	}

	return true
}

func (r *inMemoryBetRepository) applySorting(bets []domain.Bet, sortParams domain.SortParams) []domain.Bet {
	sorted := make([]domain.Bet, len(bets))
	copy(sorted, bets)

	sort.Slice(sorted, func(i, j int) bool {
		var less bool

		switch sortParams.SortBy {
		case "amount":
			less = sorted[i].Amount < sorted[j].Amount
		case "created_at":
			less = sorted[i].CreatedAt.Before(sorted[j].CreatedAt)
		default:
			less = sorted[i].CreatedAt.Before(sorted[j].CreatedAt)
		}

		if sortParams.Order == "desc" {
			return !less
		}
		return less
	})

	return sorted
}

func (r *inMemoryBetRepository) applyPagination(bets []domain.Bet, pagination domain.PaginationParams) []domain.Bet {
	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.Limit < 1 {
		pagination.Limit = 10
	}

	start := (pagination.Page - 1) * pagination.Limit
	end := start + pagination.Limit

	if start >= len(bets) {
		return make([]domain.Bet, 0)
	}

	if end > len(bets) {
		end = len(bets)
	}

	result := make([]domain.Bet, end-start)
	copy(result, bets[start:end])
	return result
}

func (r *inMemoryBetRepository) HealthCheck(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	_ = len(r.bets)
	return nil
}
