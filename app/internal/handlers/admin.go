package handlers

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/wfabjanczuk/sawler_bookings/internal/constants"
	"github.com/wfabjanczuk/sawler_bookings/internal/forms"
	"github.com/wfabjanczuk/sawler_bookings/internal/helpers"
	"github.com/wfabjanczuk/sawler_bookings/internal/models"
	"github.com/wfabjanczuk/sawler_bookings/internal/render"
	"net/http"
	"strconv"
	"strings"
	"time"
)

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

func (m *Repository) AdminPostReservation(w http.ResponseWriter, r *http.Request) {
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

	err = r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't parse form")
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations/%d/%s", id, src), http.StatusSeeOther)
		return
	}

	reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName = r.Form.Get("last_name")
	reservation.Email = r.Form.Get("email")
	reservation.Phone = r.Form.Get("phone")

	err = m.DB.UpdateReservation(reservation)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	m.App.Session.Put(r.Context(), "flash", "Reservation updated")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
}

func (m *Repository) getCalendarTime(r *http.Request) (time.Time, error) {
	var yearString, monthString string
	now := time.Now().UTC()

	if r.URL.Query().Get("y") == "" {
		yearString = m.App.Session.GetString(r.Context(), "calendar_current_year")
		monthString = m.App.Session.GetString(r.Context(), "calendar_current_month")
	} else {
		yearString = r.URL.Query().Get("y")
		monthString = r.URL.Query().Get("m")
	}

	currentDate, err := time.Parse("2006-01", fmt.Sprintf("%s-%s", yearString, monthString))
	if err != nil {
		return now, err
	}

	return currentDate.UTC(), nil
}

func (m *Repository) AdminReservationsCalendar(w http.ResponseWriter, r *http.Request) {
	currentDate, err := m.getCalendarTime(r)
	if err != nil {
		m.App.ErrorLog.Println(err)
	}

	currentYear := currentDate.Format("2006")
	currentMonth := currentDate.Format("01")

	if r.URL.Query().Get("y") == "" || err != nil {
		http.Redirect(w, r,
			fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", currentYear, currentMonth),
			http.StatusSeeOther,
		)
		return
	}

	m.App.Session.Put(r.Context(), "calendar_current_year", currentYear)
	m.App.Session.Put(r.Context(), "calendar_current_month", currentMonth)

	next := currentDate.AddDate(0, 1, 0)
	previous := currentDate.AddDate(0, -1, 0)

	currentLocation := currentDate.Location()
	firstDayOfMonth := time.Date(currentDate.Year(), currentDate.Month(), 1, 0, 0, 0, 0, currentLocation)
	lastDayOfMonth := firstDayOfMonth.AddDate(0, 1, -1)

	rooms, err := m.DB.AllRooms()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := map[string]interface{}{
		"now":         time.Now().UTC(),
		"currentDate": currentDate,
		"rooms":       rooms,
		"weeks":       helpers.GetMonthWeeks(firstDayOfMonth, lastDayOfMonth, time.Sunday),
		"weekDays":    helpers.GetWeekDays(),
	}

	for _, room := range rooms {
		reservationMap := make(map[string]int)
		blockMap := make(map[string]int)

		for day := firstDayOfMonth; !day.After(lastDayOfMonth); day = helpers.NextDay(day) {
			reservationMap[day.Format(constants.DefaultDateFormat)] = 0
			blockMap[day.Format(constants.DefaultDateFormat)] = 0
		}

		restrictions, err := m.DB.GetRoomRestrictionsByDate(room.ID, firstDayOfMonth, lastDayOfMonth)
		if err != nil {
			helpers.ServerError(w, err)
			return
		}

		for _, restriction := range restrictions {
			for day := restriction.StartDate; !day.After(restriction.EndDate); day = helpers.NextDay(day) {
				if restriction.ReservationID > 0 {
					reservationMap[day.Format(constants.DefaultDateFormat)] = restriction.ReservationID
				} else {
					blockMap[day.Format(constants.DefaultDateFormat)] = restriction.ID
				}
			}
		}

		data[fmt.Sprintf("reservation_map_%d", room.ID)] = reservationMap
		data[fmt.Sprintf("block_map_%d", room.ID)] = blockMap

		m.App.Session.Put(r.Context(), fmt.Sprintf("block_map_%d", room.ID), blockMap)
	}

	render.Template(w, r, "admin-reservations-calendar.page.tmpl", &models.TemplateData{
		StringMap: helpers.GetCalendarStringMap(previous, currentDate, next),
		Data:      data,
	})
}

func (m *Repository) AdminPostReservationsCalendar(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	currentDate, err := time.Parse("2006-01", fmt.Sprintf("%s-%s", r.PostForm.Get("y"), r.PostForm.Get("m")))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	currentYear := currentDate.Format("2006")
	currentMonth := currentDate.Format("01")

	rooms, err := m.DB.AllRooms()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	form := forms.New(r.PostForm)

	for _, room := range rooms {
		currentBlockMap, ok := m.App.Session.Get(r.Context(), fmt.Sprintf("block_map_%d", room.ID)).(map[string]int)
		if !ok {
			helpers.ServerError(w, errors.New("cannot get block map from session"))
			return
		}

		for key, value := range currentBlockMap {
			if value > 0 && !form.Has(fmt.Sprintf("remove_block_%d_%s", room.ID, key)) {
				err = m.DB.DeleteRoomRestriction(value)

				if err != nil {
					helpers.ServerError(w, err)
					return
				}
			}
		}
	}

	for name, _ := range r.PostForm {
		if strings.HasPrefix(name, "add_block") {
			exploded := strings.Split(name, "_")
			roomID, err := strconv.Atoi(exploded[2])

			if err != nil {
				helpers.ServerError(w, err)
				return
			}

			dateString := exploded[3]
			layout := constants.DefaultDateFormat
			date, err := time.Parse(layout, dateString)

			if err != nil {
				helpers.ServerError(w, err)
				return
			}

			_, err = m.DB.InsertRoomRestriction(models.RoomRestriction{
				RoomID:        roomID,
				RestrictionID: 2,
				StartDate:     date,
				EndDate:       date,
			})

			if err != nil {
				helpers.ServerError(w, err)
				return
			}
		}
	}

	m.App.Session.Put(r.Context(), "flash", "Calendar changes saved")
	redirectUrl := fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", currentYear, currentMonth)
	http.Redirect(w, r, redirectUrl, http.StatusSeeOther)
}

func (m *Repository) AdminProcessReservation(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	src := chi.URLParam(r, "src")

	err = m.DB.UpdateReservationProcessed(id, 1)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	m.App.Session.Put(r.Context(), "flash", "Reservation marked as processed")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
}

func (m *Repository) AdminDeleteReservation(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	src := chi.URLParam(r, "src")

	err = m.DB.DeleteReservation(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	m.App.Session.Put(r.Context(), "flash", "Reservation deleted")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
}
