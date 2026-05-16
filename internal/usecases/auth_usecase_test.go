package usecases_test

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"github.com/SoVii11/auth-service/config"
	"github.com/SoVii11/auth-service/internal/entities"
	"github.com/SoVii11/auth-service/internal/usecases"
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

// ───── Register ─────

func TestRegister_Success(t *testing.T) {
	userRepo := new(usecases.MockUserRepository)
	resetRepo := new(usecases.MockResetCodeRepository)

	userRepo.On("FindByEmail", "new@test.com").Return(nil, sql.ErrNoRows)
	userRepo.On("Create", mock.AnythingOfType("*entities.User")).Return(nil)

	uc := usecases.NewAuthUsecase(userRepo, resetRepo, newTestConfig())
	user, err := uc.Register("new@test.com", "password123")

	assert.NoError(t, err)
	assert.Equal(t, "new@test.com", user.Email)
	assert.Equal(t, "user", user.Role)
	userRepo.AssertExpectations(t)
}

func TestRegister_AlreadyExists(t *testing.T) {
	userRepo := new(usecases.MockUserRepository)
	resetRepo := new(usecases.MockResetCodeRepository)

	existingUser := &entities.User{ID: 1, Email: "exists@test.com"}
	userRepo.On("FindByEmail", "exists@test.com").Return(existingUser, nil)

	uc := usecases.NewAuthUsecase(userRepo, resetRepo, newTestConfig())
	user, err := uc.Register("exists@test.com", "password123")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "user with this email already exists", err.Error())
	userRepo.AssertExpectations(t)
}

// ───── Login ─────

func TestLogin_Success(t *testing.T) {
	userRepo := new(usecases.MockUserRepository)
	resetRepo := new(usecases.MockResetCodeRepository)

	existingUser := &entities.User{
		ID:       1,
		Email:    "test@test.com",
		Password: hashPassword("password123"),
		Role:     "user",
	}
	userRepo.On("FindByEmail", "test@test.com").Return(existingUser, nil)

	uc := usecases.NewAuthUsecase(userRepo, resetRepo, newTestConfig())
	token, err := uc.Login("test@test.com", "password123")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	userRepo.AssertExpectations(t)
}

func TestLogin_WrongPassword(t *testing.T) {
	userRepo := new(usecases.MockUserRepository)
	resetRepo := new(usecases.MockResetCodeRepository)

	existingUser := &entities.User{
		ID:       1,
		Email:    "test@test.com",
		Password: hashPassword("correctpassword"),
		Role:     "user",
	}
	userRepo.On("FindByEmail", "test@test.com").Return(existingUser, nil)

	uc := usecases.NewAuthUsecase(userRepo, resetRepo, newTestConfig())
	token, err := uc.Login("test@test.com", "wrongpassword")

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Equal(t, "invalid email or password", err.Error())
	userRepo.AssertExpectations(t)
}

func TestLogin_UserNotFound(t *testing.T) {
	userRepo := new(usecases.MockUserRepository)
	resetRepo := new(usecases.MockResetCodeRepository)

	userRepo.On("FindByEmail", "nobody@test.com").Return(nil, sql.ErrNoRows)

	uc := usecases.NewAuthUsecase(userRepo, resetRepo, newTestConfig())
	token, err := uc.Login("nobody@test.com", "password123")

	assert.Error(t, err)
	assert.Empty(t, token)
	userRepo.AssertExpectations(t)
}

// ───── ResetPassword ─────

