package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	_ "github.com/sainakuo/scalable-notification-system/docs"
	"github.com/sainakuo/scalable-notification-system/internal/app"
	"github.com/sainakuo/scalable-notification-system/internal/config"
	"github.com/sainakuo/scalable-notification-system/internal/handler"
	"github.com/sainakuo/scalable-notification-system/internal/tracing"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// @title Scalable Notification System API
// @version 1.0
// @description REST API for asynchronous task processing with Go, Gin, PostgreSQL, Redis, Workers and gRPC.
// @host localhost:8080
// @BasePath /
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
		"sns-api",
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

	apiApp, err := app.BuildAPI(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer apiApp.Close()

	router := gin.Default()

	router.Use(otelgin.Middleware("sns-api"))
	router.Use(handler.RequestIDMiddleware())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	handler.RegisterRoutes(router, apiApp.TaskHandler, apiApp.HealthHandler)

	server := &http.Server{
		Addr:    ":" + cfg.APIPort,
		Handler: router,
	}

	go func() {
		log.Println("API server started on port", cfg.APIPort)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()

	log.Println("Shutting down API server...")

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctxShutdown); err != nil {
		log.Fatal("API server forced to shutdown:", err)
	}

	log.Println("API server exited")
}
