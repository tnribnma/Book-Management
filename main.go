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
	"book-management/validators"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	utils.SetJWTSecret(cfg.JWT.Secret)

	validators.Init()

	db, err := config.OpenDB(cfg.DB.ConnectionString())
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer db.Close()

	authHandler := handlers.NewAuthHandler(db)
	bookHandler := handlers.NewBookHandler(db)
	userHandler := handlers.NewUserHandler(db)
	categoryHandler := handlers.NewCategoryHandler(db)
	borrowingHandler := handlers.NewBorrowingHandler(db)
	reportHandler := handlers.NewReportHandler(db)
	reservationHandler := handlers.NewReservationHandler(db)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handlers.HealthCheck)
	mux.HandleFunc("POST /users/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)

	mux.Handle("GET /books", middleware.Auth(middleware.CORS(bookHandler.ListBooks)))
	mux.Handle("GET /books/{id}", middleware.Auth(middleware.CORS(bookHandler.GetBook)))
	mux.Handle("POST /books", middleware.Auth(
		middleware.CORS(
			middleware.Role("librarian", "admin")(bookHandler.CreateBook),
		),
	))
	mux.Handle("PUT /books/{id}", middleware.Auth(
		middleware.CORS(
			middleware.Role("librarian", "admin")(bookHandler.UpdateBook),
		),
	))
	mux.Handle("DELETE /books/{id}", middleware.Auth(
		middleware.CORS(
			middleware.Role("admin")(bookHandler.DeleteBook),
		),
	))

	mux.Handle("GET /categories", middleware.Auth(middleware.CORS(categoryHandler.List)))
	mux.Handle("POST /categories", middleware.Auth(
		middleware.CORS(
			middleware.Role("librarian", "admin")(categoryHandler.Create),
		),
	))

	mux.Handle("POST /borrow", middleware.Auth(middleware.CORS(borrowingHandler.IssueBook)))
	mux.Handle("POST /return", middleware.Auth(middleware.CORS(borrowingHandler.ReturnBook)))
	mux.Handle("GET /my-borrowings", middleware.Auth(middleware.CORS(borrowingHandler.GetMyBorrowings)))

	mux.Handle("GET /profile", middleware.Auth(middleware.CORS(userHandler.GetProfile)))

	mux.Handle("GET /admin/users", middleware.Auth(
		middleware.CORS(
			middleware.Role("admin")(userHandler.ListUsers),
		),
	))

	mux.Handle("GET /reports/dashboard", middleware.Auth(
		middleware.CORS(
			middleware.Role("admin", "librarian")(reportHandler.GetDashboard),
		),
	))

	mux.Handle("POST /reserve", middleware.Auth(middleware.CORS(reservationHandler.Create)))

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