package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type Repo struct {
	db *DB
}

func NewRepo(db *DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) CreateUser(ctx context.Context, username string, passwordHash string) (User, error) {
	var u User
	err := r.db.Pool.QueryRow(ctx, `
		insert into users (username, password_hash)
		values ($1, $2)
		returning id, username, password_hash, created_at
	`, username, passwordHash).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

func (r *Repo) GetUserByUsername(ctx context.Context, username string) (User, error) {
	var u User
	err := r.db.Pool.QueryRow(ctx, `
		select id, username, password_hash, created_at
		from users
		where username = $1
	`, username).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return User{}, fmt.Errorf("not found")
		}
		return User{}, fmt.Errorf("get user: %w", err)
	}
	return u, nil
}

func (r *Repo) GetUserByID(ctx context.Context, id int64) (User, error) {
	var u User
	err := r.db.Pool.QueryRow(ctx, `
		select id, username, password_hash, created_at
		from users
		where id = $1
	`, id).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return User{}, fmt.Errorf("not found")
		}
		return User{}, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

func HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func (r *Repo) StoreRefreshToken(ctx context.Context, userID int64, rawToken string, expiresAt time.Time) error {
	_, err := r.db.Pool.Exec(ctx, `
		insert into refresh_tokens (user_id, token_hash, expires_at)
		values ($1, $2, $3)
	`, userID, HashToken(rawToken), expiresAt)
	if err != nil {
		return fmt.Errorf("store refresh: %w", err)
	}
	return nil
}

func (r *Repo) ValidateRefreshToken(ctx context.Context, userID int64, rawToken string) error {
	var revokedAt *time.Time
	var expiresAt time.Time
	err := r.db.Pool.QueryRow(ctx, `
		select revoked_at, expires_at
		from refresh_tokens
		where user_id = $1 and token_hash = $2
	`, userID, HashToken(rawToken)).Scan(&revokedAt, &expiresAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("invalid refresh token")
		}
		return fmt.Errorf("validate refresh: %w", err)
	}
	if revokedAt != nil {
		return fmt.Errorf("refresh token revoked")
	}
	if time.Now().After(expiresAt) {
		return fmt.Errorf("refresh token expired")
	}
	return nil
}

func (r *Repo) RevokeRefreshToken(ctx context.Context, userID int64, rawToken string) error {
	now := time.Now()
	_, err := r.db.Pool.Exec(ctx, `
		update refresh_tokens
		set revoked_at = $3
		where user_id = $1 and token_hash = $2 and revoked_at is null
	`, userID, HashToken(rawToken), &now)
	if err != nil {
		return fmt.Errorf("revoke refresh: %w", err)
	}
	return nil
}

func (r *Repo) CreateRoom(ctx context.Context, name string) (Room, error) {
	var rm Room
	err := r.db.Pool.QueryRow(ctx, `
		insert into rooms (name)
		values ($1)
		returning id, name, created_at
	`, name).Scan(&rm.ID, &rm.Name, &rm.CreatedAt)
	if err != nil {
		return Room{}, fmt.Errorf("create room: %w", err)
	}
	return rm, nil
}

func (r *Repo) ListRooms(ctx context.Context) ([]Room, error) {
	rows, err := r.db.Pool.Query(ctx, `
		select id, name, created_at
		from rooms
		order by id asc
	`)
	if err != nil {
		return nil, fmt.Errorf("list rooms: %w", err)
	}
	defer rows.Close()

	var out []Room
	for rows.Next() {
		var rm Room
		if err := rows.Scan(&rm.ID, &rm.Name, &rm.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan room: %w", err)
		}
		out = append(out, rm)
	}
	return out, nil
}

func (r *Repo) CreateMessage(ctx context.Context, roomID int64, userID int64, text string) (Message, error) {
	var m Message
	err := r.db.Pool.QueryRow(ctx, `
		insert into messages (room_id, user_id, text)
		values ($1, $2, $3)
		returning id, room_id, user_id, text, created_at
	`, roomID, userID, text).Scan(&m.ID, &m.RoomID, &m.UserID, &m.Text, &m.CreatedAt)
	if err != nil {
		return Message{}, fmt.Errorf("create message: %w", err)
	}

	// include username for response
	err = r.db.Pool.QueryRow(ctx, `select username from users where id = $1`, userID).Scan(&m.Username)
	if err != nil {
		return Message{}, fmt.Errorf("load username: %w", err)
	}
	return m, nil
}

func (r *Repo) ListMessages(ctx context.Context, roomID int64, limit int) ([]Message, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := r.db.Pool.Query(ctx, `
		select m.id, m.room_id, m.user_id, u.username, m.text, m.created_at
		from messages m
		join users u on u.id = m.user_id
		where m.room_id = $1
		order by m.id desc
		limit $2
	`, roomID, limit)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	defer rows.Close()

	var out []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.RoomID, &m.UserID, &m.Username, &m.Text, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		out = append(out, m)
	}
	return out, nil
}
