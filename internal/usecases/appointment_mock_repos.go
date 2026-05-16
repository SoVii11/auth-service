package usecases

import (
	"github.com/SoVii11/auth-service/internal/entities"
	"github.com/stretchr/testify/mock"
)

type MockAppointmentRepository struct {
	mock.Mock
}

func (m *MockAppointmentRepository) Create(a *entities.Appointment) error {
	args := m.Called(a)
	return args.Error(0)
}

func (m *MockAppointmentRepository) GetAll() ([]entities.Appointment, error) {
	args := m.Called()
	return args.Get(0).([]entities.Appointment), args.Error(1)
}

func (m *MockAppointmentRepository) GetByUserID(userID int64) ([]entities.Appointment, error) {
	args := m.Called(userID)
	return args.Get(0).([]entities.Appointment), args.Error(1)
}

func (m *MockAppointmentRepository) UpdateStatus(id int64, status string) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func (m *MockAppointmentRepository) GetByID(id int64) (*entities.Appointment, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Appointment), args.Error(1)
}

type MockPsychologistRepository struct {
	mock.Mock
}

func (m *MockPsychologistRepository) GetAll() ([]entities.Psychologist, error) {
	args := m.Called()
	return args.Get(0).([]entities.Psychologist), args.Error(1)
}

func (m *MockPsychologistRepository) GetByID(id int64) (*entities.Psychologist, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Psychologist), args.Error(1)
}
