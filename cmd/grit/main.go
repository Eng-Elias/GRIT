package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/grit-app/grit/internal/ai"
	"github.com/grit-app/grit/internal/analysis/blame"
	"github.com/grit-app/grit/internal/analysis/churn"
	"github.com/grit-app/grit/internal/analysis/complexity"
	"github.com/grit-app/grit/internal/analysis/core"
	"github.com/grit-app/grit/internal/cache"
	"github.com/grit-app/grit/internal/clone"
	"github.com/grit-app/grit/internal/config"
	"github.com/grit-app/grit/internal/handler"
	"github.com/grit-app/grit/internal/job"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	redisCache, err := cache.New(cfg.RedisURL)
	if err != nil {
		slog.Error("failed to connect to Redis", "error", err)
		os.Exit(1)
	}
	defer redisCache.Close()
	slog.Info("connected to Redis", "url", cfg.RedisURL)

	nc, err := nats.Connect(cfg.NATSURL)
	if err != nil {
		slog.Error("failed to connect to NATS", "error", err)
		os.Exit(1)
	}
	defer nc.Close()
	slog.Info("connected to NATS", "url", cfg.NATSURL)

	js, err := nc.JetStream()
	if err != nil {
		slog.Error("failed to create JetStream context", "error", err)
		os.Exit(1)
	}

	if err := job.EnsureStream(js); err != nil {
		slog.Error("failed to ensure NATS stream", "error", err)
		os.Exit(1)
	}

	analyzer := core.NewAnalyzer(cfg.CloneDir, cfg.CloneSizeThresholdKB)
	publisher := job.NewPublisher(js, redisCache)
	worker := job.NewWorker(js, analyzer, redisCache, publisher)

	complexityAnalyzer := complexity.NewAnalyzer()
	complexityWorker := job.NewComplexityWorker(js, complexityAnalyzer, redisCache, cfg.CloneDir, publisher)

	churnAnalyzer := churn.NewAnalyzer()
	churnWorker := job.NewChurnWorker(js, churnAnalyzer, redisCache, cfg.CloneDir)

	blameAnalyzer := blame.NewAnalyzer()
	blameWorker := job.NewBlameWorker(js, blameAnalyzer, redisCache, cfg.CloneDir)

	temporalWorker := job.NewTemporalWorker(js, redisCache, cfg.CloneDir)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := worker.Start(ctx); err != nil {
		slog.Error("failed to start worker", "error", err)
		os.Exit(1)
	}
	slog.Info("core worker started")

	if err := complexityWorker.Start(ctx); err != nil {
		slog.Error("failed to start complexity worker", "error", err)
		os.Exit(1)
	}
	slog.Info("complexity worker started")

	if err := churnWorker.Start(ctx); err != nil {
		slog.Error("failed to start churn worker", "error", err)
		os.Exit(1)
	}
	slog.Info("churn worker started")

	if err := blameWorker.Start(ctx); err != nil {
		slog.Error("failed to start blame worker", "error", err)
		os.Exit(1)
	}
	slog.Info("blame worker started")

	if err := temporalWorker.Start(ctx); err != nil {
		slog.Error("failed to start temporal worker", "error", err)
		os.Exit(1)
	}
	slog.Info("temporal worker started")

	clone.StartCleanup(ctx, cfg.CloneDir, 1*time.Hour, 10*time.Minute)
	slog.Info("clone cleanup goroutine started")

	// Initialize Gemini AI client (nil if GEMINI_API_KEY not set).
	var aiClient *ai.Client
	if cfg.GeminiAPIKey != "" {
		aiClient, err = ai.NewClient(ctx, cfg.GeminiAPIKey)
		if err != nil {
			slog.Warn("AI features disabled", "error", err)
		} else {
			slog.Info("Gemini AI client initialized")
		}
	} else {
		slog.Info("GEMINI_API_KEY not set, AI features disabled")
	}

	analysisHandler := handler.NewAnalysisHandler(redisCache, publisher, cfg.GitHubToken)
	statusHandler := handler.NewStatusHandler(redisCache)
	cacheHandler := handler.NewCacheHandler(redisCache)
	badgeHandler := handler.NewBadgeHandler(redisCache)
	complexityHandler := handler.NewComplexityHandler(redisCache)
	churnHandler := handler.NewChurnHandler(redisCache)
	contributorHandler := handler.NewContributorHandler(redisCache)
	temporalHandler := handler.NewTemporalHandler(redisCache)
	aiSummaryHandler := handler.NewAISummaryHandler(redisCache, aiClient, cfg.CloneDir)
	chatRateLimiter := ai.NewRateLimiter()
	defer chatRateLimiter.Stop()
	aiChatHandler := handler.NewAIChatHandler(redisCache, aiClient, chatRateLimiter, cfg.CloneDir)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Handle("/metrics", promhttp.Handler())

	r.Route("/api/{owner}/{repo}", func(r chi.Router) {
		r.Get("/", analysisHandler.HandleAnalysis)
		r.Get("/status", statusHandler.HandleStatus)
		r.Get("/badge", badgeHandler.HandleBadge)
		r.Get("/complexity", complexityHandler.HandleComplexity)
		r.Get("/churn-matrix", churnHandler.HandleChurnMatrix)
		r.Get("/contributors", contributorHandler.HandleContributors)
		r.Get("/contributors/files", contributorHandler.HandleContributorFiles)
		r.Get("/temporal", temporalHandler.HandleTemporal)
		r.Post("/ai/summary", aiSummaryHandler.HandleAISummary)
		r.Post("/ai/chat", aiChatHandler.HandleAIChat)
		r.Delete("/cache", cacheHandler.HandleDeleteCache)
	})

	addr := fmt.Sprintf(":%d", cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server starting", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "error", err)
	}

	slog.Info("server stopped")
}
