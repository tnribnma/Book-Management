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
	"book-management/repository"
	"book-management/service"
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

	userRepo := repository.NewUserRepository(db)
	bookRepo := repository.NewBookRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	borrowingRepo := repository.NewBorrowingRepository(db)
	reservationRepo := repository.NewReservationRepository(db)
	reportRepo := repository.NewReportRepository(db)

	authService := service.NewUserService(userRepo)
	userService := service.NewUserService(userRepo)
	bookService := service.NewBookService(bookRepo)
	categoryService := service.NewCategoryService(categoryRepo, bookRepo)
	borrowingService := service.NewBorrowingService(borrowingRepo, bookRepo)
	reservationService := service.NewReservationService(reservationRepo, bookRepo)
	reportService := service.NewReportService(reportRepo, bookRepo, borrowingRepo)

	authHandler := handlers.NewAuthHandler(authService)
	bookHandler := handlers.NewBookHandler(bookService)
	userHandler := handlers.NewUserHandler(userService)
	categoryHandler := handlers.NewCategoryHandler(categoryService)
	borrowingHandler := handlers.NewBorrowingHandler(borrowingService)
	reservationHandler := handlers.NewReservationHandler(reservationService)
	reportHandler := handlers.NewReportHandler(reportService)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Book Management API is running"))
	})
	mux.Handle("GET /health", http.HandlerFunc(handlers.HealthCheck))
	mux.Handle("POST /users/register", http.HandlerFunc(authHandler.Register))
	mux.Handle("POST /auth/login", http.HandlerFunc(authHandler.Login))

	mux.Handle("GET /books", middleware.Auth(http.HandlerFunc(bookHandler.ListBooks)))
	mux.Handle("GET /books/{id}", middleware.Auth(http.HandlerFunc(bookHandler.GetBook)))

	mux.Handle("POST /books",
		middleware.Auth(
			middleware.Role("librarian", "admin")(
				http.HandlerFunc(bookHandler.CreateBook),
			),
		),
	)

	mux.Handle("PUT /books/{id}",
		middleware.Auth(
			middleware.Role("librarian", "admin")(
				http.HandlerFunc(bookHandler.UpdateBook),
			),
		),
	)

	mux.Handle("DELETE /books/{id}",
		middleware.Auth(
			middleware.Role("admin")(
				http.HandlerFunc(bookHandler.DeleteBook),
			),
		),
	)

	mux.Handle("GET /categories",
		middleware.Auth(http.HandlerFunc(categoryHandler.List)),
	)

	mux.Handle("POST /categories",
		middleware.Auth(
			middleware.Role("librarian", "admin")(
				http.HandlerFunc(categoryHandler.Create),
			),
		),
	)

	mux.Handle("POST /borrow",
		middleware.Auth(http.HandlerFunc(borrowingHandler.IssueBook)),
	)

	mux.Handle("POST /return",
		middleware.Auth(http.HandlerFunc(borrowingHandler.ReturnBook)),
	)

	mux.Handle("GET /my-borrowings",
		middleware.Auth(http.HandlerFunc(borrowingHandler.GetMyBorrowings)),
	)

	mux.Handle("POST /reserve",
		middleware.Auth(http.HandlerFunc(reservationHandler.Create)),
	)

	mux.Handle("GET /profile",
		middleware.Auth(http.HandlerFunc(userHandler.GetProfile)),
	)

	mux.Handle("GET /admin/users",
		middleware.Auth(
			middleware.Role("admin")(
				http.HandlerFunc(userHandler.ListUsers),
			),
		),
	)

	mux.Handle("PUT /admin/users/{id}/role",
		middleware.Auth(
			middleware.Role("admin")(
				http.HandlerFunc(userHandler.UpdateRole),
			),
		),
	)

	mux.Handle("GET /reports/dashboard",
		middleware.Auth(
			middleware.Role("admin", "librarian")(
				http.HandlerFunc(reportHandler.GetDashboard),
			),
		),
	)

	mux.Handle("GET /my-reservations",
		middleware.Auth(http.HandlerFunc(reservationHandler.GetMyReservations)),
	)
 
	mux.Handle("DELETE /reservations/{id}",
		middleware.Auth(http.HandlerFunc(reservationHandler.Cancel)),
	)
 
	mux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir("public"))))

	srv := &http.Server{
	Addr:         ":" + cfg.Server.Port,
	Handler:      middleware.CORS(mux),
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