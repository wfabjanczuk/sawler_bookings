package main

import (
	"encoding/gob"
	"github.com/alexedwards/scs/v2"
	"github.com/wfabjanczuk/sawler_bookings/internal/driver"
	"github.com/wfabjanczuk/sawler_bookings/internal/handlers"
	"github.com/wfabjanczuk/sawler_bookings/internal/helpers"
	"github.com/wfabjanczuk/sawler_bookings/internal/models"
	"github.com/wfabjanczuk/sawler_bookings/internal/render"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

func Test(t *testing.T) {
	_, err := testInitialize()

	if err != nil {
		t.Error("failed initialize()")
	}
}

func testInitialize() (*driver.DB, error) {
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(map[string]int{})

	app.MailChannel = make(chan models.MailData)
	app.InProduction = false
	app.InfoLog = log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime)
	app.ErrorLog = log.New(os.Stdout, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	tc, err := render.CreateTemplateCache()

	if err != nil {
		return nil, err
	}

	app.TemplateCache = tc
	app.UseCache = false

	repo := handlers.NewTestRepo(&app)
	render.NewRenderer(&app)
	handlers.NewHandlers(repo)

	helpers.NewHelpers(&app)

	return nil, nil
}
