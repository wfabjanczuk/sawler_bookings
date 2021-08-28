package render

import (
	"encoding/gob"
	"github.com/alexedwards/scs/v2"
	"github.com/wfabjanczuk/sawler_bookings/internal/config"
	"github.com/wfabjanczuk/sawler_bookings/internal/models"
	"net/http"
	"os"
	"testing"
	"time"
)

var session *scs.SessionManager
var testApp config.AppConfig

func TestMain(m *testing.M) {
	gob.Register(models.Reservation{})

	testApp.InProduction = false

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = testApp.InProduction

	testApp.Session = session

	app = &testApp

	os.Exit(m.Run())
}