func TestResetPassword_Success(t *testing.T) {
	userRepo := new(usecases.MockUserRepository)
	resetRepo := new(usecases.MockResetCodeRepository)

	user := &entities.User{ID: 1, Email: "test@test.com"}
	resetCode := &entities.ResetCode{
		ID:        1,
		UserID:    1,
		Code:      "123456",
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	userRepo.On("FindByEmail", "test@test.com").Return(user, nil)
	resetRepo.On("FindByUserIDAndCode", int64(1), "123456").Return(resetCode, nil)
	userRepo.On("UpdatePassword", int64(1), mock.AnythingOfType("string")).Return(nil)
	resetRepo.On("DeleteByUserID", int64(1)).Return(nil)

	uc := usecases.NewAuthUsecase(userRepo, resetRepo, newTestConfig())
	err := uc.ResetPassword("test@test.com", "123456", "newpassword")

	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
	resetRepo.AssertExpectations(t)
}

func TestResetPassword_ExpiredCode(t *testing.T) {
	userRepo := new(usecases.MockUserRepository)
	resetRepo := new(usecases.MockResetCodeRepository)

	user := &entities.User{ID: 1, Email: "test@test.com"}
	expiredCode := &entities.ResetCode{
		ID:        1,
		UserID:    1,
		Code:      "123456",
		ExpiresAt: time.Now().Add(-10 * time.Minute),
	}

	userRepo.On("FindByEmail", "test@test.com").Return(user, nil)
	resetRepo.On("FindByUserIDAndCode", int64(1), "123456").Return(expiredCode, nil)

	uc := usecases.NewAuthUsecase(userRepo, resetRepo, newTestConfig())
	err := uc.ResetPassword("test@test.com", "123456", "newpassword")

	assert.Error(t, err)
	assert.Equal(t, "code expired", err.Error())
	userRepo.AssertExpectations(t)
	resetRepo.AssertExpectations(t)
}

func TestResetPassword_InvalidCode(t *testing.T) {
	userRepo := new(usecases.MockUserRepository)
	resetRepo := new(usecases.MockResetCodeRepository)

	user := &entities.User{ID: 1, Email: "test@test.com"}

	userRepo.On("FindByEmail", "test@test.com").Return(user, nil)
	resetRepo.On("FindByUserIDAndCode", int64(1), "000000").Return(nil, sql.ErrNoRows)

	uc := usecases.NewAuthUsecase(userRepo, resetRepo, newTestConfig())
	err := uc.ResetPassword("test@test.com", "000000", "newpassword")

	assert.Error(t, err)
	assert.Equal(t, "invalid code", err.Error())
	userRepo.AssertExpectations(t)
	resetRepo.AssertExpectations(t)
}

func TestRegister_DBError(t *testing.T) {
	userRepo := new(usecases.MockUserRepository)
	resetRepo := new(usecases.MockResetCodeRepository)

	userRepo.On("FindByEmail", "new@test.com").Return(nil, sql.ErrNoRows)
	userRepo.On("Create", mock.AnythingOfType("*entities.User")).Return(errors.New("db error"))

	uc := usecases.NewAuthUsecase(userRepo, resetRepo, newTestConfig())
	user, err := uc.Register("new@test.com", "password123")

	assert.Error(t, err)
	assert.Nil(t, user)
	userRepo.AssertExpectations(t)
}

func TestResetPassword_UserNotFound(t *testing.T) {
	userRepo := new(usecases.MockUserRepository)
	resetRepo := new(usecases.MockResetCodeRepository)

	userRepo.On("FindByEmail", "nobody@test.com").Return(nil, sql.ErrNoRows)

	uc := usecases.NewAuthUsecase(userRepo, resetRepo, newTestConfig())
	err := uc.ResetPassword("nobody@test.com", "123456", "newpassword")

	assert.Error(t, err)
	assert.Equal(t, "user not found", err.Error())
	userRepo.AssertExpectations(t)
}

func TestLogin_DBError(t *testing.T) {
	userRepo := new(usecases.MockUserRepository)
	resetRepo := new(usecases.MockResetCodeRepository)

	userRepo.On("FindByEmail", "test@test.com").Return(nil, errors.New("db error"))

	uc := usecases.NewAuthUsecase(userRepo, resetRepo, newTestConfig())
	token, err := uc.Login("test@test.com", "password123")

	assert.Error(t, err)
	assert.Empty(t, token)
	userRepo.AssertExpectations(t)
}

func TestResetPassword_UpdatePasswordError(t *testing.T) {
	userRepo := new(usecases.MockUserRepository)
	resetRepo := new(usecases.MockResetCodeRepository)

	user := &entities.User{ID: 1, Email: "test@test.com"}
	resetCode := &entities.ResetCode{
		ID:        1,
		UserID:    1,
		Code:      "123456",
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	userRepo.On("FindByEmail", "test@test.com").Return(user, nil)
	resetRepo.On("FindByUserIDAndCode", int64(1), "123456").Return(resetCode, nil)
	userRepo.On("UpdatePassword", int64(1), mock.AnythingOfType("string")).Return(errors.New("db error"))

	uc := usecases.NewAuthUsecase(userRepo, resetRepo, newTestConfig())
	err := uc.ResetPassword("test@test.com", "123456", "newpassword")

	assert.Error(t, err)
	userRepo.AssertExpectations(t)
	resetRepo.AssertExpectations(t)
}

func TestSendResetCode_UserNotFound(t *testing.T) {
	userRepo := new(usecases.MockUserRepository)
	resetRepo := new(usecases.MockResetCodeRepository)

	userRepo.On("FindByEmail", "nobody@test.com").Return(nil, sql.ErrNoRows)

	uc := usecases.NewAuthUsecase(userRepo, resetRepo, newTestConfig())
	err := uc.SendResetCode("nobody@test.com")

	assert.Error(t, err)
	assert.Equal(t, "user not found", err.Error())
	userRepo.AssertExpectations(t)
}

func TestSendResetCode_DBError(t *testing.T) {
	userRepo := new(usecases.MockUserRepository)
	resetRepo := new(usecases.MockResetCodeRepository)

	userRepo.On("FindByEmail", "test@test.com").Return(nil, errors.New("db error"))

	uc := usecases.NewAuthUsecase(userRepo, resetRepo, newTestConfig())
	err := uc.SendResetCode("test@test.com")

	assert.Error(t, err)
	userRepo.AssertExpectations(t)
}

func TestSendResetCode_CreateCodeError(t *testing.T) {
	userRepo := new(usecases.MockUserRepository)
	resetRepo := new(usecases.MockResetCodeRepository)

	user := &entities.User{ID: 1, Email: "test@test.com"}

	userRepo.On("FindByEmail", "test@test.com").Return(user, nil)
	resetRepo.On("DeleteByUserID", int64(1)).Return(nil)
	resetRepo.On("Create", mock.AnythingOfType("*entities.ResetCode")).Return(errors.New("db error"))

	uc := usecases.NewAuthUsecase(userRepo, resetRepo, newTestConfig())
	err := uc.SendResetCode("test@test.com")

	assert.Error(t, err)
	userRepo.AssertExpectations(t)
	resetRepo.AssertExpectations(t)
}

func TestResetPassword_DeleteCodeError(t *testing.T) {
	userRepo := new(usecases.MockUserRepository)
	resetRepo := new(usecases.MockResetCodeRepository)

	user := &entities.User{ID: 1, Email: "test@test.com"}
	resetCode := &entities.ResetCode{
		ID:        1,
		UserID:    1,
		Code:      "123456",
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	userRepo.On("FindByEmail", "test@test.com").Return(user, nil)
	resetRepo.On("FindByUserIDAndCode", int64(1), "123456").Return(resetCode, nil)
	userRepo.On("UpdatePassword", int64(1), mock.AnythingOfType("string")).Return(nil)
	resetRepo.On("DeleteByUserID", int64(1)).Return(errors.New("db error"))

	uc := usecases.NewAuthUsecase(userRepo, resetRepo, newTestConfig())
	err := uc.ResetPassword("test@test.com", "123456", "newpassword")

	assert.Error(t, err)
	userRepo.AssertExpectations(t)
	resetRepo.AssertExpectations(t)
}

func TestRegister_EmptyEmail(t *testing.T) {
	userRepo := new(usecases.MockUserRepository)
	resetRepo := new(usecases.MockResetCodeRepository)

	userRepo.On("FindByEmail", "").Return(nil, sql.ErrNoRows)
	userRepo.On("Create", mock.AnythingOfType("*entities.User")).Return(nil)

	uc := usecases.NewAuthUsecase(userRepo, resetRepo, newTestConfig())
	user, err := uc.Register("", "password123")

	assert.NoError(t, err)
	assert.Equal(t, "", user.Email)
}

func TestLogin_EmptyFields(t *testing.T) {
	userRepo := new(usecases.MockUserRepository)
	resetRepo := new(usecases.MockResetCodeRepository)

	userRepo.On("FindByEmail", "").Return(nil, sql.ErrNoRows)

	uc := usecases.NewAuthUsecase(userRepo, resetRepo, newTestConfig())
	token, err := uc.Login("", "")

	assert.Error(t, err)
	assert.Empty(t, token)
	userRepo.AssertExpectations(t)
}
