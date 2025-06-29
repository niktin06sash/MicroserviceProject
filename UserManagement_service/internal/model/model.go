package model

import (
	"github.com/google/uuid"
)

type User struct {
	Id       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Password string    `json:"-"`
}
type RegistrationRequest struct {
	Name     string `json:"name" validate:"required,min=3,max=30"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=30"`
}

type AuthenticationRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=30"`
}

type DeletionRequest struct {
	Password string `json:"password" validate:"required,min=8,max=30"`
}
type UpdateRequest struct {
	Name         string `json:"name,omitempty" validate:"omitempty,min=3,max=30"`
	Email        string `json:"email,omitempty" validate:"omitempty,email"`
	LastPassword string `json:"last_password,omitempty" validate:"omitempty,min=8,max=30"`
	NewPassword  string `json:"new_password,omitempty" validate:"omitempty,min=8,max=30"`
}
