package web

import (
	"net/http"
	"strconv"

	"github.com/nilBora/service-catalog/app/store"
)

func (h *Handler) handleServicesPanel(w http.ResponseWriter, r *http.Request) {
	services, err := h.store.ListServices(r.Context())
	if err != nil {
		h.renderError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := h.tmpl.ExecuteTemplate(w, "services-panel", templateData{
		Services: services,
		Theme:    h.getTheme(r),
	}); err != nil {
		h.renderError(w, http.StatusInternalServerError, err.Error())
	}
}

func (h *Handler) handleCategoriesPanel(w http.ResponseWriter, r *http.Request) {
	cats, err := h.store.ListManagedCategories(r.Context())
	if err != nil {
		h.renderError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := h.tmpl.ExecuteTemplate(w, "categories-panel", templateData{
		ManagedCategories: cats,
		Theme:             h.getTheme(r),
	}); err != nil {
		h.renderError(w, http.StatusInternalServerError, err.Error())
	}
}

func (h *Handler) handleCategoryList(w http.ResponseWriter, r *http.Request) {
	cats, err := h.store.ListManagedCategories(r.Context())
	if err != nil {
		h.renderError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := h.tmpl.ExecuteTemplate(w, "category-list", templateData{
		ManagedCategories: cats,
		Theme:             h.getTheme(r),
	}); err != nil {
		h.renderError(w, http.StatusInternalServerError, err.Error())
	}
}

func (h *Handler) handleCategoryForm(w http.ResponseWriter, r *http.Request) {
	if err := h.tmpl.ExecuteTemplate(w, "category-form", templateData{
		Theme: h.getTheme(r),
	}); err != nil {
		h.renderError(w, http.StatusInternalServerError, err.Error())
	}
}

func (h *Handler) handleCategoryEditForm(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		h.renderError(w, http.StatusBadRequest, "Invalid category ID")
		return
	}
	cat, err := h.store.GetCategory(r.Context(), id)
	if err != nil {
		h.renderError(w, http.StatusNotFound, "Category not found")
		return
	}
	if err := h.tmpl.ExecuteTemplate(w, "category-form", templateData{
		Category: cat,
		Theme:    h.getTheme(r),
	}); err != nil {
		h.renderError(w, http.StatusInternalServerError, err.Error())
	}
}

func (h *Handler) handleCategoryCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.renderError(w, http.StatusBadRequest, "Invalid form data")
		return
	}
	name := r.FormValue("name")
	if name == "" {
		h.renderError(w, http.StatusBadRequest, "Name is required")
		return
	}
	sortOrder, _ := strconv.Atoi(r.FormValue("sort_order"))
	cat := &store.Category{Name: name, SortOrder: sortOrder}
	if err := h.store.CreateCategory(r.Context(), cat); err != nil {
		h.renderError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("HX-Trigger", "categoryUpdated")
	h.handleCategoryList(w, r)
}

func (h *Handler) handleCategoryUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		h.renderError(w, http.StatusBadRequest, "Invalid category ID")
		return
	}
	if err := r.ParseForm(); err != nil {
		h.renderError(w, http.StatusBadRequest, "Invalid form data")
		return
	}
	name := r.FormValue("name")
	if name == "" {
		h.renderError(w, http.StatusBadRequest, "Name is required")
		return
	}
	sortOrder, _ := strconv.Atoi(r.FormValue("sort_order"))
	cat := &store.Category{ID: id, Name: name, SortOrder: sortOrder}
	if err := h.store.UpdateCategory(r.Context(), cat); err != nil {
		h.renderError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("HX-Trigger", "categoryUpdated")
	h.handleCategoryList(w, r)
}

func (h *Handler) handleCategoryReorder(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := h.store.ReorderCategories(r.Context(), r.Form["ids"]); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) handleCategoryDelete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		h.renderError(w, http.StatusBadRequest, "Invalid category ID")
		return
	}
	if err := h.store.DeleteCategory(r.Context(), id); err != nil {
		h.renderError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("HX-Trigger", "categoryUpdated")
	h.handleCategoryList(w, r)
}
