package model

import (
	"github.com/google/uuid"
)

type User struct {
	Id       uuid.UUID
	Name     string
	Email    string
	Password string
}
type RegistrationRequest struct {
	Name     string `json:"name" validate:"required,min=3"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type AuthenticationRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type DeletionRequest struct {
	Password string `json:"password" validate:"required,min=8"`
}
type UpdateRequest struct {
	Name         string `json:"name,omitempty" validate:"min=3"`
	Email        string `json:"email,omitempty" validate:"email"`
	LastPassword string `json:"last_password,omitempty" validate:"min=8"`
	NewPassword  string `json:"new_password,omitempty" validate:"min=8"`
}
