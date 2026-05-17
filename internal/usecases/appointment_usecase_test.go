package usecases_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/SoVii11/auth-service/internal/entities"
	"github.com/SoVii11/auth-service/internal/usecases"
)

// ───── CreateAppointment ─────

func TestCreateAppointment_Success(t *testing.T) {
	appointmentRepo := new(usecases.MockAppointmentRepository)
	psychologistRepo := new(usecases.MockPsychologistRepository)

	psychologist := &entities.Psychologist{ID: 1, Name: "Анна Иванова"}
	psychologistRepo.On("GetByID", int64(1)).Return(psychologist, nil)
	appointmentRepo.On("Create", mock.AnythingOfType("*entities.Appointment")).Return(nil)

	scheduleRepo := new(usecases.MockScheduleRepository)
	uc := usecases.NewAppointmentUsecase(appointmentRepo, psychologistRepo, scheduleRepo)
	a, err := uc.CreateAppointment(1, 1, "Хочу на консультацию")

	assert.NoError(t, err)
	assert.Equal(t, int64(1), a.UserID)
	assert.Equal(t, "pending", a.Status)
	appointmentRepo.AssertExpectations(t)
	psychologistRepo.AssertExpectations(t)
}

func TestCreateAppointment_PsychologistNotFound(t *testing.T) {
	appointmentRepo := new(usecases.MockAppointmentRepository)
	psychologistRepo := new(usecases.MockPsychologistRepository)

	psychologistRepo.On("GetByID", int64(99)).Return(nil, errors.New("not found"))

	scheduleRepo := new(usecases.MockScheduleRepository)
	uc := usecases.NewAppointmentUsecase(appointmentRepo, psychologistRepo, scheduleRepo)
	a, err := uc.CreateAppointment(1, 99, "Комментарий")

	assert.Error(t, err)
	assert.Nil(t, a)
	assert.Equal(t, "psychologist not found", err.Error())
	psychologistRepo.AssertExpectations(t)
}

func TestCreateAppointment_DBError(t *testing.T) {
	appointmentRepo := new(usecases.MockAppointmentRepository)
	psychologistRepo := new(usecases.MockPsychologistRepository)

	psychologist := &entities.Psychologist{ID: 1, Name: "Анна Иванова"}
	psychologistRepo.On("GetByID", int64(1)).Return(psychologist, nil)
	appointmentRepo.On("Create", mock.AnythingOfType("*entities.Appointment")).Return(errors.New("db error"))

	scheduleRepo := new(usecases.MockScheduleRepository)
	uc := usecases.NewAppointmentUsecase(appointmentRepo, psychologistRepo, scheduleRepo)
	a, err := uc.CreateAppointment(1, 1, "Комментарий")

	assert.Error(t, err)
	assert.Nil(t, a)
	appointmentRepo.AssertExpectations(t)
}

// ───── GetMyAppointments ─────

func TestGetMyAppointments_Success(t *testing.T) {
	appointmentRepo := new(usecases.MockAppointmentRepository)
	psychologistRepo := new(usecases.MockPsychologistRepository)

	appointments := []entities.Appointment{
		{ID: 1, UserID: 1, PsychologistID: 1, Status: "pending"},
		{ID: 2, UserID: 1, PsychologistID: 1, Status: "approved"},
	}
	appointmentRepo.On("GetByUserID", int64(1)).Return(appointments, nil)

	scheduleRepo := new(usecases.MockScheduleRepository)
	uc := usecases.NewAppointmentUsecase(appointmentRepo, psychologistRepo, scheduleRepo)
	result, err := uc.GetMyAppointments(1)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	appointmentRepo.AssertExpectations(t)
}

