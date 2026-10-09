package main

import (
	"bytes"
	"html/template"
	"net/http"
	"net/mail"
	"strings"
	"time"

	sqlc "website.tabirianalytics.com/internal/database/sqlc"
)

const maxFormBodyBytes = 64 << 10

type ContactForm struct {
	FullName           string
	CorporateEmail     string
	WorkflowBottleneck string
	Errors             map[string]string
}

type PartnerContactForm struct {
	AgencyName         string
	CorporateEmail     string
	ClientVolume       string
	PrimaryVertical    string
	DeliveryConstraint string
	Errors             map[string]string
}

func (app *application) contactSubmit(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxFormBodyBytes)
	if err := r.ParseForm(); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := ContactForm{
		FullName:           strings.TrimSpace(r.PostForm.Get("full_name")),
		CorporateEmail:     strings.TrimSpace(r.PostForm.Get("corporate_email")),
		WorkflowBottleneck: strings.TrimSpace(r.PostForm.Get("workflow_bottleneck")),
		Errors:             make(map[string]string),
	}
	form.validate()

	if len(form.Errors) > 0 {
		app.renderContact(w, r, http.StatusUnprocessableEntity, form)
		return
	}

	_, err := app.queries.CreateContactSubmission(r.Context(), sqlc.CreateContactSubmissionParams{
		FullName:           form.FullName,
		CorporateEmail:     form.CorporateEmail,
		WorkflowBottleneck: form.WorkflowBottleneck,
	})
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, "/contact?submitted=true", http.StatusSeeOther)
}

func (app *application) partnerContactSubmit(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxFormBodyBytes)
	if err := r.ParseForm(); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := PartnerContactForm{
		AgencyName:         strings.TrimSpace(r.PostForm.Get("agency_name")),
		CorporateEmail:     strings.TrimSpace(r.PostForm.Get("corporate_email")),
		ClientVolume:       strings.TrimSpace(r.PostForm.Get("client_volume")),
		PrimaryVertical:    strings.TrimSpace(r.PostForm.Get("primary_vertical")),
		DeliveryConstraint: strings.TrimSpace(r.PostForm.Get("delivery_constraint")),
		Errors:             make(map[string]string),
	}
	form.validate()

	if len(form.Errors) > 0 {
		app.renderPartnerContact(w, r, http.StatusUnprocessableEntity, form)
		return
	}

	_, err := app.queries.CreatePartnerContactSubmission(r.Context(), sqlc.CreatePartnerContactSubmissionParams{
		AgencyName:         form.AgencyName,
		CorporateEmail:     form.CorporateEmail,
		ClientVolume:       form.ClientVolume,
		PrimaryVertical:    form.PrimaryVertical,
		DeliveryConstraint: form.DeliveryConstraint,
	})
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, "/partner-contact?submitted=true", http.StatusSeeOther)
}

func (form *ContactForm) validate() {
	validateRequiredLength(form.Errors, "full_name", form.FullName, "Enter your full name.", 200)
	validateEmail(form.Errors, form.CorporateEmail)
	validateRequiredLength(form.Errors, "workflow_bottleneck", form.WorkflowBottleneck, "Describe your workflow bottleneck.", 5000)
}

func (form *PartnerContactForm) validate() {
	validateRequiredLength(form.Errors, "agency_name", form.AgencyName, "Enter your agency name.", 200)
	validateEmail(form.Errors, form.CorporateEmail)
	validateRequiredLength(form.Errors, "primary_vertical", form.PrimaryVertical, "Enter your primary vertical.", 200)
	validateRequiredLength(form.Errors, "delivery_constraint", form.DeliveryConstraint, "Describe your current delivery constraint.", 5000)

	switch form.ClientVolume {
	case "1-10", "11-50", "50+":
	default:
		form.Errors["client_volume"] = "Select a valid client volume."
	}
}

func validateRequiredLength(errors map[string]string, field, value, requiredMessage string, maxLength int) {
	if value == "" {
		errors[field] = requiredMessage
	} else if len(value) > maxLength {
		errors[field] = "This field is too long."
	}
}

func validateEmail(errors map[string]string, email string) {
	if email == "" {
		errors["corporate_email"] = "Enter your corporate email."
		return
	}
	if len(email) > 320 {
		errors["corporate_email"] = "Enter a valid email address."
		return
	}

	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		errors["corporate_email"] = "Enter a valid email address."
	}
}

func (app *application) renderContact(w http.ResponseWriter, r *http.Request, status int, form ContactForm) {
	data := PageData{
		PageTitle:            "Commercial Contact | Tabiri Analytics",
		PageDescription:      "Contact Tabiri's commercial team to plan secure AI deployment, workflow modernization, and data-protected operations for your organization.",
		OpenGraphDescription: "Connect with Tabiri's commercial team to evaluate secure AI solutions for business operations.",
		CurrentPage:          "contact",
		CurrentYear:          time.Now().Year(),
		FormSubmitted:        r.URL.Query().Get("submitted") == "true",
		ContactForm:          form,
	}
	app.renderContactTemplate(w, r, status, data, "./ui/html/pages/contact.tmpl")
}

func (app *application) renderPartnerContact(w http.ResponseWriter, r *http.Request, status int, form PartnerContactForm) {
	data := PageData{
		PageTitle:            "Partner Contact | Channel Partner Application | Tabiri",
		PageDescription:      "Apply to become a Tabiri Analytics channel partner and co-deliver secure AI solutions for enterprise and public-sector clients.",
		OpenGraphDescription: "Submit your channel partner application and tell us how your team delivers secure AI outcomes at scale.",
		CurrentPage:          "partner-contact",
		CurrentYear:          time.Now().Year(),
		FormSubmitted:        r.URL.Query().Get("submitted") == "true",
		PartnerContactForm:   form,
	}
	app.renderContactTemplate(w, r, status, data, "./ui/html/pages/partner-contact.tmpl")
}

func (app *application) renderContactTemplate(w http.ResponseWriter, r *http.Request, status int, data PageData, page string) {
	w.Header().Add("Server", "Go")

	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		page,
		"./ui/html/partials/footer.tmpl",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	var buf bytes.Buffer
	if err := ts.ExecuteTemplate(&buf, "base", data); err != nil {
		app.serverError(w, r, err)
		return
	}

	w.WriteHeader(status)
	if _, err := buf.WriteTo(w); err != nil {
		app.logger.Error(err.Error(), "method", r.Method, "uri", r.URL.RequestURI())
	}
}
