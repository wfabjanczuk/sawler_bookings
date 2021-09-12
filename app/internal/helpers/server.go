package helpers

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

func ServerError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.ErrorLog.Println(trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func IsAuthenticated(r *http.Request) bool {
	return app.Session.Exists(r.Context(), "user_id")
}

func IsAuthorized(r *http.Request) bool {
	accessLevel := app.Session.GetInt(r.Context(), "access_level")

	return r.Method != "POST" || accessLevel > 0
}
