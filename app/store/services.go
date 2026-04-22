package store

import (
	"context"
	"fmt"
	"time"
)

// CreateService inserts a new service
func (s *DB) CreateService(ctx context.Context, svc *Service) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `INSERT INTO services (name, url, description, category, logo_url, color, status, sort_order, created_at, updated_at)
		VALUES (:name, :url, :description, :category, :logo_url, :color, :status, :sort_order, :created_at, :updated_at)`

	svc.CreatedAt = time.Now()
	svc.UpdatedAt = time.Now()
	if svc.Status == "" {
		svc.Status = "active"
	}

	res, err := s.db.NamedExecContext(ctx, query, svc)
	if err != nil {
		return fmt.Errorf("create service: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}
	svc.ID = id
	return nil
}

// GetService returns a service by ID
func (s *DB) GetService(ctx context.Context, id int64) (*Service, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var svc Service
	err := s.db.GetContext(ctx, &svc, `SELECT * FROM services WHERE id = ?`, id)
	if err != nil {
		return nil, fmt.Errorf("get service %d: %w", id, err)
	}
	return &svc, nil
}

// ListServices returns all services ordered by category and sort_order
func (s *DB) ListServices(ctx context.Context) ([]Service, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var services []Service
	err := s.db.SelectContext(ctx, &services, `SELECT * FROM services ORDER BY category, sort_order, name`)
	if err != nil {
		return nil, fmt.Errorf("list services: %w", err)
	}
	return services, nil
}

// ListServicesByCategory returns services filtered by category
func (s *DB) ListServicesByCategory(ctx context.Context, category string) ([]Service, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var services []Service
	err := s.db.SelectContext(ctx, &services, `SELECT * FROM services WHERE category = ? ORDER BY sort_order, name`, category)
	if err != nil {
		return nil, fmt.Errorf("list services by category: %w", err)
	}
	return services, nil
}

// ListServicesByStatus returns services filtered by status
func (s *DB) ListServicesByStatus(ctx context.Context, status string) ([]Service, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var services []Service
	err := s.db.SelectContext(ctx, &services, `SELECT * FROM services WHERE status = ? ORDER BY category, sort_order, name`, status)
	if err != nil {
		return nil, fmt.Errorf("list services by status: %w", err)
	}
	return services, nil
}

// UpdateService updates an existing service
func (s *DB) UpdateService(ctx context.Context, svc *Service) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	svc.UpdatedAt = time.Now()
	query := `UPDATE services SET name=:name, url=:url, description=:description, category=:category,
		logo_url=:logo_url, color=:color, status=:status, sort_order=:sort_order, updated_at=:updated_at
		WHERE id=:id`

	_, err := s.db.NamedExecContext(ctx, query, svc)
	if err != nil {
		return fmt.Errorf("update service %d: %w", svc.ID, err)
	}
	return nil
}

// UpdateServiceStatus updates only the status field
func (s *DB) UpdateServiceStatus(ctx context.Context, id int64, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.ExecContext(ctx, `UPDATE services SET status=?, updated_at=? WHERE id=?`, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("update service status %d: %w", id, err)
	}
	return nil
}

// DeleteService removes a service by ID
func (s *DB) DeleteService(ctx context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.ExecContext(ctx, `DELETE FROM services WHERE id=?`, id)
	if err != nil {
		return fmt.Errorf("delete service %d: %w", id, err)
	}
	return nil
}

// GetDashboardStats returns aggregated stats
func (s *DB) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &DashboardStats{}

	err := s.db.GetContext(ctx, &stats.Total, `SELECT COUNT(*) FROM services`)
	if err != nil {
		return nil, fmt.Errorf("count total: %w", err)
	}
	err = s.db.GetContext(ctx, &stats.Active, `SELECT COUNT(*) FROM services WHERE status='active'`)
	if err != nil {
		return nil, fmt.Errorf("count active: %w", err)
	}
	err = s.db.GetContext(ctx, &stats.Paused, `SELECT COUNT(*) FROM services WHERE status='paused'`)
	if err != nil {
		return nil, fmt.Errorf("count paused: %w", err)
	}

	rows, err := s.db.QueryxContext(ctx, `SELECT category, COUNT(*) as count FROM services GROUP BY category ORDER BY category`)
	if err != nil {
		return nil, fmt.Errorf("count by category: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cc CategoryCount
		if err := rows.Scan(&cc.Category, &cc.Count); err != nil {
			return nil, fmt.Errorf("scan category count: %w", err)
		}
		stats.ByCategory = append(stats.ByCategory, cc)
	}

	return stats, nil
}

// ListCategories returns all distinct categories
func (s *DB) ListCategories(ctx context.Context) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var categories []string
	err := s.db.SelectContext(ctx, &categories, `SELECT DISTINCT category FROM services ORDER BY category`)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	return categories, nil
}

// GetServicesGroupedByCategory returns services grouped by category, ordered by categories.sort_order
func (s *DB) GetServicesGroupedByCategory(ctx context.Context) (map[string][]Service, []string, error) {
	services, err := s.ListServices(ctx)
	if err != nil {
		return nil, nil, err
	}

	managedCats, err := s.ListManagedCategories(ctx)
	if err != nil {
		return nil, nil, err
	}

	grouped := make(map[string][]Service)
	for _, svc := range services {
		grouped[svc.Category] = append(grouped[svc.Category], svc)
	}

	// Managed categories first (in sort_order), then orphan categories
	seen := make(map[string]bool)
	var orderedCats []string
	for _, cat := range managedCats {
		if _, exists := grouped[cat.Name]; exists {
			orderedCats = append(orderedCats, cat.Name)
			seen[cat.Name] = true
		}
	}
	for cat := range grouped {
		if !seen[cat] {
			orderedCats = append(orderedCats, cat)
		}
	}

	return grouped, orderedCats, nil
}
