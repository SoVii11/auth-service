package repository

import (
	"database/sql"
	"time"

	"github.com/SoVii11/auth-service/services/auth-service/internal/domain"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *domain.User) error {
	query := `INSERT INTO users (email, password, role, created_at) VALUES ($1, $2, $3, $4) RETURNING id`
	return r.db.QueryRow(query, user.Email, user.Password, user.Role, time.Now()).Scan(&user.ID)
}

func (r *UserRepository) FindByEmail(email string) (*domain.User, error) {
	query := `SELECT id, email, password, role, created_at FROM users WHERE email = $1`
	user := &domain.User{}
	err := r.db.QueryRow(query, email).Scan(&user.ID, &user.Email, &user.Password, &user.Role, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) FindByID(id int64) (*domain.User, error) {
	query := `SELECT id, email, password, role, created_at FROM users WHERE id = $1`
	user := &domain.User{}
	err := r.db.QueryRow(query, id).Scan(&user.ID, &user.Email, &user.Password, &user.Role, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) UpdatePassword(userID int64, hashedPassword string) error {
	_, err := r.db.Exec(`UPDATE users SET password = $1 WHERE id = $2`, hashedPassword, userID)
	return err
}
