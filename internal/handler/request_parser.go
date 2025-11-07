package handler

import (
	"bet/internal/domain"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var dangerousCharsRegex = regexp.MustCompile(`[<>\"'%;)(&+]`)

func sanitizeQueryParam(s string) string {
	s = strings.TrimSpace(s)
	s = dangerousCharsRegex.ReplaceAllString(s, "")
	return s
}

func validateQueryParam(s string, allowedValues []string) bool {
	if len(allowedValues) == 0 {
		return true
	}
	s = strings.ToLower(strings.TrimSpace(s))
	for _, allowed := range allowedValues {
		if strings.ToLower(allowed) == s {
			return true
		}
	}
	return false
}

func ParseListBetsRequest(r *http.Request) domain.ListBetsRequest {
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		pageStr = sanitizeQueryParam(pageStr)
		if parsed, err := strconv.Atoi(pageStr); err == nil && parsed > 0 && parsed <= 10000 {
			page = parsed
		}
	}

	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limitStr = sanitizeQueryParam(limitStr)
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	var filters domain.BetFilters

	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		userIDStr = sanitizeQueryParam(userIDStr)
		if userID, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
			filters.UserID = &userID
		}
	}

	if minAmountStr := r.URL.Query().Get("min_amount"); minAmountStr != "" {
		minAmountStr = sanitizeQueryParam(minAmountStr)
		if minAmount, err := strconv.ParseFloat(minAmountStr, 64); err == nil {
			filters.MinAmount = &minAmount
		}
	}

	if maxAmountStr := r.URL.Query().Get("max_amount"); maxAmountStr != "" {
		maxAmountStr = sanitizeQueryParam(maxAmountStr)
		if maxAmount, err := strconv.ParseFloat(maxAmountStr, 64); err == nil {
			filters.MaxAmount = &maxAmount
		}
	}

	sortBy := r.URL.Query().Get("sort_by")
	if sortBy == "" {
		sortBy = "created_at"
	} else {
		sortBy = sanitizeQueryParam(sortBy)
		allowedSortBy := []string{"amount", "created_at"}
		if !validateQueryParam(sortBy, allowedSortBy) {
			sortBy = "created_at"
		}
	}

	order := r.URL.Query().Get("order")
	if order == "" {
		order = "desc"
	} else {
		order = sanitizeQueryParam(order)
		allowedOrder := []string{"asc", "desc"}
		if !validateQueryParam(order, allowedOrder) {
			order = "desc"
		}
	}

	return domain.ListBetsRequest{
		Filters: filters,
		Pagination: domain.PaginationParams{
			Page:  page,
			Limit: limit,
		},
		Sort: domain.SortParams{
			SortBy: sortBy,
			Order:  order,
		},
	}
}
