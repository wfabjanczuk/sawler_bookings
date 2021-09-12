package helpers

import (
	"github.com/wfabjanczuk/sawler_bookings/internal/config"
)

var app *config.AppConfig

func NewHelpers(a *config.AppConfig) {
	app = a
}
