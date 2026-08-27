package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sainakuo/scalable-notification-system/internal/app"
	"github.com/sainakuo/scalable-notification-system/internal/config"
	"github.com/sainakuo/scalable-notification-system/internal/tracing"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	cfg := config.LoadConfig()

	tracerProvider, err := tracing.NewTracerProvider(
		ctx,
		"sns-worker",
	)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		if err := tracerProvider.Shutdown(shutdownCtx); err != nil {
			log.Println("tracer shutdown error:", err)
		}
	}()

	workerApp, err := app.BuildWorker(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer workerApp.Close()

	metricsServer := &http.Server{
		Addr:    ":9090",
		Handler: promhttp.Handler(),
	}

	go func() {
		log.Println("worker metrics server started on :9090")

		if err := metricsServer.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			log.Println("metrics server error:", err)
		}
	}()

	log.Println("Worker started")

	if err := workerApp.TaskWorker.Run(ctx); err != nil &&
		!errors.Is(err, context.Canceled) {
		log.Fatal(err)
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := metricsServer.Shutdown(shutdownCtx); err != nil {
		log.Println("metrics server shutdown error:", err)
	}

	log.Println("Worker stopped")
}
