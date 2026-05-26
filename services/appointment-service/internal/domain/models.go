package domain

import "time"

type Appointment struct {
	ID             int64
	UserID         int64
	PsychologistID int64
	Status         string
	Comment        string
	CreatedAt      time.Time
}

type Psychologist struct {
	ID          int64
	Name        string
	Description string
	PhotoURL    string
	CreatedAt   time.Time
}

type Message struct {
	ID        int64
	UserID    int64
	Text      string
	IsAdmin   bool
	CreatedAt time.Time
}

type PsychologistSchedule struct {
	ID             int64
	PsychologistID int64
	DateFrom       time.Time
	DateTo         time.Time
	Reason         string
	CreatedAt      time.Time
}
