package usecase

import "github.com/SoVii11/auth-service/services/auth-service/internal/domain"

type UserRepo interface {
	Create(user *domain.User) error
	FindByEmail(email string) (*domain.User, error)
	FindByID(id int64) (*domain.User, error)
	UpdatePassword(userID int64, hashedPassword string) error
}

type ResetCodeRepo interface {
	Create(code *domain.ResetCode) error
	FindByUserIDAndCode(userID int64, code string) (*domain.ResetCode, error)
	DeleteByUserID(userID int64) error
}
