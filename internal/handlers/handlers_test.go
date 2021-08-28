package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
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
	{"/make-reservation", "GET", []postData{}, http.StatusOK},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	for _, handlerTest := range theTests {
		if handlerTest.method == "GET" {
			response, err := ts.Client().Get(ts.URL + handlerTest.url)

			if err != nil {
				t.Log(err)
				t.Fatal(err)
			}

			if response.StatusCode != handlerTest.expectedStatusCode {
				t.Errorf("for %s, expected %d but got %d", handlerTest.url, handlerTest.expectedStatusCode, response.StatusCode)
			}
		} else {

		}
	}
}
