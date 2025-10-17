package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"kbtg-backend/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetAll() ([]models.User, error) {
	query := `
		SELECT id, member_id, first_name, last_name, phone, email, 
		       membership_date, membership_level, points, created_at, updated_at 
		FROM users ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.MemberID, &user.FirstName, &user.LastName,
			&user.Phone, &user.Email, &user.MembershipDate, &user.MembershipLevel,
			&user.Points, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (r *UserRepository) GetByID(id int) (*models.User, error) {
	query := `
		SELECT id, member_id, first_name, last_name, phone, email,
		       membership_date, membership_level, points, created_at, updated_at 
		FROM users WHERE id = ?`

	var user models.User
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.MemberID, &user.FirstName, &user.LastName,
		&user.Phone, &user.Email, &user.MembershipDate, &user.MembershipLevel,
		&user.Points, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) Create(req models.CreateUserRequest) (*models.User, error) {
	memberID := r.generateMemberID()
	now := time.Now()

	query := `
		INSERT INTO users (member_id, first_name, last_name, phone, email, 
		                  membership_date, membership_level, points, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id, member_id, first_name, last_name, phone, email,
		          membership_date, membership_level, points, created_at, updated_at`

	var user models.User
	err := r.db.QueryRow(
		query, memberID, req.FirstName, req.LastName, req.Phone, req.Email,
		now, req.MembershipLevel, req.Points, now, now,
	).Scan(
		&user.ID, &user.MemberID, &user.FirstName, &user.LastName,
		&user.Phone, &user.Email, &user.MembershipDate, &user.MembershipLevel,
		&user.Points, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) Update(id int, req models.UpdateUserRequest) (*models.User, error) {
	// First, get the current user
	user, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}

	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}

	if req.FirstName != nil {
		setParts = append(setParts, "first_name = ?")
		args = append(args, *req.FirstName)
	}
	if req.LastName != nil {
		setParts = append(setParts, "last_name = ?")
		args = append(args, *req.LastName)
	}
	if req.Phone != nil {
		setParts = append(setParts, "phone = ?")
		args = append(args, *req.Phone)
	}
	if req.Email != nil {
		setParts = append(setParts, "email = ?")
		args = append(args, *req.Email)
	}
	if req.MembershipLevel != nil {
		setParts = append(setParts, "membership_level = ?")
		args = append(args, *req.MembershipLevel)
	}
	if req.Points != nil {
		setParts = append(setParts, "points = ?")
		args = append(args, *req.Points)
	}

	if len(setParts) == 0 {
		return user, nil // No updates
	}

	// Add updated_at
	setParts = append(setParts, "updated_at = ?")
	args = append(args, time.Now())
	args = append(args, id)

	query := fmt.Sprintf("UPDATE users SET %s WHERE id = ?",
		fmt.Sprintf("%s", setParts[0]))
	for i := 1; i < len(setParts); i++ {
		query = fmt.Sprintf("%s, %s", query, setParts[i])
	}

	_, err = r.db.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	return r.GetByID(id)
}

func (r *UserRepository) Delete(id int) error {
	query := "DELETE FROM users WHERE id = ?"
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *UserRepository) generateMemberID() string {
	// Simple member ID generation - in production, you might want something more sophisticated
	var count int
	r.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return fmt.Sprintf("LBK%06d", count+1)
}
