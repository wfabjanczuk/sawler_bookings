package handlers

import (
	"context"
	"github.com/wfabjanczuk/sawler_bookings/internal/models"
	"net/http"
	"net/http/httptest"
	"net/url"
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
	params             []postData
	expectedStatusCode int
}{
	{"/", "GET", []postData{}, http.StatusOK},
	{"/about", "GET", []postData{}, http.StatusOK},
	{"/generals-quarters", "GET", []postData{}, http.StatusOK},
	{"/majors-suite", "GET", []postData{}, http.StatusOK},
	{"/search-availability", "GET", []postData{}, http.StatusOK},
	{"/contact", "GET", []postData{}, http.StatusOK},
	//{"/make-reservation", "GET", []postData{}, http.StatusOK},

	//{"/search-availability", "POST", []postData{
	//	{key: "start_date", value: "2020-01-01"},
	//	{key: "end_date", value: "2020-01-02"},
	//}, http.StatusOK},
	//{"/search-availability-json", "POST", []postData{
	//	{key: "start_date", value: "2020-01-01"},
	//	{key: "end_date", value: "2020-01-02"},
	//}, http.StatusOK},
	//{"/make-reservation", "POST", []postData{
	//	{key: "first_name", value: "John"},
	//	{key: "last_name", value: "Doe"},
	//	{key: "email", value: "john@doe.com"},
	//	{key: "phone", value: "555 555 555"},
	//}, http.StatusOK},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	var response *http.Response
	var err error

	for _, handlerTest := range theTests {
		if handlerTest.method == "GET" {
			response, err = ts.Client().Get(ts.URL + handlerTest.url)
		} else {
			values := url.Values{}

			for _, param := range handlerTest.params {
				values.Add(param.key, param.value)
			}

			response, err = ts.Client().PostForm(ts.URL+handlerTest.url, values)
		}

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

	request, err := http.NewRequest("GET", "/make-reservation", nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx := getContext(request, t)
	request = request.WithContext(ctx)

	responseRecorder := httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.Reservation)
	handler.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Errorf("Reservation handler returned wrong response code: got %d, expected %d", responseRecorder.Code, http.StatusOK)
	}
}

func getContext(request *http.Request, t *testing.T) context.Context {
	ctx, err := session.Load(request.Context(), request.Header.Get("X-Session"))
	if err != nil {
		t.Fatal(err)
	}

	return ctx
}
