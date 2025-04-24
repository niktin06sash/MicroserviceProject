package model

// swagger:model PersonReg
type PersonReg struct {
	Name     string `json:"name" validate:"required,min=3"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// swagger:model PersonAuth
type PersonAuth struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// swagger:model PersonDelete
type PersonDelete struct {
	Password string `json:"password" validate:"required,min=8"`
}
