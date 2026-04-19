// Package web provides HTTP handlers for the web UI
package web

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/nilBora/service-catalog/app/store"
)

//go:embed static
var staticFS embed.FS

//go:embed templates
var templatesFS embed.FS

// StaticFS returns the embedded static filesystem
func StaticFS() (fs.FS, error) {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		return nil, fmt.Errorf("failed to get static sub-filesystem: %w", err)
	}
	return sub, nil
}

// Handler handles web UI requests
type Handler struct {
	store store.Store
	tmpl  *template.Template
}

// New creates a new web handler
func New(st store.Store) (*Handler, error) {
	tmpl, err := parseTemplates()
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &Handler{
		store: st,
		tmpl:  tmpl,
	}, nil
}

// Register registers web UI routes on the given router
func (h *Handler) Register(r chi.Router) {
	// Public routes
	r.Get("/login", h.handleLogin)
	r.Post("/login", h.handleLoginPost)
	r.Get("/setup", h.handleSetup)
	r.Post("/setup", h.handleSetupPost)
	r.Get("/logout", h.handleLogout)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(h.AuthMiddleware)

		r.Get("/", h.handleDashboard)
		r.Get("/services", h.handleServices)

		// Services CRUD (HTMX)
		r.Get("/web/services/panel", h.handleServicesPanel)
		r.Get("/web/services", h.handleServiceTable)
		r.Get("/web/services/new", h.handleServiceForm)
		r.Get("/web/services/{id}/edit", h.handleServiceEditForm)
		r.Post("/web/services", h.handleServiceCreate)
		r.Put("/web/services/{id}", h.handleServiceUpdate)
		r.Put("/web/services/{id}/status", h.handleServiceStatusUpdate)
		r.Delete("/web/services/{id}", h.handleServiceDelete)

		// Categories CRUD (HTMX)
		r.Get("/web/categories/panel", h.handleCategoriesPanel)
		r.Get("/web/categories", h.handleCategoryList)
		r.Get("/web/categories/new", h.handleCategoryForm)
		r.Get("/web/categories/{id}/edit", h.handleCategoryEditForm)
		r.Post("/web/categories", h.handleCategoryCreate)
		r.Put("/web/categories/{id}", h.handleCategoryUpdate)
		r.Delete("/web/categories/{id}", h.handleCategoryDelete)

		// Dashboard
		r.Get("/web/dashboard", h.handleDashboardContent)
		r.Get("/web/dashboard/stats", h.handleDashboardStats)

		// Theme
		r.Post("/web/theme", h.handleThemeToggle)
	})
}

// templateFuncs returns custom template functions
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.Format("2006-01-02 15:04")
		},
		"title": func(s string) string {
			if len(s) == 0 {
				return s
			}
			return strings.ToUpper(s[:1]) + s[1:]
		},
		"statusClass": func(status string) string {
			switch status {
			case "active":
				return "status-active"
			case "paused":
				return "status-paused"
			}
			return ""
		},
		"hasLogo": func(logoURL string) bool {
			return logoURL != "" && !strings.HasPrefix(logoURL, "#")
		},
		"isEmoji": func(logoURL string) bool {
			return strings.HasPrefix(logoURL, "#")
		},
		"emojiValue": func(logoURL string) string {
			return strings.TrimPrefix(logoURL, "#")
		},
		"categoryColor": func(category string) string {
			colors := map[string]string{
				"Monitoring":    "#059669",
				"CRM":           "#0062ff",
				"Documentation": "#d97706",
				"Internal":      "#7c3aed",
				"Infrastructure": "#dc2626",
				"Analytics":     "#0891b2",
				"Other":         "#6b7280",
			}
			if c, ok := colors[category]; ok {
				return c
			}
			return "#6b7280"
		},
	}
}

// parseTemplates parses all templates from embedded filesystem
func parseTemplates() (*template.Template, error) {
	tmpl := template.New("").Funcs(templateFuncs())

	partials := []string{
		"nav",
		"service-card",
		"service-form",
		"service-table",
		"services-panel",
		"category-list",
		"category-form",
		"dashboard-stats",
		"dashboard-services",
	}

	for _, name := range partials {
		content, err := templatesFS.ReadFile("templates/partials/" + name + ".html")
		if err != nil {
			return nil, fmt.Errorf("read partial %s: %w", name, err)
		}
		_, err = tmpl.New(name).Parse(string(content))
		if err != nil {
			return nil, fmt.Errorf("parse partial %s: %w", name, err)
		}
	}

	pages := []string{
		"dashboard.html",
		"services.html",
		"login.html",
		"setup.html",
	}

	for _, name := range pages {
		content, err := templatesFS.ReadFile("templates/" + name)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", name, err)
		}
		_, err = tmpl.New(name).Parse(string(content))
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", name, err)
		}
	}

	return tmpl, nil
}

// templateData holds common data passed to templates
type templateData struct {
	Theme      string
	ActivePage string
	Error      string
	Success    string

	// dashboard
	Stats      *store.DashboardStats
	Grouped    map[string][]store.Service
	Categories []string

	// categories
	ManagedCategories []store.Category
	Category          *store.Category

	// services list
	Services []store.Service
	Service  *store.Service

	// filter
	StatusFilter   string
	CategoryFilter string
}

// getTheme returns the current theme from cookie
func (h *Handler) getTheme(r *http.Request) string {
	if c, err := r.Cookie("theme"); err == nil {
		switch c.Value {
		case "dark", "dark-electric", "dark-cyber":
			return c.Value
		}
	}
	return ""
}

// parseID parses an ID from the URL parameter
func parseID(r *http.Request, param string) (int64, error) {
	idStr := chi.URLParam(r, param)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid ID: %s", idStr)
	}
	return id, nil
}

// renderError renders an error response
func (h *Handler) renderError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	w.Write([]byte(fmt.Sprintf(`<div class="error">%s</div>`, message)))
}

// render executes a template with the given data
func (h *Handler) render(w http.ResponseWriter, r *http.Request, name string, data templateData) {
	data.Theme = h.getTheme(r)
	if err := h.tmpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
