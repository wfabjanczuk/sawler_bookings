package handlers

import (
	"github.com/wfabjanczuk/sawler_bookings/internal/forms"
	"github.com/wfabjanczuk/sawler_bookings/internal/helpers"
	"github.com/wfabjanczuk/sawler_bookings/internal/models"
	"github.com/wfabjanczuk/sawler_bookings/internal/render"
	"net/http"
)

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

	id, accessLevel, err := m.DB.Authenticate(r.Form.Get("email"), r.Form.Get("password"))
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Invalid email or password")
		form.Set("password", "")
		render.Template(w, r, "login.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}

	m.App.Session.Put(r.Context(), "user_id", id)
	m.App.Session.Put(r.Context(), "access_level", accessLevel)
	m.App.Session.Put(r.Context(), "flash", "Successfully logged in")
	http.Redirect(w, r, "/admin/reservations-calendar", http.StatusSeeOther)
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
