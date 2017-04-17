package utils

import (
	"net/http"
	"time"
)

// SetCookie sets a cookie
func SetCookie(w http.ResponseWriter, name, value string, days int) {
	http.SetCookie(w, &http.Cookie{
		Name:    name,
		Value:   value,
		Expires: time.Now().AddDate(0, 0, days),
		Path:    "/",
	})
}

// GetCookie gets a cookie, or an empty string as a fallback
func GetCookie(r *http.Request, name string) string {
	cookie, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// DeleteCookie deletes a cookie by causing it to expire
func DeleteCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:    name,
		Expires: time.Now().AddDate(0, 0, -1),
		Path:    "/",
	})
}
