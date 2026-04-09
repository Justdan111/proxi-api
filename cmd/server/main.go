package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/zap"

	"github.com/Justdan111/proxi-api/internal/activity"
	"github.com/Justdan111/proxi-api/internal/auth"
	"github.com/Justdan111/proxi-api/internal/config"
	"github.com/Justdan111/proxi-api/internal/reminder"
	"github.com/Justdan111/proxi-api/internal/user"
	"github.com/Justdan111/proxi-api/pkg/database"
	"github.com/Justdan111/proxi-api/pkg/logger"
	"github.com/Justdan111/proxi-api/pkg/response"
)

func main() {
	// Load config
	cfg := config.Load()
    logger.Init(cfg.AppEnv)
    defer logger.Sync()

    logger.Log.Info("Starting Proxi API",
        zap.String("env", cfg.AppEnv),
        zap.String("port", cfg.Port),
    )

	// Connect to MongoDB
	db := database.NewMongoDB(cfg.MongoURI, cfg.MongoDBName)
	defer db.Disconnect()

	// Wire up dependencies
	userRepo := user.NewRepository(db.DB)
	authService := auth.NewService(userRepo, cfg.JWTSecret, cfg.JWTExpiryHours)
	authHandler := auth.NewHandler(authService)

	reminderRepo := reminder.NewRepository(db.DB)
	reminderService := reminder.NewService(reminderRepo)
	reminderHandler := reminder.NewHandler(reminderService)

	activityRepo := activity.NewRepository(db.DB)
	activityService := activity.NewService(activityRepo)
	activityHandler := activity.NewHandler(activityService)

	// Set up router
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // tighten in production
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
    // Ping MongoDB
    ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
    defer cancel()

    mongoStatus := "ok"
    if err := db.Client.Ping(ctx, nil); err != nil {
        mongoStatus = "unreachable"
    }

    status := map[string]interface{}{
        "status":  "ok",
        "service": "proxi-api",
        "mongo":   mongoStatus,
        "env":     cfg.AppEnv,
    }
    response.JSON(w, http.StatusOK, status)
})

	// Auth routes (public + profile)
	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/signup", authHandler.Signup)
		r.Post("/login", authHandler.Login)
		r.Post("/logout", authHandler.Logout)

		r.Group(func(r chi.Router) {
			r.Use(authService.Middleware)
			r.Get("/me", authHandler.GetMe)
		})
	})

	// Protected API routes
	r.Group(func(r chi.Router) {
		r.Use(authService.Middleware)

		r.Route("/api/reminders", func(r chi.Router) {
			r.Get("/", reminderHandler.GetAll)
			r.Post("/", reminderHandler.Create)
			r.Get("/{id}", reminderHandler.GetOne)
			r.Put("/{id}", reminderHandler.Update)
			r.Patch("/{id}/toggle", reminderHandler.Toggle)
			r.Delete("/{id}", reminderHandler.Delete)
		})

		r.Route("/api/activities", func(r chi.Router) {
			r.Get("/", activityHandler.GetAll)
			r.Post("/", activityHandler.Log)
		})
	})

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		logger.Log.Info("Starting Proxi API",
			zap.String("addr", addr),
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Error("Server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("⏳ Shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	server.Shutdown(ctx)
	logger.Log.Info(" Server stopped")
}
