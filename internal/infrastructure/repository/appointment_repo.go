package repository

import (
	"database/sql"

	"github.com/SoVii11/auth-service/internal/entities"
)

type AppointmentRepository struct {
	db *sql.DB
}

func NewAppointmentRepository(db *sql.DB) *AppointmentRepository {
	return &AppointmentRepository{db: db}
}

func (r *AppointmentRepository) Create(a *entities.Appointment) error {
	query := `
		INSERT INTO appointments (user_id, psychologist_id, status, comment, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id`
	return r.db.QueryRow(query, a.UserID, a.PsychologistID, "pending", a.Comment).Scan(&a.ID)
}

func (r *AppointmentRepository) GetAll() ([]entities.Appointment, error) {
	query := `SELECT id, user_id, psychologist_id, status, comment, created_at FROM appointments`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var appointments []entities.Appointment
	for rows.Next() {
		var a entities.Appointment
		if err := rows.Scan(&a.ID, &a.UserID, &a.PsychologistID, &a.Status, &a.Comment, &a.CreatedAt); err != nil {
			return nil, err
		}
		appointments = append(appointments, a)
	}
	return appointments, nil
}

func (r *AppointmentRepository) GetByUserID(userID int64) ([]entities.Appointment, error) {
	query := `SELECT id, user_id, psychologist_id, status, comment, created_at FROM appointments WHERE user_id = $1`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var appointments []entities.Appointment
	for rows.Next() {
		var a entities.Appointment
		if err := rows.Scan(&a.ID, &a.UserID, &a.PsychologistID, &a.Status, &a.Comment, &a.CreatedAt); err != nil {
			return nil, err
		}
		appointments = append(appointments, a)
	}
	return appointments, nil
}

func (r *AppointmentRepository) UpdateStatus(id int64, status string) error {
	query := `UPDATE appointments SET status = $1 WHERE id = $2`
	_, err := r.db.Exec(query, status, id)
	return err
}

func (r *AppointmentRepository) GetByID(id int64) (*entities.Appointment, error) {
	query := `SELECT id, user_id, psychologist_id, status, comment, created_at FROM appointments WHERE id = $1`
	a := &entities.Appointment{}
	err := r.db.QueryRow(query, id).Scan(&a.ID, &a.UserID, &a.PsychologistID, &a.Status, &a.Comment, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	return a, nil
}
