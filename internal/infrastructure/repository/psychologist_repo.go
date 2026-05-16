package repository

import (
	"database/sql"

	"github.com/SoVii11/auth-service/internal/entities"
)

type PsychologistRepository struct {
	db *sql.DB
}

func NewPsychologistRepository(db *sql.DB) *PsychologistRepository {
	return &PsychologistRepository{db: db}
}

func (r *PsychologistRepository) GetAll() ([]entities.Psychologist, error) {
	query := `SELECT id, name, description, photo_url, created_at FROM psychologists`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var psychologists []entities.Psychologist
	for rows.Next() {
		var p entities.Psychologist
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.PhotoURL, &p.CreatedAt); err != nil {
			return nil, err
		}
		psychologists = append(psychologists, p)
	}
	return psychologists, nil
}

func (r *PsychologistRepository) GetByID(id int64) (*entities.Psychologist, error) {
	query := `SELECT id, name, description, photo_url, created_at FROM psychologists WHERE id = $1`
	p := &entities.Psychologist{}
	err := r.db.QueryRow(query, id).Scan(&p.ID, &p.Name, &p.Description, &p.PhotoURL, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}