func TestGetMyAppointments_Empty(t *testing.T) {
	appointmentRepo := new(usecases.MockAppointmentRepository)
	psychologistRepo := new(usecases.MockPsychologistRepository)

	appointmentRepo.On("GetByUserID", int64(1)).Return([]entities.Appointment{}, nil)

	scheduleRepo := new(usecases.MockScheduleRepository)
	uc := usecases.NewAppointmentUsecase(appointmentRepo, psychologistRepo, scheduleRepo)
	result, err := uc.GetMyAppointments(1)

	assert.NoError(t, err)
	assert.Len(t, result, 0)
	appointmentRepo.AssertExpectations(t)
}

// ───── GetAllAppointments ─────

func TestGetAllAppointments_Success(t *testing.T) {
	appointmentRepo := new(usecases.MockAppointmentRepository)
	psychologistRepo := new(usecases.MockPsychologistRepository)

	appointments := []entities.Appointment{
		{ID: 1, UserID: 1, Status: "pending"},
		{ID: 2, UserID: 2, Status: "approved"},
		{ID: 3, UserID: 3, Status: "rejected"},
	}
	appointmentRepo.On("GetAll").Return(appointments, nil)

	scheduleRepo := new(usecases.MockScheduleRepository)
	uc := usecases.NewAppointmentUsecase(appointmentRepo, psychologistRepo, scheduleRepo)
	result, err := uc.GetAllAppointments()

	assert.NoError(t, err)
	assert.Len(t, result, 3)
	appointmentRepo.AssertExpectations(t)
}

// ───── ApproveAppointment ─────

func TestApproveAppointment_Success(t *testing.T) {
	appointmentRepo := new(usecases.MockAppointmentRepository)
	psychologistRepo := new(usecases.MockPsychologistRepository)

	appointment := &entities.Appointment{ID: 1, UserID: 1, Status: "pending"}
	appointmentRepo.On("GetByID", int64(1)).Return(appointment, nil)
	appointmentRepo.On("UpdateStatus", int64(1), "approved").Return(nil)

	scheduleRepo := new(usecases.MockScheduleRepository)
	uc := usecases.NewAppointmentUsecase(appointmentRepo, psychologistRepo, scheduleRepo)
	err := uc.ApproveAppointment(1)

	assert.NoError(t, err)
	appointmentRepo.AssertExpectations(t)
}

func TestApproveAppointment_NotFound(t *testing.T) {
	appointmentRepo := new(usecases.MockAppointmentRepository)
	psychologistRepo := new(usecases.MockPsychologistRepository)

	appointmentRepo.On("GetByID", int64(99)).Return(nil, errors.New("not found"))

	scheduleRepo := new(usecases.MockScheduleRepository)
	uc := usecases.NewAppointmentUsecase(appointmentRepo, psychologistRepo, scheduleRepo)
	err := uc.ApproveAppointment(99)

	assert.Error(t, err)
	assert.Equal(t, "appointment not found", err.Error())
	appointmentRepo.AssertExpectations(t)
}

func TestApproveAppointment_AlreadyApproved(t *testing.T) {
	appointmentRepo := new(usecases.MockAppointmentRepository)
	psychologistRepo := new(usecases.MockPsychologistRepository)

	appointment := &entities.Appointment{ID: 1, Status: "approved"}
	appointmentRepo.On("GetByID", int64(1)).Return(appointment, nil)

	scheduleRepo := new(usecases.MockScheduleRepository)
	uc := usecases.NewAppointmentUsecase(appointmentRepo, psychologistRepo, scheduleRepo)
	err := uc.ApproveAppointment(1)

	assert.Error(t, err)
	assert.Equal(t, "appointment is not pending", err.Error())
	appointmentRepo.AssertExpectations(t)
}

// ───── RejectAppointment ─────

func TestRejectAppointment_Success(t *testing.T) {
	appointmentRepo := new(usecases.MockAppointmentRepository)
	psychologistRepo := new(usecases.MockPsychologistRepository)

	appointment := &entities.Appointment{ID: 1, Status: "pending"}
	appointmentRepo.On("GetByID", int64(1)).Return(appointment, nil)
	appointmentRepo.On("UpdateStatus", int64(1), "rejected").Return(nil)

	scheduleRepo := new(usecases.MockScheduleRepository)
	uc := usecases.NewAppointmentUsecase(appointmentRepo, psychologistRepo, scheduleRepo)
	err := uc.RejectAppointment(1)

	assert.NoError(t, err)
	appointmentRepo.AssertExpectations(t)
}

