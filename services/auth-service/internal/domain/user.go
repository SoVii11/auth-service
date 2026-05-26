package domain

import "time"

type User struct {
	ID        int64
	Email     string
	Password  string
	Role      string
	CreatedAt time.Time
}

type ResetCode struct {
	ID        int64
	UserID    int64
	Code      string
	ExpiresAt time.Time
	CreatedAt time.Time
}
