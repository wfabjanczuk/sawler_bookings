package main

import (
	"github.com/wfabjanczuk/sawler_bookings/internal/helpers"
	"net/http"

	"github.com/justinas/nosurf"
)

func NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   app.InProduction,
		SameSite: http.SameSiteLaxMode,
	})

	return csrfHandler
}

func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}

func Authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !helpers.IsAuthenticated(r) {
			session.Put(r.Context(), "error", "You must log in first")
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func Authorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !helpers.IsAuthorized(r) {
			session.Put(r.Context(), "error", "Not authorized")
			http.Redirect(w, r, "/admin/reservations-calendar", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
