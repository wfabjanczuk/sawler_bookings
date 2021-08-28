package main

import (
	"encoding/gob"
	"fmt"
	"github.com/alexedwards/scs/v2"
	"github.com/wfabjanczuk/sawler_bookings/internal/config"
	"github.com/wfabjanczuk/sawler_bookings/internal/handlers"
	"github.com/wfabjanczuk/sawler_bookings/internal/models"
	"github.com/wfabjanczuk/sawler_bookings/internal/render"
	"log"
	"net/http"
	"time"
)

const portNumber = 8080

var app config.AppConfig
var session *scs.SessionManager

func main() {
	err := initialize()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Starting application on port %d\n", portNumber)

	addr := fmt.Sprintf(":%d", portNumber)
	srv := &http.Server{
		Addr:    addr,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	log.Fatal(err)
}

func initialize() error {
	gob.Register(models.Reservation{})

	app.InProduction = false

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	tc, err := render.CreateTemplateCache()

	if err != nil {
		return err
	}

	app.TemplateCache = tc
	app.UseCache = false

	repo := handlers.NewRepo(&app)
	render.NewTemplates(&app)
	handlers.NewHandlers(repo)

	return nil
}
