package render

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/justinas/nosurf"
	"github.com/wfabjanczuk/sawler_bookings/internal/config"
	"github.com/wfabjanczuk/sawler_bookings/internal/models"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

var functions = template.FuncMap{}

var app *config.AppConfig
var pathToTemplates = "./templates/pages"

func NewRenderer(a *config.AppConfig) {
	app = a
}

func AddDefaultData(r *http.Request, td *models.TemplateData) *models.TemplateData {
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Warning = app.Session.PopString(r.Context(), "warning")
	td.Error = app.Session.PopString(r.Context(), "error")
	td.CsrfToken = nosurf.Token(r)

	return td
}

func Template(w http.ResponseWriter, r *http.Request, tmpl string, td *models.TemplateData) error {
	var tc map[string]*template.Template
	var err error

	if app.UseCache {
		tc = app.TemplateCache
	} else {
		tc, _ = CreateTemplateCache()
	}

	parsedTemplate, ok := tc[tmpl]

	if !ok {
		return errors.New("could not get template from cache")
	}

	buf := new(bytes.Buffer)
	td = AddDefaultData(r, td)

	err = parsedTemplate.Execute(buf, td)

	if err != nil {
		log.Fatal(err)
		return err
	}

	_, err = buf.WriteTo(w)

	if err != nil {
		fmt.Println("error writing to writer", err)
		return err
	}

	return nil
}

func CreateTemplateCache() (map[string]*template.Template, error) {
	myCache := map[string]*template.Template{}

	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.tmpl", pathToTemplates))

	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		ts, err := template.New(name).Funcs(functions).ParseFiles(page)

		if err != nil {
			return nil, err
		}

		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))

		if err != nil {
			return nil, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))

			if err != nil {
				return nil, err
			}
		}

		myCache[name] = ts
	}

	return myCache, nil
}
