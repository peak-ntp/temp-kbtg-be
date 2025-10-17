package services

import (
	"errors"
	"kbtg-backend/internal/models"
	"kbtg-backend/internal/repositories"
)

type UserService struct {
	repo *repositories.UserRepository
}

func NewUserService(repo *repositories.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetAllUsers() ([]models.User, error) {
	return s.repo.GetAll()
}

func (s *UserService) GetUserByID(id int) (*models.User, error) {
	if id <= 0 {
		return nil, errors.New("invalid user ID")
	}
	return s.repo.GetByID(id)
}

func (s *UserService) CreateUser(req models.CreateUserRequest) (*models.User, error) {
	// Validate business rules
	if len(req.FirstName) > 3 {
		return nil, errors.New("first name cannot exceed 3 characters")
	}
	if len(req.LastName) > 3 {
		return nil, errors.New("last name cannot exceed 3 characters")
	}
	if req.FirstName == "" || req.LastName == "" {
		return nil, errors.New("first name and last name are required")
	}
	if req.Email == "" {
		return nil, errors.New("email is required")
	}
	if req.Phone == "" {
		return nil, errors.New("phone is required")
	}
	if req.MembershipLevel == "" {
		req.MembershipLevel = "Bronze" // Default membership level
	}
	if req.Points < 0 {
		return nil, errors.New("points cannot be negative")
	}

	return s.repo.Create(req)
}

func (s *UserService) UpdateUser(id int, req models.UpdateUserRequest) (*models.User, error) {
	if id <= 0 {
		return nil, errors.New("invalid user ID")
	}

	// Validate business rules
	if req.FirstName != nil && len(*req.FirstName) > 3 {
		return nil, errors.New("first name cannot exceed 3 characters")
	}
	if req.LastName != nil && len(*req.LastName) > 3 {
		return nil, errors.New("last name cannot exceed 3 characters")
	}
	if req.Points != nil && *req.Points < 0 {
		return nil, errors.New("points cannot be negative")
	}

	return s.repo.Update(id, req)
}

func (s *UserService) DeleteUser(id int) error {
	if id <= 0 {
		return errors.New("invalid user ID")
	}
	return s.repo.Delete(id)
}
