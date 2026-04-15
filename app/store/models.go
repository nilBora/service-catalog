package store

import "time"

// Service represents a service/microservice entry
type Service struct {
	ID          int64     `db:"id"`
	Name        string    `db:"name"`
	URL         string    `db:"url"`
	Description string    `db:"description"`
	Category    string    `db:"category"`
	LogoURL     string    `db:"logo_url"`
	Color       string    `db:"color"`
	Status      string    `db:"status"` // active, paused
	SortOrder   int       `db:"sort_order"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// Category represents a service category
type Category struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	SortOrder int       `db:"sort_order"`
	CreatedAt time.Time `db:"created_at"`
}

// User represents a user account
type User struct {
	ID           int64     `db:"id"`
	Username     string    `db:"username"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// Session represents an active user session
type Session struct {
	ID        string    `db:"id"`
	UserID    int64     `db:"user_id"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}

// DashboardStats holds aggregated dashboard statistics
type DashboardStats struct {
	Total   int
	Active  int
	Paused  int
	ByCategory []CategoryCount
}

// CategoryCount holds service count per category
type CategoryCount struct {
	Category string
	Count    int
}
