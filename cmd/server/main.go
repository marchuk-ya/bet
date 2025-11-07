package main

import (
	"bet/configs"
	"bet/internal/handler"
	"bet/internal/middleware"
	"bet/internal/repository"
	"bet/internal/service"
	"bet/internal/validator"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func main() {
	logger := initLogger()
	defer logger.Sync()

	cfg := loadConfig(logger)

	betRepo := repository.NewInMemoryBetRepository()
	betService := service.NewBetService(betRepo)
	betHandler := handler.NewBetHandler(betService, validator.NewBetValidator(), logger)
	healthHandler := handler.NewHealthHandler(logger, betRepo)
	rateLimiter := middleware.NewRateLimiter(middleware.RateLimitConfig{
		RequestsPerMinute: cfg.RateLimit.RequestsPerMinute,
		Logger:            logger,
	})

	srv := setupServer(cfg, betHandler, healthHandler, rateLimiter, logger)

	startServer(srv, cfg, logger)
	shutdownServer(srv, rateLimiter, logger)
}

func initLogger() *zap.Logger {
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	return logger
}

func loadConfig(logger *zap.Logger) *configs.Config {
	cfg, err := configs.LoadConfig()
	if err != nil {
		logger.Fatal("invalid configuration", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("configuration loaded successfully",
		zap.Int("port", cfg.Server.Port),
		zap.Int("read_timeout", cfg.Server.ReadTimeout),
		zap.Int("write_timeout", cfg.Server.WriteTimeout),
		zap.Int("idle_timeout", cfg.Server.IdleTimeout),
		zap.Int("rate_limit", cfg.RateLimit.RequestsPerMinute),
	)

	return cfg
}

func setupServer(cfg *configs.Config, betHandler *handler.BetHandler, healthHandler *handler.HealthHandler, rateLimiter *middleware.RateLimiter, logger *zap.Logger) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", healthHandler.Health)
	mux.HandleFunc("GET /ready", healthHandler.Ready)
	mux.HandleFunc("GET /live", healthHandler.Live)

	mux.HandleFunc("POST /bets", betHandler.CreateBet)
	mux.HandleFunc("GET /bets", betHandler.ListBets)
	mux.HandleFunc("GET /bets/{id}", betHandler.GetBet)

	httpHandler := middleware.RequestIDMiddleware(mux)
	httpHandler = middleware.RateLimitMiddleware(rateLimiter)(httpHandler)
	httpHandler = middleware.LoggingMiddleware(logger)(httpHandler)

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      httpHandler,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}
}

func startServer(srv *http.Server, cfg *configs.Config, logger *zap.Logger) {
	go func() {
		logger.Info("starting server", zap.Int("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")
}

func shutdownServer(srv *http.Server, rateLimiter *middleware.RateLimiter, logger *zap.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := rateLimiter.Shutdown(ctx); err != nil {
		logger.Warn("rate limiter shutdown error", zap.Error(err))
	}

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("server exited")
}
