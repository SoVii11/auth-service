package usecases

import (
	"errors"

	"github.com/SoVii11/auth-service/internal/entities"
)

type AppointmentUsecase struct {
	appointmentRepo  AppointmentRepo
	psychologistRepo PsychologistRepo
}

func NewAppointmentUsecase(
	appointmentRepo AppointmentRepo,
	psychologistRepo PsychologistRepo,
) *AppointmentUsecase {
	return &AppointmentUsecase{
		appointmentRepo:  appointmentRepo,
		psychologistRepo: psychologistRepo,
	}
}
func (u *AppointmentUsecase) CreateAppointment(userID, psychologistID int64, comment string) (*entities.Appointment, error) {
	_, err := u.psychologistRepo.GetByID(psychologistID)
	if err != nil {
		return nil, errors.New("psychologist not found")
	}

	a := &entities.Appointment{
		UserID:         userID,
		PsychologistID: psychologistID,
		Comment:        comment,
		Status:         "pending",
	}

	if err := u.appointmentRepo.Create(a); err != nil {
		return nil, err
	}

	return a, nil
}

func (u *AppointmentUsecase) GetMyAppointments(userID int64) ([]entities.Appointment, error) {
	return u.appointmentRepo.GetByUserID(userID)
}

func (u *AppointmentUsecase) GetAllAppointments() ([]entities.Appointment, error) {
	return u.appointmentRepo.GetAll()
}

func (u *AppointmentUsecase) ApproveAppointment(id int64) error {
	a, err := u.appointmentRepo.GetByID(id)
	if err != nil {
		return errors.New("appointment not found")
	}
	if a.Status != "pending" {
		return errors.New("appointment is not pending")
	}
	return u.appointmentRepo.UpdateStatus(id, "approved")
}

func (u *AppointmentUsecase) RejectAppointment(id int64) error {
	a, err := u.appointmentRepo.GetByID(id)
	if err != nil {
		return errors.New("appointment not found")
	}
	if a.Status != "pending" {
		return errors.New("appointment is not pending")
	}
	return u.appointmentRepo.UpdateStatus(id, "rejected")
}

func (u *AppointmentUsecase) GetPsychologists() ([]entities.Psychologist, error) {
	return u.psychologistRepo.GetAll()
}

func (u *AppointmentUsecase) GetPsychologistByID(id int64) (*entities.Psychologist, error) {
	p, err := u.psychologistRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("psychologist not found")
	}
	return p, nil
}
