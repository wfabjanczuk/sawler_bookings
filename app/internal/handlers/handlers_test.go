package handlers

import (
	"context"
	"encoding/json"
	"github.com/wfabjanczuk/sawler_bookings/internal/models"
	"net/http"
	"net/http/httptest"
	"strings"
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
	{"/user/login", "GET", http.StatusOK},
	{"/user/logout", "GET", http.StatusOK},
	{"/admin/reservations-all", "GET", http.StatusOK},
	{"/admin/reservations-new", "GET", http.StatusOK},
	{"/admin/reservations-calendar", "GET", http.StatusOK},
	{"/admin/reservations/all/1", "GET", http.StatusOK},
	{"/non-existent-route", "GET", http.StatusNotFound},
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

	if responseRecorder.Code != http.StatusSeeOther {
		t.Errorf("Reservation handler returned wrong response code: got %d, expected %d", responseRecorder.Code, http.StatusSeeOther)
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

	if responseRecorder.Code != http.StatusSeeOther {
		t.Errorf("Reservation handler returned wrong response code: got %d, expected %d", responseRecorder.Code, http.StatusSeeOther)
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

	if responseRecorder.Code != http.StatusSeeOther {
		t.Errorf("Reservation handler returned wrong response code: got %d, expected %d", responseRecorder.Code, http.StatusSeeOther)
	}
}

func TestRepository_PostReservation(t *testing.T) {
	requestBody := strings.Join([]string{
		"start_date=2050-01-01",
		"end_date=2050-01-02",
		"first_name=John",
		"last_name=Smith",
		"email=john@smith.com",
		"phone=900900900",
		"room_id=1",
	}, "&")

	// Correct workflow
	request, err := http.NewRequest("POST", "/make-reservation", strings.NewReader(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	ctx := getContext(request, t)
	request = request.WithContext(ctx)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	responseRecorder := httptest.NewRecorder()
	handler := http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusSeeOther {
		t.Errorf("PostReservation handler returned wrong response code: got %d, expected %d", responseRecorder.Code, http.StatusSeeOther)
	}

	// Missing post body
	request, err = http.NewRequest("POST", "/make-reservation", nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx = getContext(request, t)
	request = request.WithContext(ctx)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	responseRecorder = httptest.NewRecorder()
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusSeeOther {
		t.Errorf("PostReservation handler returned wrong response code: got %d, expected %d", responseRecorder.Code, http.StatusSeeOther)
	}

	// Invalid start date
	requestBody = strings.Join([]string{
		"start_date=invalid",
		"end_date=2050-01-02",
		"first_name=John",
		"last_name=Smith",
		"email=john@smith.com",
		"phone=900900900",
		"room_id=1",
	}, "&")

	request, err = http.NewRequest("POST", "/make-reservation", strings.NewReader(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	ctx = getContext(request, t)
	request = request.WithContext(ctx)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	responseRecorder = httptest.NewRecorder()
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusSeeOther {
		t.Errorf("PostReservation handler returned wrong response code: got %d, expected %d", responseRecorder.Code, http.StatusSeeOther)
	}

	// Invalid end date
	requestBody = strings.Join([]string{
		"start_date=2050-01-01",
		"end_date=invalid",
		"first_name=John",
		"last_name=Smith",
		"email=john@smith.com",
		"phone=900900900",
		"room_id=1",
	}, "&")

	request, err = http.NewRequest("POST", "/make-reservation", strings.NewReader(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	ctx = getContext(request, t)
	request = request.WithContext(ctx)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	responseRecorder = httptest.NewRecorder()
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusSeeOther {
		t.Errorf("PostReservation handler returned wrong response code: got %d, expected %d", responseRecorder.Code, http.StatusSeeOther)
	}

	// Invalid room id
	requestBody = strings.Join([]string{
		"start_date=2050-01-01",
		"end_date=2050-01-02",
		"first_name=John",
		"last_name=Smith",
		"email=john@smith.com",
		"phone=900900900",
		"room_id=invalid",
	}, "&")

	request, err = http.NewRequest("POST", "/make-reservation", strings.NewReader(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	ctx = getContext(request, t)
	request = request.WithContext(ctx)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	responseRecorder = httptest.NewRecorder()
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusSeeOther {
		t.Errorf("PostReservation handler returned wrong response code: got %d, expected %d", responseRecorder.Code, http.StatusSeeOther)
	}

	// Invalid data (first name shorter than 3 characters)
	requestBody = strings.Join([]string{
		"start_date=2050-01-01",
		"end_date=2050-01-02",
		"first_name=J",
		"last_name=Smith",
		"email=john@smith.com",
		"phone=900900900",
		"room_id=1",
	}, "&")

	request, err = http.NewRequest("POST", "/make-reservation", strings.NewReader(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	ctx = getContext(request, t)
	request = request.WithContext(ctx)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	responseRecorder = httptest.NewRecorder()
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Errorf("PostReservation handler returned wrong response code: got %d, expected %d", responseRecorder.Code, http.StatusOK)
	}

	// Can't insert reservation to database
	requestBody = strings.Join([]string{
		"start_date=2050-01-01",
		"end_date=2050-01-02",
		"first_name=John",
		"last_name=Smith",
		"email=john@smith.com",
		"phone=900900900",
		"room_id=1000",
	}, "&")

	request, err = http.NewRequest("POST", "/make-reservation", strings.NewReader(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	ctx = getContext(request, t)
	request = request.WithContext(ctx)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	responseRecorder = httptest.NewRecorder()
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusSeeOther {
		t.Errorf("PostReservation handler returned wrong response code: got %d, expected %d", responseRecorder.Code, http.StatusSeeOther)
	}

	// Can't insert reservation to database
	requestBody = strings.Join([]string{
		"start_date=2050-01-01",
		"end_date=2050-01-02",
		"first_name=John",
		"last_name=Smith",
		"email=john@smith.com",
		"phone=900900900",
		"room_id=2",
	}, "&")

	request, err = http.NewRequest("POST", "/make-reservation", strings.NewReader(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	ctx = getContext(request, t)
	request = request.WithContext(ctx)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	responseRecorder = httptest.NewRecorder()
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusSeeOther {
		t.Errorf("PostReservation handler returned wrong response code: got %d, expected %d", responseRecorder.Code, http.StatusSeeOther)
	}

	// Can't insert room restriction to database
	requestBody = strings.Join([]string{
		"start_date=2050-01-01",
		"end_date=2050-01-02",
		"first_name=John",
		"last_name=Smith",
		"email=john@smith.com",
		"phone=900900900",
		"room_id=1000",
	}, "&")

	request, err = http.NewRequest("POST", "/make-reservation", strings.NewReader(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	ctx = getContext(request, t)
	request = request.WithContext(ctx)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	responseRecorder = httptest.NewRecorder()
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusSeeOther {
		t.Errorf("PostReservation handler returned wrong response code: got %d, expected %d", responseRecorder.Code, http.StatusSeeOther)
	}
}

func TestRepository_AvailabilityJson(t *testing.T) {
	requestBody := strings.Join([]string{
		"start_date=2050-01-01",
		"end_date=2050-01-02",
		"room_id=1",
	}, "&")

	request, err := http.NewRequest("POST", "/search-availability-json", strings.NewReader(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	ctx := getContext(request, t)
	request = request.WithContext(ctx)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	responseRecorder := httptest.NewRecorder()
	handler := http.HandlerFunc(Repo.AvailabilityJson)
	handler.ServeHTTP(responseRecorder, request)

	var jsonResponse availabilityJsonResponse
	err = json.Unmarshal([]byte(responseRecorder.Body.String()), &jsonResponse)
	if err != nil {
		t.Error("failed to parse json")
	}

	if !jsonResponse.Ok {
		t.Error("Room supposed to be available")
	}

	requestBody = strings.Join([]string{
		"start_date=2050-01-01",
		"end_date=2050-01-02",
		"room_id=1000",
	}, "&")

	request, err = http.NewRequest("POST", "/search-availability-json", strings.NewReader(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	ctx = getContext(request, t)
	request = request.WithContext(ctx)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	responseRecorder = httptest.NewRecorder()
	handler.ServeHTTP(responseRecorder, request)

	err = json.Unmarshal([]byte(responseRecorder.Body.String()), &jsonResponse)
	if err != nil {
		t.Error("failed to parse json")
	}

	if jsonResponse.Ok {
		t.Error("Room supposed not to be available")
	}
}

func getContext(request *http.Request, t *testing.T) context.Context {
	ctx, err := session.Load(request.Context(), request.Header.Get("X-Session"))
	if err != nil {
		t.Fatal(err)
	}

	return ctx
}
