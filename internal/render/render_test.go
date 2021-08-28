package render

import (
	"github.com/wfabjanczuk/sawler_bookings/internal/models"
	"net/http"
	"testing"
)

func TestAddDefaultData(t *testing.T) {
	var td models.TemplateData
	r, err := getRequestWithSession()

	if err != nil {
		t.Error(err)
	}

	session.Put(r.Context(), "flash", "123")
	result := AddDefaultData(r, &td)

	if result.Flash != "123" {
		t.Error("Flash value of 123 not found in session")
	}
}

func getRequestWithSession() (*http.Request, error) {
	r, err := http.NewRequest("GET", "/", nil)

	if err != nil {
		return nil, err
	}

	context := r.Context()
	context, _ = session.Load(context, r.Header.Get("X-Session"))

	return r.WithContext(context), nil
}
