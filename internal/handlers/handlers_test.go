package handlers

import (
	"context"
	"github.com/wfabjanczuk/sawler_bookings/internal/models"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type postData struct {
	key   string
	value string
}

var theTests = []struct {
	url                string
	method             string
	expectedStatusCode int
}{
	{"/", "GET", http.StatusOK},
	{"/about", "GET", http.StatusOK},
	{"/generals-quarters", "GET", http.StatusOK},
	{"/majors-suite", "GET", http.StatusOK},
	{"/search-availability", "GET", http.StatusOK},
	{"/contact", "GET", http.StatusOK},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	var response *http.Response
	var err error

	for _, handlerTest := range theTests {
		response, err = ts.Client().Get(ts.URL + handlerTest.url)

		if err != nil {
			t.Fatal(err)
		}

		if response.StatusCode != handlerTest.expectedStatusCode {
			t.Errorf("for %s %s, expected %d but got %d", handlerTest.method, handlerTest.url, handlerTest.expectedStatusCode, response.StatusCode)
		}
	}
}

func TestRepository_Reservation(t *testing.T) {
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
		StartDate: time.Now().AddDate(0, 0, 10),
		EndDate:   time.Now().AddDate(0, 0, 20),
	}

	// Correct workflow
	request, err := http.NewRequest("GET", "/make-reservation", nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx := getContext(request, t)
	request = request.WithContext(ctx)
	session.Put(ctx, "reservation", reservation)

	responseRecorder := httptest.NewRecorder()
	handler := http.HandlerFunc(Repo.Reservation)
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Errorf("Reservation handler returned wrong response code: got %d, expected %d", responseRecorder.Code, http.StatusOK)
	}

	// Reservation not found in session
	request, err = http.NewRequest("GET", "/make-reservation", nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx = getContext(request, t)
	request = request.WithContext(ctx)

	responseRecorder = httptest.NewRecorder()
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned wrong response code: got %d, expected %d", responseRecorder.Code, http.StatusTemporaryRedirect)
	}

	// Room not found in database
	request, err = http.NewRequest("GET", "/make-reservation", nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx = getContext(request, t)
	request = request.WithContext(ctx)

	reservation.RoomID = 1000
	session.Put(ctx, "reservation", reservation)

	responseRecorder = httptest.NewRecorder()
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned wrong response code: got %d, expected %d", responseRecorder.Code, http.StatusTemporaryRedirect)
	}

	// Wrong dates
	request, err = http.NewRequest("GET", "/make-reservation", nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx = getContext(request, t)
	request = request.WithContext(ctx)

	reservation.RoomID = 1
	reservation.StartDate = time.Now().AddDate(0, 0, 20)
	reservation.EndDate = time.Now().AddDate(0, 0, 10)
	session.Put(ctx, "reservation", reservation)

	responseRecorder = httptest.NewRecorder()
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned wrong response code: got %d, expected %d", responseRecorder.Code, http.StatusTemporaryRedirect)
	}
}

func getContext(request *http.Request, t *testing.T) context.Context {
	ctx, err := session.Load(request.Context(), request.Header.Get("X-Session"))
	if err != nil {
		t.Fatal(err)
	}

	return ctx
}
