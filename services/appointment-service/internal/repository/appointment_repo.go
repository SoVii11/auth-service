package repository

import (
	"database/sql"

	"github.com/SoVii11/auth-service/services/appointment-service/internal/domain"
)

type AppointmentRepository struct {
	db *sql.DB
}

func NewAppointmentRepository(db *sql.DB) *AppointmentRepository {
	return &AppointmentRepository{db: db}
}

func (r *AppointmentRepository) Create(a *domain.Appointment) error {
	query := `INSERT INTO appointments (user_id, psychologist_id, status, comment, created_at) VALUES ($1, $2, $3, $4, NOW()) RETURNING id`
	return r.db.QueryRow(query, a.UserID, a.PsychologistID, "pending", a.Comment).Scan(&a.ID)
}

func (r *AppointmentRepository) GetAll() ([]domain.Appointment, error) {
	rows, err := r.db.Query(`SELECT id, user_id, psychologist_id, status, comment, created_at FROM appointments`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var appointments []domain.Appointment
	for rows.Next() {
		var a domain.Appointment
		if err := rows.Scan(&a.ID, &a.UserID, &a.PsychologistID, &a.Status, &a.Comment, &a.CreatedAt); err != nil {
			return nil, err
		}
		appointments = append(appointments, a)
	}
	return appointments, nil
}

func (r *AppointmentRepository) GetByUserID(userID int64) ([]domain.Appointment, error) {
	rows, err := r.db.Query(`SELECT id, user_id, psychologist_id, status, comment, created_at FROM appointments WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var appointments []domain.Appointment
	for rows.Next() {
		var a domain.Appointment
		if err := rows.Scan(&a.ID, &a.UserID, &a.PsychologistID, &a.Status, &a.Comment, &a.CreatedAt); err != nil {
			return nil, err
		}
		appointments = append(appointments, a)
	}
	return appointments, nil
}

func (r *AppointmentRepository) UpdateStatus(id int64, status string) error {
	_, err := r.db.Exec(`UPDATE appointments SET status = $1 WHERE id = $2`, status, id)
	return err
}

func (r *AppointmentRepository) GetByID(id int64) (*domain.Appointment, error) {
	a := &domain.Appointment{}
	err := r.db.QueryRow(`SELECT id, user_id, psychologist_id, status, comment, created_at FROM appointments WHERE id = $1`, id).
		Scan(&a.ID, &a.UserID, &a.PsychologistID, &a.Status, &a.Comment, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	return a, nil
}
