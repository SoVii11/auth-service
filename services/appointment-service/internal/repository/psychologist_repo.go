package repository

import (
	"database/sql"

	"github.com/SoVii11/auth-service/services/appointment-service/internal/domain"
)

type PsychologistRepository struct {
	db *sql.DB
}

func NewPsychologistRepository(db *sql.DB) *PsychologistRepository {
	return &PsychologistRepository{db: db}
}

func (r *PsychologistRepository) GetAll() ([]domain.Psychologist, error) {
	rows, err := r.db.Query(`SELECT id, name, description, photo_url, created_at FROM psychologists`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var psychologists []domain.Psychologist
	for rows.Next() {
		var p domain.Psychologist
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.PhotoURL, &p.CreatedAt); err != nil {
			return nil, err
		}
		psychologists = append(psychologists, p)
	}
	return psychologists, nil
}

func (r *PsychologistRepository) GetByID(id int64) (*domain.Psychologist, error) {
	p := &domain.Psychologist{}
	err := r.db.QueryRow(`SELECT id, name, description, photo_url, created_at FROM psychologists WHERE id = $1`, id).
		Scan(&p.ID, &p.Name, &p.Description, &p.PhotoURL, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}
