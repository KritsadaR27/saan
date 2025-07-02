package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/saan/chat-service/internal/config"
	"github.com/saan/chat-service/internal/infrastructure/database"
	"github.com/saan/chat-service/internal/infrastructure/kafka"
	"github.com/saan/chat-service/internal/infrastructure/redis"
	"github.com/saan/chat-service/internal/infrastructure/websocket"
	httpTransport "github.com/saan/chat-service/internal/transport/http"
	"github.com/saan/chat-service/internal/application"
	"github.com/saan/chat-service/internal/domain/repository"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Setup logger
	setupLogger(cfg.LogLevel, cfg.LogFormat)

	logrus.Info("Starting Chat Service...")

	// Initialize database
	db, err := database.NewConnection(cfg.DatabaseURL)
	if err != nil {
		logrus.Fatal("Failed to connect to database: ", err)
	}

	// Auto-migrate database schema
	if err := database.AutoMigrate(db); err != nil {
		logrus.Fatal("Failed to migrate database: ", err)
	}

	// Initialize Redis
	redisClient, err := redis.NewClient(cfg.RedisAddr, cfg.RedisPassword)
	if err != nil {
		logrus.Fatal("Failed to connect to Redis: ", err)
	}

	// Initialize Kafka
	kafkaProducer, err := kafka.NewProducer(cfg.KafkaBrokers)
	if err != nil {
		logrus.Fatal("Failed to connect to Kafka: ", err)
	}
	defer kafkaProducer.Close()

	// Initialize WebSocket hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// Initialize repositories
	messageRepo := repository.NewMessageRepository(db)
	conversationRepo := repository.NewConversationRepository(db)
	userRepo := repository.NewUserRepository(db)

	// Initialize application services
	chatService := application.NewChatService(
		messageRepo,
		conversationRepo,
		userRepo,
		redisClient,
		kafkaProducer,
		wsHub,
		cfg,
	)

	// Initialize HTTP handlers
	handlers := httpTransport.NewHandlers(chatService, wsHub, cfg)

	// Setup Gin router
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	
	// Setup routes
	handlers.SetupRoutes(router)

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logrus.Infof("Chat Service listening on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatal("Failed to start server: ", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down Chat Service...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logrus.Fatal("Server forced to shutdown: ", err)
	}

	logrus.Info("Chat Service stopped")
}

func setupLogger(level, format string) {
	// Set log level
	switch level {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}

	// Set log format
	if format == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
	}
}
