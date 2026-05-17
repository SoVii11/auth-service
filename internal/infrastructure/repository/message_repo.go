package repository

import (
	"database/sql"
	"time"

	"github.com/SoVii11/auth-service/internal/entities"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(msg *entities.Message) error {
	query := `
		INSERT INTO messages (user_id, text, is_admin, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id`
	return r.db.QueryRow(query, msg.UserID, msg.Text, msg.IsAdmin, time.Now()).Scan(&msg.ID)
}

func (r *MessageRepository) GetByUserID(userID int64) ([]entities.Message, error) {
	query := `
		SELECT id, user_id, text, is_admin, created_at 
		FROM messages 
		WHERE user_id = $1 
		ORDER BY created_at ASC`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []entities.Message
	for rows.Next() {
		var m entities.Message
		if err := rows.Scan(&m.ID, &m.UserID, &m.Text, &m.IsAdmin, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, nil
}

func (r *MessageRepository) GetAllChats() ([]entities.Message, error) {
	query := `
		SELECT DISTINCT ON (user_id) id, user_id, text, is_admin, created_at
		FROM messages
		ORDER BY user_id, created_at DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []entities.Message
	for rows.Next() {
		var m entities.Message
		if err := rows.Scan(&m.ID, &m.UserID, &m.Text, &m.IsAdmin, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, nil
}
