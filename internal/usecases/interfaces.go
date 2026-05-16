package usecases

import "github.com/SoVii11/auth-service/internal/entities"

type UserRepo interface {
	Create(user *entities.User) error
	FindByEmail(email string) (*entities.User, error)
	FindByID(id int64) (*entities.User, error)
	UpdatePassword(userID int64, hashedPassword string) error
}

type ResetCodeRepo interface {
	Create(code *entities.ResetCode) error
	FindByUserIDAndCode(userID int64, code string) (*entities.ResetCode, error)
	DeleteByUserID(userID int64) error
}
