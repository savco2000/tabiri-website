package main

import (
	"html/template"
	"net/http"
	"time"
)

type PageData struct {
	PageTitle   string
	CurrentPage string
	CurrentYear int
}

// Change the signature of the home handler so it is defined as a method against
// *application.
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Server", "Go")

	data := PageData{
		PageTitle:   "Home - Secure AI-Powered Employees | Tabiri Analytics",
		CurrentPage: "home",
		CurrentYear: time.Now().Year(),
	}

	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/pages/home.tmpl",
		"./ui/html/partials/footer.tmpl",
	}

	ts, err := template.ParseFiles(files...)

	if err != nil {
		app.serverError(w, r, err) // Use the serverError() helper.
		return
	}
	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		app.serverError(w, r, err) // Use the serverError() helper.
	}
}
