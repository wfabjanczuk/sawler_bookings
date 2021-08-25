package render

import (
	"bytes"
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

func NewTemplates(a *config.AppConfig) {
	app = a
}

func AddDefaultData(r *http.Request, td *models.TemplateData) *models.TemplateData {
	td.CsrfToken = nosurf.Token(r)

	return td
}

func RenderTemplate(w http.ResponseWriter, r *http.Request, tmpl string, td *models.TemplateData) {
	var tc map[string]*template.Template
	var err error

	if app.UseCache {
		tc = app.TemplateCache
	} else {
		tc, _ = CreateTemplateCache()
	}

	parsedTemplate, ok := tc[tmpl]

	if !ok {
		log.Fatal("Could not get template", tmpl)
	}

	buf := new(bytes.Buffer)
	td = AddDefaultData(r, td)

	_ = parsedTemplate.Execute(buf, td)
	_, err = buf.WriteTo(w)

	if err != nil {
		fmt.Println("error writing to writer", err)
	}
}

func CreateTemplateCache() (map[string]*template.Template, error) {
	myCache := map[string]*template.Template{}

	pages, err := filepath.Glob("./templates/*.page.html")

	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		ts, err := template.New(name).Funcs(functions).ParseFiles(page)

		if err != nil {
			return nil, err
		}

		matches, err := filepath.Glob("./templates/*.layout.html")

		if err != nil {
			return nil, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob("./templates/*.layout.html")

			if err != nil {
				return nil, err
			}
		}

		myCache[name] = ts
	}

	return myCache, nil
}
