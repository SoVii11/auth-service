package usecase_test

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"github.com/SoVii11/auth-service/services/auth-service/internal/config"
	"github.com/SoVii11/auth-service/services/auth-service/internal/domain"
	usecase "github.com/SoVii11/auth-service/services/auth-service/internal/usecases"
)

func newTestConfig() *config.Config {
	return &config.Config{
		JWTSecret: "testsecret",
		SMTPHost:  "smtp.gmail.com",
		SMTPPort:  "587",
		SMTPUser:  "test@test.com",
		SMTPPass:  "testpass",
	}
}

func hashPassword(password string) string {
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashed)
}

func TestRegister_Success(t *testing.T) {
	userRepo := new(usecase.MockUserRepository)
	resetRepo := new(usecase.MockResetCodeRepository)
	userRepo.On("FindByEmail", "new@test.com").Return(nil, sql.ErrNoRows)
	userRepo.On("Create", mock.AnythingOfType("*domain.User")).Return(nil)
	uc := usecase.NewAuthUsecase(userRepo, resetRepo, newTestConfig())
	user, err := uc.Register("new@test.com", "password123")
	assert.NoError(t, err)
	assert.Equal(t, "new@test.com", user.Email)
	userRepo.AssertExpectations(t)
}

func TestRegister_AlreadyExists(t *testing.T) {
	userRepo := new(usecase.MockUserRepository)
	resetRepo := new(usecase.MockResetCodeRepository)
	userRepo.On("FindByEmail", "exists@test.com").Return(&domain.User{ID: 1}, nil)
	uc := usecase.NewAuthUsecase(userRepo, resetRepo, newTestConfig())
	user, err := uc.Register("exists@test.com", "password123")
	assert.Error(t, err)
	assert.Nil(t, user)
	userRepo.AssertExpectations(t)
}

func TestRegister_DBError(t *testing.T) {
	userRepo := new(usecase.MockUserRepository)
	resetRepo := new(usecase.MockResetCodeRepository)
	userRepo.On("FindByEmail", "new@test.com").Return(nil, sql.ErrNoRows)
	userRepo.On("Create", mock.AnythingOfType("*domain.User")).Return(errors.New("db error"))
	uc := usecase.NewAuthUsecase(userRepo, resetRepo, newTestConfig())
	user, err := uc.Register("new@test.com", "password123")
	assert.Error(t, err)
	assert.Nil(t, user)
	userRepo.AssertExpectations(t)
}

func TestLogin_Success(t *testing.T) {
	userRepo := new(usecase.MockUserRepository)
	resetRepo := new(usecase.MockResetCodeRepository)
	userRepo.On("FindByEmail", "test@test.com").Return(&domain.User{ID: 1, Email: "test@test.com", Password: hashPassword("password123"), Role: "user"}, nil)
	uc := usecase.NewAuthUsecase(userRepo, resetRepo, newTestConfig())
	token, err := uc.Login("test@test.com", "password123")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	userRepo.AssertExpectations(t)
}

func TestLogin_WrongPassword(t *testing.T) {
	userRepo := new(usecase.MockUserRepository)
	resetRepo := new(usecase.MockResetCodeRepository)
	userRepo.On("FindByEmail", "test@test.com").Return(&domain.User{ID: 1, Password: hashPassword("correct")}, nil)
	uc := usecase.NewAuthUsecase(userRepo, resetRepo, newTestConfig())
	token, err := uc.Login("test@test.com", "wrong")
	assert.Error(t, err)
	assert.Empty(t, token)
	userRepo.AssertExpectations(t)
}

func TestLogin_UserNotFound(t *testing.T) {
	userRepo := new(usecase.MockUserRepository)
	resetRepo := new(usecase.MockResetCodeRepository)
	userRepo.On("FindByEmail", "nobody@test.com").Return(nil, sql.ErrNoRows)
	uc := usecase.NewAuthUsecase(userRepo, resetRepo, newTestConfig())
	token, err := uc.Login("nobody@test.com", "password123")
	assert.Error(t, err)
	assert.Empty(t, token)
	userRepo.AssertExpectations(t)
}

