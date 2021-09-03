package handlers

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/wfabjanczuk/sawler_bookings/internal/config"
	"github.com/wfabjanczuk/sawler_bookings/internal/driver"
	"github.com/wfabjanczuk/sawler_bookings/internal/forms"
	"github.com/wfabjanczuk/sawler_bookings/internal/helpers"
	"github.com/wfabjanczuk/sawler_bookings/internal/models"
	"github.com/wfabjanczuk/sawler_bookings/internal/render"
	"github.com/wfabjanczuk/sawler_bookings/internal/repository"
	"github.com/wfabjanczuk/sawler_bookings/internal/repository/dbrepo"
	"log"
	"net/http"
	"strconv"
	"time"
)

var Repo *Repository

type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewPostgresRepo(db.SQL, a),
	}
}

func NewHandlers(r *Repository) {
	Repo = r
}

func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "home.page.tmpl", &models.TemplateData{})
}

func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "about.page.tmpl", &models.TemplateData{})
}

func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)

	if !ok {
		helpers.ServerError(w, errors.New("Cannot get reservation from session"))
		return
	}

	var err error
	reservation.Room, err = m.DB.GetRoomById(reservation.RoomID)

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	startDate := reservation.StartDate.Format("2006-01-02")
	endDate := reservation.EndDate.Format("2006-01-02")

	render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
		Data: map[string]interface{}{
			"reservation": reservation,
		},
		StringMap: map[string]string{
			"start_date": startDate,
			"end_date":   endDate,
		},
	})
}

func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)

	if !ok {
		helpers.ServerError(w, errors.New("Cannot get reservation from session"))
		return
	}

	err := r.ParseForm()

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName = r.Form.Get("last_name")
	reservation.Email = r.Form.Get("email")
	reservation.Phone = r.Form.Get("phone")

	form := forms.New(r.PostForm)
	form.Required("first_name", "last_name", "email")
	form.MinLength("first_name", 3)
	form.IsEmail("email")

	if !form.Valid() {
		render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: map[string]interface{}{
				"reservation": reservation,
			},
		})

		return
	}

	newReservationID, err := m.DB.InsertReservation(reservation)

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	m.App.Session.Put(r.Context(), "reservation", reservation)

	_, err = m.DB.InsertRoomRestriction(models.RoomRestriction{
		RoomID:        reservation.RoomID,
		ReservationID: newReservationID,
		RestrictionID: 1,
		StartDate:     reservation.StartDate,
		EndDate:       reservation.EndDate,
	})

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	m.App.Session.Put(r.Context(), "reservation", reservation)
	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "generals.page.tmpl", &models.TemplateData{})
}

func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "majors.page.tmpl", &models.TemplateData{})
}

func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "search-availability.page.tmpl", &models.TemplateData{})
}

func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	layout := "2006-01-02"
	startDate, err := time.Parse(layout, r.Form.Get("start_date"))

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	endDate, err := time.Parse(layout, r.Form.Get("end_date"))

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	rooms, err := m.DB.SearchAvailabilityByDates(startDate, endDate)

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	if len(rooms) == 0 {
		m.App.Session.Put(r.Context(), "error", "No availability")
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})
	data["rooms"] = rooms

	currentDate := time.Now()

	if startDate.Before(currentDate) || endDate.Before(currentDate) || endDate.Before(startDate) {
		helpers.ServerError(w, errors.New("Invalid dates provided"))
		return
	}

	reservation := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}

	m.App.Session.Put(r.Context(), "reservation", reservation)

	render.Template(w, r, "choose-room.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

type availabilityJsonResponse struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
}

func (m *Repository) AvailabilityJson(w http.ResponseWriter, r *http.Request) {
	layout := "2006-01-02"
	startDate, err := time.Parse(layout, r.Form.Get("start_date"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	endDate, err := time.Parse(layout, r.Form.Get("end_date"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	roomID, err := strconv.Atoi(r.Form.Get("room_id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	isAvailable, err := m.DB.SearchAvailabilityByDatesByRoomID(startDate, endDate, roomID)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	response := availabilityJsonResponse{
		Ok:      isAvailable,
		Message: "",
	}

	out, err := json.MarshalIndent(response, "", "    ")

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	log.Println(string(out))
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "contact.page.tmpl", &models.TemplateData{})
}

func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)

	if !ok {
		m.App.ErrorLog.Println("Cannot get item from session")
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)

		return
	}

	m.App.Session.Remove(r.Context(), "reservation")
	startDate := reservation.StartDate.Format("2006-01-02")
	endDate := reservation.EndDate.Format("2006-01-02")

	render.Template(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		Data: map[string]interface{}{
			"reservation": reservation,
		},
		StringMap: map[string]string{
			"start_date": startDate,
			"end_date":   endDate,
		},
	})
}

func (m *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	roomID, err := strconv.Atoi(chi.URLParam(r, "id"))

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)

	if !ok {
		helpers.ServerError(w, errors.New("Cannot get reservation from session"))
		return
	}

	reservation.RoomID = roomID
	reservation.Room, err = m.DB.GetRoomById(roomID)

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	m.App.Session.Put(r.Context(), "reservation", reservation)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}
