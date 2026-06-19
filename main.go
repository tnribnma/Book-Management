package main

import (
	"log"
	"net/http"
	"github.com/joho/godotenv"
	"book-management/config"
	"book-management/handlers"
	"book-management/middleware"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}
	cfg := config.NewDBConfig()

	db, err := config.OpenDB(cfg.ConnectionString())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	authHandler := handlers.NewAuthHandler(db)
	bookHandler := handlers.NewBookHandler(db)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /users/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)

	mux.Handle("GET /books", middleware.Auth(http.HandlerFunc(bookHandler.ListBooks)))
	mux.Handle("POST /books", middleware.Auth(http.HandlerFunc(bookHandler.CreateBook)))
	mux.Handle("GET /books/{id}", middleware.Auth(http.HandlerFunc(bookHandler.GetBook)))
	mux.Handle("PUT /books/{id}", middleware.Auth(http.HandlerFunc(bookHandler.UpdateBook)))
	mux.Handle("DELETE /books/{id}", middleware.Auth(http.HandlerFunc(bookHandler.DeleteBook)))

	log.Println("server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
