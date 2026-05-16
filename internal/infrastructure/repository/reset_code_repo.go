package repository

import (
	"database/sql"
	"time"

	"github.com/SoVii11/auth-service/internal/entities"
)

type ResetCodeRepository struct {
	db *sql.DB
}

func NewResetCodeRepository(db *sql.DB) *ResetCodeRepository {
	return &ResetCodeRepository{db: db}
}

func (r *ResetCodeRepository) Create(code *entities.ResetCode) error {
	query := `
		INSERT INTO reset_codes (user_id, code, expires_at, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	return r.db.QueryRow(query,
		code.UserID,
		code.Code,
		code.ExpiresAt,
		time.Now(),
	).Scan(&code.ID)
}

func (r *ResetCodeRepository) FindByUserIDAndCode(userID int64, code string) (*entities.ResetCode, error) {
	query := `
		SELECT id, user_id, code, expires_at, created_at 
		FROM reset_codes 
		WHERE user_id = $1 AND code = $2`

	rc := &entities.ResetCode{}
	err := r.db.QueryRow(query, userID, code).Scan(
		&rc.ID,
		&rc.UserID,
		&rc.Code,
		&rc.ExpiresAt,
		&rc.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return rc, nil
}

func (r *ResetCodeRepository) DeleteByUserID(userID int64) error {
	query := `DELETE FROM reset_codes WHERE user_id = $1`
	_, err := r.db.Exec(query, userID)
	return err
}
