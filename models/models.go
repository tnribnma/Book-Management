package models

type User struct{
	ID int64 `json:"id"`
	Email string `json:"email"`
	PasswordHash string `json:"password_hash"`
	CreatedAt string `json:"created_at"`
}

type Book struct{
	ID int64 `json:"id"`
	UserID int64 `json:"user_id"`
	Title string `json:"title"`
	Author string `json:"author"`
	PublishedYear int `json:"published_year"`
	CreatedAt string `json:"created_at"`
}
