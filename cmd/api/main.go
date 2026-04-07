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

	infrastructurepostgres "example.com/taskservice/internal/infrastructure/postgres"
	postgresrepo "example.com/taskservice/internal/repository/postgres"
	transporthttp "example.com/taskservice/internal/transport/http"
	swaggerdocs "example.com/taskservice/internal/transport/http/docs"
	httphandlers "example.com/taskservice/internal/transport/http/handlers"
	"example.com/taskservice/internal/usecase/task"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg := loadConfig()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := infrastructurepostgres.Open(ctx, cfg.DatabaseDSN)
	if err != nil {
		logger.Error("open postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// ========== СУЩЕСТВУЮЩИЕ КОМПОНЕНТЫ ДЛЯ ОБЫЧНЫХ ЗАДАЧ ==========
	taskRepo := postgresrepo.New(pool)
	taskUsecase := task.NewService(taskRepo)
	taskHandler := httphandlers.NewTaskHandler(taskUsecase)

	// ========== НОВЫЕ КОМПОНЕНТЫ ДЛЯ ПЕРИОДИЧЕСКИХ ЗАДАЧ ==========

	// Репозитории
	templateRepo := postgresrepo.NewTemplateRepository(pool)
	instanceRepo := postgresrepo.NewInstanceRepository(pool)

	// Сервис для периодических задач
	recurringUsecase := task.NewRecurringService(templateRepo, instanceRepo)

	// Хендлер для периодических задач
	recurringHandler := httphandlers.NewRecurringHandler(recurringUsecase)

	// Документация Swagger
	docsHandler := swaggerdocs.NewHandler()

	// Роутер с обоими хендлерами
	router := transporthttp.NewRouter(taskHandler, recurringHandler, docsHandler)

	// ========== ЗАПУСК ФОНОВОЙ ГОРУТИНЫ ДЛЯ ГЕНЕРАЦИИ ЗАДАЧ ==========
	// Запускаем генерацию экземпляров на 30 дней вперед
	go startRecurrenceScheduler(ctx, recurringUsecase, logger)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown http server", "error", err)
		}
	}()

	logger.Info("http server started", "addr", cfg.HTTPAddr)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("listen and serve", "error", err)
		os.Exit(1)
	}
}

// startRecurrenceScheduler - фоновая задача для генерации экземпляров
func startRecurrenceScheduler(ctx context.Context, recurringUsecase task.RecurringUsecase, logger *slog.Logger) {
	// Первый запуск сразу после старта
	if err := recurringUsecase.GenerateFutureInstances(ctx, 30); err != nil {
		logger.Error("initial generation failed", "error", err)
	}

	// Помечаем просроченные задачи при старте
	if err := recurringUsecase.MarkOverdue(ctx); err != nil {
		logger.Error("initial mark overdue failed", "error", err)
	}

	// Затем каждый день в полночь
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			logger.Info("running scheduled generation of task instances")
			if err := recurringUsecase.GenerateFutureInstances(ctx, 30); err != nil {
				logger.Error("scheduled generation failed", "error", err)
			}

			// Помечаем просроченные задачи
			if err := recurringUsecase.MarkOverdue(ctx); err != nil {
				logger.Error("mark overdue failed", "error", err)
			}
		case <-ctx.Done():
			logger.Info("recurrence scheduler stopped")
			return
		}
	}
}

type config struct {
	HTTPAddr    string
	DatabaseDSN string
}

func loadConfig() config {
	cfg := config{
		HTTPAddr:    envOrDefault("HTTP_ADDR", ":8080"),
		DatabaseDSN: envOrDefault("DATABASE_DSN", "postgres://postgres:postgres@localhost:5432/taskservice?sslmode=disable"),
	}

	if cfg.DatabaseDSN == "" {
		panic(fmt.Errorf("DATABASE_DSN is required"))
	}

	return cfg
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
