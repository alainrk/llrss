// cmd/server/main.go
package main

import (
	"context"
	"llrss/internal/handler"
	"llrss/internal/repository/json"
	"llrss/internal/service"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// dbConfig := config.NewDatabaseConfig()
	// db, err := config.InitDatabase(dbConfig)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// // Auto migrate the schema
	// if err := db.AutoMigrate(&models.Feed{}, &models.Item{}); err != nil {
	// 	log.Fatal("failed to migrate database:", err)
	// }

	// Initialize repository
	// feedRepo := repository.NewGormFeedRepository(db)

	feedRepo, err := json.NewJSONFileFeedRepository("feeds.json")
	if err != nil {
		log.Fatalf("Failed to create feed repository: %v", err)
	}

	feedService := service.NewFeedService(feedRepo)
	feedHandler := handler.NewFeedHandler(feedService)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/api/v1", func(r chi.Router) {
		feedHandler.RegisterRoutes(r)
	})

	// Create server
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// TODO: Add timeout cancellable context
	// ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	// defer cancel()

	if err := srv.Shutdown(context.TODO()); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped gracefully")
}
