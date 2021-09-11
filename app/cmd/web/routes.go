package main

import (
	"github.com/wfabjanczuk/sawler_bookings/internal/config"
	"github.com/wfabjanczuk/sawler_bookings/internal/handlers"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func routes(app *config.AppConfig) http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(NoSurf)
	mux.Use(SessionLoad)

	mux.Get("/", handlers.Repo.Home)
	mux.Post("/", handlers.Repo.Home)
	mux.Get("/about", handlers.Repo.About)
	mux.Get("/contact", handlers.Repo.Contact)

	mux.Get("/generals-quarters", handlers.Repo.Generals)
	mux.Get("/majors-suite", handlers.Repo.Majors)

	mux.Get("/search-availability", handlers.Repo.Availability)
	mux.Post("/search-availability", handlers.Repo.PostAvailability)
	mux.Post("/search-availability-json", handlers.Repo.AvailabilityJson)

	mux.Get("/choose-room/{id}", handlers.Repo.ChooseRoom)
	mux.Get("/book-room", handlers.Repo.BookRoom)

	mux.Get("/make-reservation", handlers.Repo.Reservation)
	mux.Post("/make-reservation", handlers.Repo.PostReservation)
	mux.Get("/reservation-summary", handlers.Repo.ReservationSummary)

	mux.Get("/user/login", handlers.Repo.ShowLogin)
	mux.Post("/user/login", handlers.Repo.PostShowLogin)
	mux.Get("/user/logout", handlers.Repo.Logout)

	mux.Route("/admin", func(mux chi.Router) {
		mux.Use(Authentication)
		mux.Use(Authorization)

		mux.Get("/reservations-new", handlers.Repo.AdminReservationsNew)
		mux.Get("/reservations-all", handlers.Repo.AdminReservationsAll)
		mux.Get("/reservations-calendar", handlers.Repo.AdminReservationsCalendar)
		mux.Get("/reservations/{src}/{id}", handlers.Repo.AdminShowReservation)

		mux.Post("/reservations-calendar", handlers.Repo.AdminPostReservationsCalendar)
		mux.Post("/reservations/{src}/{id}", handlers.Repo.AdminPostReservation)
		mux.Post("/process-reservation/{src}/{id}", handlers.Repo.AdminProcessReservation)
		mux.Post("/delete-reservation/{src}/{id}", handlers.Repo.AdminDeleteReservation)
	})

	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}
