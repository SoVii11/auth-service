package usecases

import "github.com/SoVii11/auth-service/internal/entities"

type AppointmentRepo interface {
	Create(a *entities.Appointment) error
	GetAll() ([]entities.Appointment, error)
	GetByUserID(userID int64) ([]entities.Appointment, error)
	UpdateStatus(id int64, status string) error
	GetByID(id int64) (*entities.Appointment, error)
}

type PsychologistRepo interface {
	GetAll() ([]entities.Psychologist, error)
	GetByID(id int64) (*entities.Psychologist, error)
}
