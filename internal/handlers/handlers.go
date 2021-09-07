package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/wfabjanczuk/sawler_bookings/internal/config"
	"github.com/wfabjanczuk/sawler_bookings/internal/driver"
	"github.com/wfabjanczuk/sawler_bookings/internal/forms"
	"github.com/wfabjanczuk/sawler_bookings/internal/helpers"
	"github.com/wfabjanczuk/sawler_bookings/internal/models"
	"github.com/wfabjanczuk/sawler_bookings/internal/render"
	"github.com/wfabjanczuk/sawler_bookings/internal/repository"
	"github.com/wfabjanczuk/sawler_bookings/internal/repository/dbrepo"
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

func NewTestRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewTestRepo(a),
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
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	currentDate := time.Now()
	if reservation.StartDate.Before(currentDate) || reservation.EndDate.Before(currentDate) || reservation.EndDate.Before(reservation.StartDate) {
		m.App.Session.Put(r.Context(), "error", "Invalid dates provided")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	var err error
	reservation.Room, err = m.DB.GetRoomById(reservation.RoomID)

	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Room not found")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
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
	var reservation models.Reservation

	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't parse form")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName = r.Form.Get("last_name")
	reservation.Email = r.Form.Get("email")
	reservation.Phone = r.Form.Get("phone")

	layout := "2006-01-02"
	sd := r.Form.Get("start_date")
	reservation.StartDate, err = time.Parse(layout, sd)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't parse start date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	ed := r.Form.Get("end_date")
	reservation.EndDate, err = time.Parse(layout, ed)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't parse end date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	reservation.RoomID, err = strconv.Atoi(r.Form.Get("room_id"))
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Invalid data")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	reservation.Room, err = m.DB.GetRoomById(reservation.RoomID)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Room not found")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

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
			StringMap: map[string]string{
				"start_date": sd,
				"end_date":   ed,
			},
		})

		return
	}

	currentDate := time.Now()
	if reservation.StartDate.Before(currentDate) || reservation.EndDate.Before(currentDate) || reservation.EndDate.Before(reservation.StartDate) {
		m.App.Session.Put(r.Context(), "error", "Invalid dates provided")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	isAvailable, err := m.DB.SearchAvailabilityByDatesByRoomID(reservation.StartDate, reservation.EndDate, reservation.RoomID)
	if !isAvailable || err != nil {
		m.App.Session.Put(r.Context(), "error", "Room is not available anymore")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	newReservationID, err := m.DB.InsertReservation(reservation)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't insert reservation to database")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
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
		m.App.Session.Put(r.Context(), "error", "Can't insert room restriction to database")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	message := models.MailData{
		To:            reservation.Email,
		From:          "me@here.com",
		Subject:       "Reservation confirmation",
		Template:      "basic.html",
		TemplateTitle: "Reservation confirmation",
		TemplateBody: fmt.Sprintf(
			`<p class="text-center">Dear %s,</p>
<p class="text-center">This is to confirm your reservation from %s to %s in the %s room.</p>`,
			reservation.FirstName,
			reservation.StartDate.Format(layout),
			reservation.EndDate.Format(layout),
			reservation.Room.RoomName,
		),
	}

	m.App.MailChannel <- message

	message = models.MailData{
		To:            reservation.Email,
		From:          "property@owner.com",
		Subject:       "New reservation",
		Template:      "basic.html",
		TemplateTitle: "New reservation",
		TemplateBody: fmt.Sprintf(
			`<p class="text-center">Guest %s %s (email: %s) reserved %s room from %s to %s.</p>`,
			reservation.FirstName,
			reservation.LastName,
			reservation.Email,
			reservation.Room.RoomName,
			reservation.StartDate.Format(layout),
			reservation.EndDate.Format(layout),
		),
	}

	m.App.MailChannel <- message

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
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't parse form")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

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
		helpers.ServerError(w, errors.New("invalid dates provided"))
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
	Ok        bool   `json:"ok"`
	Message   string `json:"message"`
	RoomID    string `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

func (m *Repository) AvailabilityJson(w http.ResponseWriter, r *http.Request) {
	layout := "2006-01-02"

	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	sd := r.Form.Get("start_date")
	startDate, err := time.Parse(layout, sd)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	ed := r.Form.Get("end_date")
	endDate, err := time.Parse(layout, ed)
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
		response := availabilityJsonResponse{
			Ok:      false,
			Message: "Error connecting to database",
		}

		out, err := json.MarshalIndent(response, "", "    ")

		if err != nil {
			helpers.ServerError(w, err)
			return
		}

		m.App.InfoLog.Println(string(out))
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	response := availabilityJsonResponse{
		Ok:        isAvailable,
		Message:   "",
		RoomID:    strconv.Itoa(roomID),
		StartDate: sd,
		EndDate:   ed,
	}

	out, err := json.MarshalIndent(response, "", "    ")

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	m.App.InfoLog.Println(string(out))
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
		helpers.ServerError(w, errors.New("cannot get reservation from session"))
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

func (m *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {
	roomID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	layout := "2006-01-02"
	sd := r.URL.Query().Get("s")
	startDate, err := time.Parse(layout, sd)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	ed := r.URL.Query().Get("e")
	endDate, err := time.Parse(layout, ed)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	reservation := models.Reservation{
		RoomID:    roomID,
		StartDate: startDate,
		EndDate:   endDate,
	}

	reservation.Room, err = m.DB.GetRoomById(roomID)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	m.App.Session.Put(r.Context(), "reservation", reservation)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

func (m *Repository) ShowLogin(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "login.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
	})
}

func (m *Repository) PostShowLogin(w http.ResponseWriter, r *http.Request) {
	err := m.App.Session.RenewToken(r.Context())
	if err != nil {
		m.App.ErrorLog.Println(err)
	}

	err = r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("email", "password")
	form.IsEmail("email")

	if !form.Valid() {
		m.App.Session.Put(r.Context(), "error", "Invalid email or password")
		form.Set("password", "")
		render.Template(w, r, "login.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}

	id, _, err := m.DB.Authenticate(r.Form.Get("email"), r.Form.Get("password"))
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Invalid email or password")
		form.Set("password", "")
		render.Template(w, r, "login.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}

	m.App.Session.Put(r.Context(), "user_id", id)
	m.App.Session.Put(r.Context(), "flash", "Successfully logged in")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	err := m.App.Session.Destroy(r.Context())
	if err != nil {
		m.App.ErrorLog.Println(err)
	}

	err = m.App.Session.RenewToken(r.Context())
	if err != nil {
		m.App.ErrorLog.Println(err)
	}

	m.App.Session.Put(r.Context(), "flash", "Successfully logged out")
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (m *Repository) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "admin-dashboard.page.tmpl", &models.TemplateData{})
}

func (m *Repository) AdminReservationsAll(w http.ResponseWriter, r *http.Request) {
	reservations, err := m.DB.AllReservations()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	render.Template(w, r, "admin-reservations-all.page.tmpl", &models.TemplateData{
		Data: map[string]interface{}{
			"reservations": reservations,
		},
	})
}

func (m *Repository) AdminReservationsNew(w http.ResponseWriter, r *http.Request) {
	reservations, err := m.DB.NewReservations()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	render.Template(w, r, "admin-reservations-new.page.tmpl", &models.TemplateData{
		Data: map[string]interface{}{
			"reservations": reservations,
		},
	})
}

func (m *Repository) AdminShowReservation(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	reservation, err := m.DB.GetReservationById(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	src := chi.URLParam(r, "src")

	render.Template(w, r, "admin-reservation.page.tmpl", &models.TemplateData{
		Data: map[string]interface{}{
			"reservation": reservation,
		},
		StringMap: map[string]string{
			"src": src,
		},
		Form: forms.New(nil),
	})
}

func (m *Repository) AdminReservationsCalendar(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "admin-reservations-calendar.page.tmpl", &models.TemplateData{})
}
