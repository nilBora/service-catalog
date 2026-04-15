package store

import "context"

// ServiceStore defines operations for services
type ServiceStore interface {
	CreateService(ctx context.Context, s *Service) error
	GetService(ctx context.Context, id int64) (*Service, error)
	ListServices(ctx context.Context) ([]Service, error)
	ListServicesByCategory(ctx context.Context, category string) ([]Service, error)
	ListServicesByStatus(ctx context.Context, status string) ([]Service, error)
	UpdateService(ctx context.Context, s *Service) error
	UpdateServiceStatus(ctx context.Context, id int64, status string) error
	DeleteService(ctx context.Context, id int64) error
	GetDashboardStats(ctx context.Context) (*DashboardStats, error)
	ListCategories(ctx context.Context) ([]string, error)
	GetServicesGroupedByCategory(ctx context.Context) (map[string][]Service, []string, error)
}

// UserStore defines operations for users
type UserStore interface {
	CreateUser(ctx context.Context, u *User) error
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	GetUserByID(ctx context.Context, id int64) (*User, error)
	CountUsers(ctx context.Context) (int, error)
}

// SessionStore defines operations for sessions
type SessionStore interface {
	CreateSession(ctx context.Context, s *Session) error
	GetSession(ctx context.Context, id string) (*Session, error)
	DeleteSession(ctx context.Context, id string) error
	DeleteExpiredSessions(ctx context.Context) error
	DeleteUserSessions(ctx context.Context, userID int64) error
}

// Store combines all store interfaces
type Store interface {
	ServiceStore
	UserStore
	SessionStore
	Close() error
}
