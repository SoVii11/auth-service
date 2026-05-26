package usecase

import (
	"time"

	"github.com/SoVii11/auth-service/services/appointment-service/internal/domain"
)

type AppointmentRepo interface {
	Create(a *domain.Appointment) error
	GetAll() ([]domain.Appointment, error)
	GetByUserID(userID int64) ([]domain.Appointment, error)
	UpdateStatus(id int64, status string) error
	GetByID(id int64) (*domain.Appointment, error)
}

type PsychologistRepo interface {
	GetAll() ([]domain.Psychologist, error)
	GetByID(id int64) (*domain.Psychologist, error)
}

type ScheduleRepo interface {
	Create(s *domain.PsychologistSchedule) error
	Delete(id int64) error
	GetByPsychologistID(psychologistID int64) ([]domain.PsychologistSchedule, error)
	IsAvailable(psychologistID int64, date time.Time) (bool, error)
}
