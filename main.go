package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"book-management/config"
	"book-management/handlers"
	"book-management/middleware"
	"book-management/utils"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(" No .env file found, using system environment variables")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	utils.SetJWTSecret(cfg.JWT.Secret)

	db, err := config.OpenDB(cfg.DB.ConnectionString())
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer db.Close()

	authHandler := handlers.NewAuthHandler(db, &cfg.JWT)
	bookHandler := handlers.NewBookHandler(db)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handlers.HealthCheck)
	
	mux.HandleFunc("POST /users/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)

	mux.Handle("GET /books", middleware.Auth(middleware.CORS(bookHandler.ListBooks)))
	mux.Handle("POST /books", middleware.Auth(middleware.CORS(bookHandler.CreateBook)))
	mux.Handle("GET /books/{id}", middleware.Auth(middleware.CORS(bookHandler.GetBook)))
	mux.Handle("PUT /books/{id}", middleware.Auth(middleware.CORS(bookHandler.UpdateBook)))
	mux.Handle("DELETE /books/{id}", middleware.Auth(middleware.CORS(bookHandler.DeleteBook)))

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server started on http://localhost:%s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server error:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited gracefully")
}