func TestRejectAppointment_NotFound(t *testing.T) {
	appointmentRepo := new(usecases.MockAppointmentRepository)
	psychologistRepo := new(usecases.MockPsychologistRepository)

	appointmentRepo.On("GetByID", int64(99)).Return(nil, errors.New("not found"))

	scheduleRepo := new(usecases.MockScheduleRepository)
	uc := usecases.NewAppointmentUsecase(appointmentRepo, psychologistRepo, scheduleRepo)
	err := uc.RejectAppointment(99)

	assert.Error(t, err)
	assert.Equal(t, "appointment not found", err.Error())
	appointmentRepo.AssertExpectations(t)
}

func TestRejectAppointment_AlreadyRejected(t *testing.T) {
	appointmentRepo := new(usecases.MockAppointmentRepository)
	psychologistRepo := new(usecases.MockPsychologistRepository)

	appointment := &entities.Appointment{ID: 1, Status: "rejected"}
	appointmentRepo.On("GetByID", int64(1)).Return(appointment, nil)

	scheduleRepo := new(usecases.MockScheduleRepository)
	uc := usecases.NewAppointmentUsecase(appointmentRepo, psychologistRepo, scheduleRepo)
	err := uc.RejectAppointment(1)

	assert.Error(t, err)
	assert.Equal(t, "appointment is not pending", err.Error())
	appointmentRepo.AssertExpectations(t)
}

// ───── GetPsychologists ─────

func TestGetPsychologists_Success(t *testing.T) {
	appointmentRepo := new(usecases.MockAppointmentRepository)
	psychologistRepo := new(usecases.MockPsychologistRepository)

	psychologists := []entities.Psychologist{
		{ID: 1, Name: "Анна Иванова", CreatedAt: time.Now()},
		{ID: 2, Name: "Иван Петров", CreatedAt: time.Now()},
	}
	psychologistRepo.On("GetAll").Return(psychologists, nil)

	scheduleRepo := new(usecases.MockScheduleRepository)
	uc := usecases.NewAppointmentUsecase(appointmentRepo, psychologistRepo, scheduleRepo)
	result, err := uc.GetPsychologists()

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	psychologistRepo.AssertExpectations(t)
}

// ───── GetPsychologistByID ─────

func TestGetPsychologistByID_Success(t *testing.T) {
	appointmentRepo := new(usecases.MockAppointmentRepository)
	psychologistRepo := new(usecases.MockPsychologistRepository)

	psychologist := &entities.Psychologist{ID: 1, Name: "Анна Иванова"}
	psychologistRepo.On("GetByID", int64(1)).Return(psychologist, nil)

	scheduleRepo := new(usecases.MockScheduleRepository)
	uc := usecases.NewAppointmentUsecase(appointmentRepo, psychologistRepo, scheduleRepo)
	result, err := uc.GetPsychologistByID(1)

	assert.NoError(t, err)
	assert.Equal(t, "Анна Иванова", result.Name)
	psychologistRepo.AssertExpectations(t)
}

func TestGetPsychologistByID_NotFound(t *testing.T) {
	appointmentRepo := new(usecases.MockAppointmentRepository)
	psychologistRepo := new(usecases.MockPsychologistRepository)

	psychologistRepo.On("GetByID", int64(99)).Return(nil, errors.New("not found"))

	scheduleRepo := new(usecases.MockScheduleRepository)
	uc := usecases.NewAppointmentUsecase(appointmentRepo, psychologistRepo, scheduleRepo)
	result, err := uc.GetPsychologistByID(99)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "psychologist not found", err.Error())
	psychologistRepo.AssertExpectations(t)
}
