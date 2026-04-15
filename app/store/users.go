package store

import (
	"context"
	"fmt"
	"time"
)

// CreateUser inserts a new user
func (s *DB) CreateUser(ctx context.Context, u *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO users (username, password_hash, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		u.Username, u.PasswordHash, u.CreatedAt, u.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}
	u.ID = id
	return nil
}

// GetUserByUsername returns a user by username
func (s *DB) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var u User
	err := s.db.GetContext(ctx, &u, `SELECT * FROM users WHERE username = ?`, username)
	if err != nil {
		return nil, fmt.Errorf("get user by username %q: %w", username, err)
	}
	return &u, nil
}

// GetUserByID returns a user by ID
func (s *DB) GetUserByID(ctx context.Context, id int64) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var u User
	err := s.db.GetContext(ctx, &u, `SELECT * FROM users WHERE id = ?`, id)
	if err != nil {
		return nil, fmt.Errorf("get user by id %d: %w", id, err)
	}
	return &u, nil
}

// CountUsers returns total number of users
func (s *DB) CountUsers(ctx context.Context) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var count int
	err := s.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM users`)
	if err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}
	return count, nil
}

// CreateSession inserts a new session
func (s *DB) CreateSession(ctx context.Context, sess *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess.CreatedAt = time.Now()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO sessions (id, user_id, expires_at, created_at) VALUES (?, ?, ?, ?)`,
		sess.ID, sess.UserID, sess.ExpiresAt, sess.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

// GetSession returns a session by ID
func (s *DB) GetSession(ctx context.Context, id string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var sess Session
	err := s.db.GetContext(ctx, &sess, `SELECT * FROM sessions WHERE id = ?`, id)
	if err != nil {
		return nil, fmt.Errorf("get session %q: %w", id, err)
	}
	return &sess, nil
}

// DeleteSession removes a session by ID
func (s *DB) DeleteSession(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

// DeleteExpiredSessions removes all expired sessions
func (s *DB) DeleteExpiredSessions(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE expires_at < ?`, time.Now())
	if err != nil {
		return fmt.Errorf("delete expired sessions: %w", err)
	}
	return nil
}

// DeleteUserSessions removes all sessions for a user
func (s *DB) DeleteUserSessions(ctx context.Context, userID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE user_id = ?`, userID)
	if err != nil {
		return fmt.Errorf("delete user sessions: %w", err)
	}
	return nil
}
