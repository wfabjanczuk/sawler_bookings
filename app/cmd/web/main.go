package main

import (
	"encoding/gob"
	"fmt"
	"github.com/alexedwards/scs/v2"
	"github.com/joho/godotenv"
	"github.com/wfabjanczuk/sawler_bookings/internal/config"
	"github.com/wfabjanczuk/sawler_bookings/internal/driver"
	"github.com/wfabjanczuk/sawler_bookings/internal/handlers"
	"github.com/wfabjanczuk/sawler_bookings/internal/helpers"
	"github.com/wfabjanczuk/sawler_bookings/internal/models"
	"github.com/wfabjanczuk/sawler_bookings/internal/render"
	"log"
	"net/http"
	"os"
	"time"
)

var app config.AppConfig
var session *scs.SessionManager

func main() {
	db, err := initialize()

	if err != nil {
		log.Fatal(err)
	}

	defer db.SQL.Close()
	defer close(app.MailChannel)

	fmt.Println("Starting mail listener")
	listenForMail()

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("Error loading PORT from .env file")
	}

	addr := fmt.Sprintf(":%s", port)
	fmt.Printf("Starting application on %s\n", addr)

	srv := &http.Server{
		Addr:    addr,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	log.Fatal(err)
}

func initialize() (*driver.DB, error) {
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

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dsn := os.Getenv("DSN")
	if dsn == "" {
		log.Fatal("Error loading DSN from .env file")
	}

	log.Println("Connecting to the database...")
	db, err := driver.ConnectSQL(dsn)

	if err != nil {
		log.Fatal("Cannot connect to database! Dying...")
	}

	log.Println("Connected to database!")

	tc, err := render.CreateTemplateCache()

	if err != nil {
		return nil, err
	}

	app.TemplateCache = tc
	app.UseCache = false

	repo := handlers.NewRepo(&app, db)
	render.NewRenderer(&app)
	handlers.NewHandlers(repo)

	helpers.NewHelpers(&app)

	return db, nil
}
