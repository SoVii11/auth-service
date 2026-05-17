package repository

import (
	"database/sql"
	"time"

	"github.com/SoVii11/auth-service/internal/entities"
)

type ScheduleRepository struct {
	db *sql.DB
}

func NewScheduleRepository(db *sql.DB) *ScheduleRepository {
	return &ScheduleRepository{db: db}
}

func (r *ScheduleRepository) Create(s *entities.PsychologistSchedule) error {
	query := `
		INSERT INTO psychologist_schedules (psychologist_id, date_from, date_to, reason, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`
	return r.db.QueryRow(query, s.PsychologistID, s.DateFrom, s.DateTo, s.Reason, time.Now()).Scan(&s.ID)
}

func (r *ScheduleRepository) Delete(id int64) error {
	query := `DELETE FROM psychologist_schedules WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *ScheduleRepository) GetByPsychologistID(psychologistID int64) ([]entities.PsychologistSchedule, error) {
	query := `
		SELECT id, psychologist_id, date_from, date_to, reason, created_at
		FROM psychologist_schedules
		WHERE psychologist_id = $1
		ORDER BY date_from ASC`
	rows, err := r.db.Query(query, psychologistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []entities.PsychologistSchedule
	for rows.Next() {
		var s entities.PsychologistSchedule
		if err := rows.Scan(&s.ID, &s.PsychologistID, &s.DateFrom, &s.DateTo, &s.Reason, &s.CreatedAt); err != nil {
			return nil, err
		}
		schedules = append(schedules, s)
	}
	return schedules, nil
}

func (r *ScheduleRepository) IsAvailable(psychologistID int64, date time.Time) (bool, error) {
	query := `
		SELECT COUNT(*) FROM psychologist_schedules
		WHERE psychologist_id = $1
		AND date_from <= $2
		AND date_to >= $2`
	var count int
	err := r.db.QueryRow(query, psychologistID, date).Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}
