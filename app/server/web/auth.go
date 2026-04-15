package web

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	log "github.com/go-pkgz/lgr"
	"golang.org/x/crypto/bcrypt"

	"github.com/nilBora/service-catalog/app/store"
)

// AuthMiddleware checks if user is authenticated
func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("session")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		sess, err := h.store.GetSession(r.Context(), c.Value)
		if err != nil || sess.ExpiresAt.Before(time.Now()) {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// handleLogin renders the login page
func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	// Check if setup is needed
	count, err := h.store.CountUsers(r.Context())
	if err != nil || count == 0 {
		http.Redirect(w, r, "/setup", http.StatusFound)
		return
	}

	h.render(w, r, "login.html", templateData{})
}

// handleLoginPost processes login form
func (h *Handler) handleLoginPost(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	user, err := h.store.GetUserByUsername(r.Context(), username)
	if err != nil {
		h.render(w, r, "login.html", templateData{Error: "Invalid username or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		h.render(w, r, "login.html", templateData{Error: "Invalid username or password"})
		return
	}

	sessionID := uuid.New().String()
	sess := &store.Session{
		ID:        sessionID,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	if err := h.store.CreateSession(r.Context(), sess); err != nil {
		h.render(w, r, "login.html", templateData{Error: "Failed to create session"})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Expires:  sess.ExpiresAt,
	})

	http.Redirect(w, r, "/", http.StatusFound)
}

// handleSetup renders the initial setup page
func (h *Handler) handleSetup(w http.ResponseWriter, r *http.Request) {
	count, err := h.store.CountUsers(r.Context())
	if err == nil && count > 0 {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	h.render(w, r, "setup.html", templateData{})
}

// handleSetupPost creates the initial admin user
func (h *Handler) handleSetupPost(w http.ResponseWriter, r *http.Request) {
	count, err := h.store.CountUsers(r.Context())
	if err == nil && count > 0 {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		h.render(w, r, "setup.html", templateData{Error: "Username and password are required"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		h.render(w, r, "setup.html", templateData{Error: "Failed to hash password"})
		return
	}

	user := &store.User{
		Username:     username,
		PasswordHash: string(hash),
	}

	if err := h.store.CreateUser(r.Context(), user); err != nil {
		h.render(w, r, "setup.html", templateData{Error: "Failed to create user"})
		return
	}

	log.Printf("[INFO] created admin user %q", username)
	http.Redirect(w, r, "/login", http.StatusFound)
}

// handleLogout destroys the session
func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie("session"); err == nil {
		_ = h.store.DeleteSession(r.Context(), c.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "session",
		Value:   "",
		Path:    "/",
		MaxAge:  -1,
		Expires: time.Unix(0, 0),
	})
	http.Redirect(w, r, "/login", http.StatusFound)
}

// handleThemeToggle cycles through themes
func (h *Handler) handleThemeToggle(w http.ResponseWriter, r *http.Request) {
	current := h.getTheme(r)
	var next string
	switch current {
	case "":
		next = "dark"
	case "dark":
		next = "dark-electric"
	case "dark-electric":
		next = "dark-cyber"
	default:
		next = ""
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "theme",
		Value:   next,
		Path:    "/",
		MaxAge:  365 * 24 * 3600,
		Expires: time.Now().Add(365 * 24 * time.Hour),
	})

	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusNoContent)
}
