package models

import "time"

type User struct {
	ID              int       `json:"id" db:"id"`
	MemberID        string    `json:"member_id" db:"member_id"`
	FirstName       string    `json:"first_name" db:"first_name"`
	LastName        string    `json:"last_name" db:"last_name"`
	Phone           string    `json:"phone" db:"phone"`
	Email           string    `json:"email" db:"email"`
	MembershipDate  time.Time `json:"membership_date" db:"membership_date"`
	MembershipLevel string    `json:"membership_level" db:"membership_level"`
	Points          int       `json:"points" db:"points"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type CreateUserRequest struct {
	FirstName       string `json:"first_name" validate:"required,max=3"`
	LastName        string `json:"last_name" validate:"required,max=3"`
	Phone           string `json:"phone" validate:"required"`
	Email           string `json:"email" validate:"required,email"`
	MembershipLevel string `json:"membership_level" validate:"required,oneof=Gold Silver Bronze"`
	Points          int    `json:"points" validate:"min=0"`
}

type UpdateUserRequest struct {
	FirstName       *string `json:"first_name,omitempty" validate:"omitempty,max=3"`
	LastName        *string `json:"last_name,omitempty" validate:"omitempty,max=3"`
	Phone           *string `json:"phone,omitempty"`
	Email           *string `json:"email,omitempty" validate:"omitempty,email"`
	MembershipLevel *string `json:"membership_level,omitempty" validate:"omitempty,oneof=Gold Silver Bronze"`
	Points          *int    `json:"points,omitempty" validate:"omitempty,min=0"`
}
