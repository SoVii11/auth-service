package usecase

import (
	"github.com/SoVii11/auth-service/services/auth-service/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByEmail(email string) (*domain.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) FindByID(id int64) (*domain.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) UpdatePassword(userID int64, hashedPassword string) error {
	args := m.Called(userID, hashedPassword)
	return args.Error(0)
}

type MockResetCodeRepository struct {
	mock.Mock
}

func (m *MockResetCodeRepository) Create(code *domain.ResetCode) error {
	args := m.Called(code)
	return args.Error(0)
}

func (m *MockResetCodeRepository) FindByUserIDAndCode(userID int64, code string) (*domain.ResetCode, error) {
	args := m.Called(userID, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ResetCode), args.Error(1)
}

func (m *MockResetCodeRepository) DeleteByUserID(userID int64) error {
	args := m.Called(userID)
	return args.Error(0)
}
