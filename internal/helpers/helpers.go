package helpers

import (
	"fmt"
	"github.com/wfabjanczuk/sawler_bookings/internal/config"
	"net/http"
	"runtime/debug"
)

var app *config.AppConfig

func NewHelpers(a *config.AppConfig) {
	app = a
}

func ClientError(w http.ResponseWriter, statusCode int) {
	app.InfoLog.Println("Client error with status of", statusCode)
	http.Error(w, http.StatusText(statusCode), statusCode)
}

func ServerError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.ErrorLog.Println(trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
