package web

import (
	"net/http"

	"github.com/nilBora/service-catalog/app/store"
)

// handleServiceTable returns the services table partial
func (h *Handler) handleServiceTable(w http.ResponseWriter, r *http.Request) {
	services, err := h.store.ListServices(r.Context())
	if err != nil {
		h.renderError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := h.tmpl.ExecuteTemplate(w, "service-table", templateData{
		Services: services,
		Theme:    h.getTheme(r),
	}); err != nil {
		h.renderError(w, http.StatusInternalServerError, err.Error())
	}
}

// handleServiceForm returns an empty service creation form
func (h *Handler) handleServiceForm(w http.ResponseWriter, r *http.Request) {
	cats, _ := h.store.ListManagedCategories(r.Context())
	if err := h.tmpl.ExecuteTemplate(w, "service-form", templateData{
		Theme:             h.getTheme(r),
		ManagedCategories: cats,
	}); err != nil {
		h.renderError(w, http.StatusInternalServerError, err.Error())
	}
}

// handleServiceEditForm returns a pre-filled service edit form
func (h *Handler) handleServiceEditForm(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		h.renderError(w, http.StatusBadRequest, "Invalid service ID")
		return
	}

	svc, err := h.store.GetService(r.Context(), id)
	if err != nil {
		h.renderError(w, http.StatusNotFound, "Service not found")
		return
	}

	cats, _ := h.store.ListManagedCategories(r.Context())
	if err := h.tmpl.ExecuteTemplate(w, "service-form", templateData{
		Service:           svc,
		Theme:             h.getTheme(r),
		ManagedCategories: cats,
	}); err != nil {
		h.renderError(w, http.StatusInternalServerError, err.Error())
	}
}

// handleServiceCreate creates a new service
func (h *Handler) handleServiceCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.renderError(w, http.StatusBadRequest, "Invalid form data")
		return
	}

	svc := &store.Service{
		Name:        r.FormValue("name"),
		URL:         r.FormValue("url"),
		Description: r.FormValue("description"),
		Category:    r.FormValue("category"),
		LogoURL:     r.FormValue("logo_url"),
		Color:       r.FormValue("color"),
		Status:      "active",
	}

	if svc.Name == "" || svc.URL == "" {
		h.renderError(w, http.StatusBadRequest, "Name and URL are required")
		return
	}
	if svc.Category == "" {
		svc.Category = "Other"
	}
	if svc.Color == "" {
		svc.Color = "#0062ff"
	}

	if err := h.store.CreateService(r.Context(), svc); err != nil {
		h.renderError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("HX-Trigger", "serviceUpdated")
	h.handleServiceTable(w, r)
}

// handleServiceUpdate updates an existing service
func (h *Handler) handleServiceUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		h.renderError(w, http.StatusBadRequest, "Invalid service ID")
		return
	}

	if err := r.ParseForm(); err != nil {
		h.renderError(w, http.StatusBadRequest, "Invalid form data")
		return
	}

	svc := &store.Service{
		ID:          id,
		Name:        r.FormValue("name"),
		URL:         r.FormValue("url"),
		Description: r.FormValue("description"),
		Category:    r.FormValue("category"),
		LogoURL:     r.FormValue("logo_url"),
		Color:       r.FormValue("color"),
		Status:      r.FormValue("status"),
	}

	if svc.Category == "" {
		svc.Category = "Other"
	}
	if svc.Color == "" {
		svc.Color = "#0062ff"
	}
	if svc.Status == "" {
		svc.Status = "active"
	}

	if err := h.store.UpdateService(r.Context(), svc); err != nil {
		h.renderError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("HX-Trigger", "serviceUpdated")
	h.handleServiceTable(w, r)
}

// handleServiceStatusUpdate updates only the status of a service
func (h *Handler) handleServiceStatusUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		h.renderError(w, http.StatusBadRequest, "Invalid service ID")
		return
	}

	status := r.FormValue("status")
	if status != "active" && status != "paused" {
		h.renderError(w, http.StatusBadRequest, "Invalid status")
		return
	}

	if err := h.store.UpdateServiceStatus(r.Context(), id, status); err != nil {
		h.renderError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("HX-Trigger", "serviceUpdated")
	h.handleServiceTable(w, r)
}

// handleServiceDelete removes a service
func (h *Handler) handleServiceDelete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		h.renderError(w, http.StatusBadRequest, "Invalid service ID")
		return
	}

	if err := h.store.DeleteService(r.Context(), id); err != nil {
		h.renderError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("HX-Trigger", "serviceUpdated")
	h.handleServiceTable(w, r)
}