func TestLogin_DBError(t *testing.T) {
	userRepo := new(usecase.MockUserRepository)
	resetRepo := new(usecase.MockResetCodeRepository)
	userRepo.On("FindByEmail", "test@test.com").Return(nil, errors.New("db error"))
	uc := usecase.NewAuthUsecase(userRepo, resetRepo, newTestConfig())
	token, err := uc.Login("test@test.com", "password123")
	assert.Error(t, err)
	assert.Empty(t, token)
	userRepo.AssertExpectations(t)
}

func TestResetPassword_Success(t *testing.T) {
	userRepo := new(usecase.MockUserRepository)
	resetRepo := new(usecase.MockResetCodeRepository)
	user := &domain.User{ID: 1, Email: "test@test.com"}
	resetCode := &domain.ResetCode{ID: 1, UserID: 1, Code: "123456", ExpiresAt: time.Now().Add(10 * time.Minute)}
	userRepo.On("FindByEmail", "test@test.com").Return(user, nil)
	resetRepo.On("FindByUserIDAndCode", int64(1), "123456").Return(resetCode, nil)
	userRepo.On("UpdatePassword", int64(1), mock.AnythingOfType("string")).Return(nil)
	resetRepo.On("DeleteByUserID", int64(1)).Return(nil)
	uc := usecase.NewAuthUsecase(userRepo, resetRepo, newTestConfig())
	err := uc.ResetPassword("test@test.com", "123456", "newpassword")
	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
}

func TestResetPassword_ExpiredCode(t *testing.T) {
	userRepo := new(usecase.MockUserRepository)
	resetRepo := new(usecase.MockResetCodeRepository)
	user := &domain.User{ID: 1, Email: "test@test.com"}
	expiredCode := &domain.ResetCode{ID: 1, UserID: 1, Code: "123456", ExpiresAt: time.Now().Add(-10 * time.Minute)}
	userRepo.On("FindByEmail", "test@test.com").Return(user, nil)
	resetRepo.On("FindByUserIDAndCode", int64(1), "123456").Return(expiredCode, nil)
	uc := usecase.NewAuthUsecase(userRepo, resetRepo, newTestConfig())
	err := uc.ResetPassword("test@test.com", "123456", "newpassword")
	assert.Error(t, err)
	assert.Equal(t, "code expired", err.Error())
	userRepo.AssertExpectations(t)
}

func TestResetPassword_InvalidCode(t *testing.T) {
	userRepo := new(usecase.MockUserRepository)
	resetRepo := new(usecase.MockResetCodeRepository)
	userRepo.On("FindByEmail", "test@test.com").Return(&domain.User{ID: 1}, nil)
	resetRepo.On("FindByUserIDAndCode", int64(1), "000000").Return(nil, sql.ErrNoRows)
	uc := usecase.NewAuthUsecase(userRepo, resetRepo, newTestConfig())
	err := uc.ResetPassword("test@test.com", "000000", "newpassword")
	assert.Error(t, err)
	assert.Equal(t, "invalid code", err.Error())
	userRepo.AssertExpectations(t)
}

func TestResetPassword_UserNotFound(t *testing.T) {
	userRepo := new(usecase.MockUserRepository)
	resetRepo := new(usecase.MockResetCodeRepository)
	userRepo.On("FindByEmail", "nobody@test.com").Return(nil, sql.ErrNoRows)
	uc := usecase.NewAuthUsecase(userRepo, resetRepo, newTestConfig())
	err := uc.ResetPassword("nobody@test.com", "123456", "newpassword")
	assert.Error(t, err)
	assert.Equal(t, "user not found", err.Error())
	userRepo.AssertExpectations(t)
}
