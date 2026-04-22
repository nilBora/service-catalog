package store

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

// CreateCategory inserts a new category
func (s *DB) CreateCategory(ctx context.Context, c *Category) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	c.CreatedAt = time.Now()
	res, err := s.db.NamedExecContext(ctx,
		`INSERT INTO categories (name, sort_order, created_at) VALUES (:name, :sort_order, :created_at)`, c)
	if err != nil {
		return fmt.Errorf("create category: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}
	c.ID = id
	return nil
}

// GetCategory returns a category by ID
func (s *DB) GetCategory(ctx context.Context, id int64) (*Category, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var c Category
	err := s.db.GetContext(ctx, &c, `SELECT * FROM categories WHERE id = ?`, id)
	if err != nil {
		return nil, fmt.Errorf("get category %d: %w", id, err)
	}
	return &c, nil
}

// ListManagedCategories returns all categories ordered by sort_order then name
func (s *DB) ListManagedCategories(ctx context.Context) ([]Category, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var cats []Category
	err := s.db.SelectContext(ctx, &cats, `SELECT * FROM categories ORDER BY sort_order, name`)
	if err != nil {
		return nil, fmt.Errorf("list managed categories: %w", err)
	}
	return cats, nil
}

// UpdateCategory updates name and sort_order of a category
func (s *DB) UpdateCategory(ctx context.Context, c *Category) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.NamedExecContext(ctx,
		`UPDATE categories SET name=:name, sort_order=:sort_order WHERE id=:id`, c)
	if err != nil {
		return fmt.Errorf("update category %d: %w", c.ID, err)
	}
	return nil
}

// ReorderCategories sets sort_order for each category by its position in the ids slice
func (s *DB) ReorderCategories(ctx context.Context, ids []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, idStr := range ids {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			continue
		}
		if _, err := s.db.ExecContext(ctx, `UPDATE categories SET sort_order=? WHERE id=?`, i, id); err != nil {
			return fmt.Errorf("reorder category %d: %w", id, err)
		}
	}
	return nil
}

// DeleteCategory removes a category by ID
func (s *DB) DeleteCategory(ctx context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.ExecContext(ctx, `DELETE FROM categories WHERE id=?`, id)
	if err != nil {
		return fmt.Errorf("delete category %d: %w", id, err)
	}
	return nil
}
