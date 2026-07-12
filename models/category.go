package models

type Category struct {
	ID          int64  `json:"id"`
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description,omitempty" validate:"omitempty,max=500"`
}