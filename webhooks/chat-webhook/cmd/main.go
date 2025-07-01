// webhooks/chat-webhook/cmd/main.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
	"github.com/go-redis/redis/v8"

	"webhooks/chat-webhook/internal/facebook"
	"webhooks/chat-webhook/internal/line"
	"webhooks/chat-webhook/internal/router"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Get configuration from environment
	port := getEnv("PORT", "8094")
	kafkaBrokers := getEnv("KAFKA_BROKERS", "kafka:9092")
	redisAddr := getEnv("REDIS_ADDR", "redis:6379")

	// Facebook configuration
	fbVerifyToken := getEnv("FACEBOOK_VERIFY_TOKEN", "")
	fbPageAccessToken := getEnv("FACEBOOK_PAGE_ACCESS_TOKEN", "")
	fbAppSecret := getEnv("FACEBOOK_APP_SECRET", "")

	// LINE configuration
	lineChannelSecret := getEnv("LINE_CHANNEL_SECRET", "")
	lineChannelAccessToken := getEnv("LINE_CHANNEL_ACCESS_TOKEN", "")

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Initialize Kafka writer
	kafkaWriter := &kafka.Writer{
		Addr:         kafka.TCP(kafkaBrokers),
		Topic:        "chat-messages",
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
	}
	defer kafkaWriter.Close()

	// Initialize handlers
	facebookHandler := facebook.NewHandler(fbVerifyToken, fbPageAccessToken, fbAppSecret, kafkaWriter, redisClient)
	lineHandler := line.NewHandler(lineChannelSecret, lineChannelAccessToken, kafkaWriter, redisClient)

	// Setup routes
	mainRouter := mux.NewRouter()
	chatRouter := router.NewChatRouter(facebookHandler, lineHandler)
	chatRouter.RegisterRoutes(mainRouter)

	// Health and readiness checks
	mainRouter.HandleFunc("/health", healthCheckHandler).Methods("GET")
	mainRouter.HandleFunc("/ready", readinessHandler(redisClient, kafkaWriter)).Methods("GET")

	// Setup server
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mainRouter,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Chat webhook service starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func readinessHandler(redisClient *redis.Client, kafkaWriter *kafka.Writer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		
		// Check Redis connection
		if err := redisClient.Ping(ctx).Err(); err != nil {
			log.Printf("Redis health check failed: %v", err)
			http.Error(w, "Redis not ready", http.StatusServiceUnavailable)
			return
		}

		// Check Kafka connection
		conn, err := kafka.Dial("tcp", "kafka:9092")
		if err != nil {
			log.Printf("Kafka health check failed: %v", err)
			http.Error(w, "Kafka not ready", http.StatusServiceUnavailable)
			return
		}
		conn.Close()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ready"))
	}
}
