package store

import "time"

type User struct {
	ID           int64
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}

type Room struct {
	ID        int64
	Name      string
	CreatedAt time.Time
}

type Message struct {
	ID        int64
	RoomID    int64
	UserID    int64
	Username  string
	Text      string
	CreatedAt time.Time
}

type RefreshToken struct {
	ID        int64
	UserID    int64
	TokenHash string
	RevokedAt *time.Time
	ExpiresAt time.Time
	CreatedAt time.Time
}
