package web

import (
	"net/http"
)

// handleDashboard renders the main dashboard page
func (h *Handler) handleDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stats, err := h.store.GetDashboardStats(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	grouped, categories, err := h.store.GetServicesGroupedByCategory(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.render(w, r, "dashboard.html", templateData{
		ActivePage: "dashboard",
		Stats:      stats,
		Grouped:    grouped,
		Categories: categories,
	})
}

// handleDashboardContent returns dashboard content for HTMX reload
func (h *Handler) handleDashboardContent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	grouped, categories, err := h.store.GetServicesGroupedByCategory(ctx)
	if err != nil {
		h.renderError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := h.tmpl.ExecuteTemplate(w, "dashboard-services", templateData{
		Grouped:    grouped,
		Categories: categories,
		Theme:      h.getTheme(r),
	}); err != nil {
		h.renderError(w, http.StatusInternalServerError, err.Error())
	}
}

// handleDashboardStats returns stats partial for HTMX reload
func (h *Handler) handleDashboardStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stats, err := h.store.GetDashboardStats(ctx)
	if err != nil {
		h.renderError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := h.tmpl.ExecuteTemplate(w, "dashboard-stats", templateData{
		Stats: stats,
		Theme: h.getTheme(r),
	}); err != nil {
		h.renderError(w, http.StatusInternalServerError, err.Error())
	}
}

// handleServices renders the services management page
func (h *Handler) handleServices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	services, err := h.store.ListServices(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.render(w, r, "services.html", templateData{
		ActivePage: "services",
		Services:   services,
	})
}
