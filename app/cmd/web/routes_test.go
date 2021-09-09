package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/wfabjanczuk/sawler_bookings/internal/config"
	"testing"
)

func TestRoutes(t *testing.T) {
	app := config.AppConfig{}

	mux := routes(&app)

	switch v := mux.(type) {
	case *chi.Mux:
	default:
		t.Error(fmt.Sprintf("type is not *chi.Mux, but is %T", v))
	}
}
