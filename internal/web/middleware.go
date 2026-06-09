package web

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"strings"
)

const sessionCookieName = "khazaneh_session"

func (h *Handlers) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login" {
			next.ServeHTTP(w, r)
			return
		}
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil || !h.validSession(cookie.Value) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handlers) newSessionValue(username string) string {
	mac := hmac.New(sha256.New, []byte(h.cfg.DashboardPassword))
	mac.Write([]byte(username))
	signature := mac.Sum(nil)
	return base64.RawURLEncoding.EncodeToString([]byte(username)) + "." + base64.RawURLEncoding.EncodeToString(signature)
}

func (h *Handlers) validSession(value string) bool {
	parts := strings.Split(value, ".")
	if len(parts) != 2 {
		return false
	}
	usernameBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	expected := h.newSessionValue(string(usernameBytes))
	return hmac.Equal([]byte(value), []byte(expected)) && string(usernameBytes) == h.cfg.DashboardUsername
}
