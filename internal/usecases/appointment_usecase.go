package usecases

import (
	"errors"
	"time"

	"github.com/SoVii11/auth-service/internal/entities"
)

type AppointmentUsecase struct {
	appointmentRepo  AppointmentRepo
	psychologistRepo PsychologistRepo
	scheduleRepo     ScheduleRepo
}

func NewAppointmentUsecase(
	appointmentRepo AppointmentRepo,
	psychologistRepo PsychologistRepo,
	scheduleRepo ScheduleRepo,
) *AppointmentUsecase {
	return &AppointmentUsecase{
		appointmentRepo:  appointmentRepo,
		psychologistRepo: psychologistRepo,
		scheduleRepo:     scheduleRepo,
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
func (u *AppointmentUsecase) AddVacation(psychologistID int64, dateFrom, dateTo time.Time, reason string) (*entities.PsychologistSchedule, error) {
	_, err := u.psychologistRepo.GetByID(psychologistID)
	if err != nil {
		return nil, errors.New("psychologist not found")
	}

	if dateFrom.After(dateTo) {
		return nil, errors.New("date_from must be before date_to")
	}

	s := &entities.PsychologistSchedule{
		PsychologistID: psychologistID,
		DateFrom:       dateFrom,
		DateTo:         dateTo,
		Reason:         reason,
	}

	if err := u.scheduleRepo.Create(s); err != nil {
		return nil, err
	}

	return s, nil
}

func (u *AppointmentUsecase) RemoveVacation(id int64) error {
	return u.scheduleRepo.Delete(id)
}

func (u *AppointmentUsecase) GetVacations(psychologistID int64) ([]entities.PsychologistSchedule, error) {
	return u.scheduleRepo.GetByPsychologistID(psychologistID)
}

func (u *AppointmentUsecase) IsPsychologistAvailable(psychologistID int64, date time.Time) (bool, error) {
	return u.scheduleRepo.IsAvailable(psychologistID, date)
}
