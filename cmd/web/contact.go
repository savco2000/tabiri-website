package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net/http"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
	"time"

	sqlc "website.tabirianalytics.com/internal/database/sqlc"
)

const maxFormBodyBytes = 64 << 10

type ContactForm struct {
	FullName           string
	CorporateEmail     string
	WorkflowBottleneck string
	WorkflowName       string
	AssessmentResult   string
	AssessmentScore    string
	AssessmentAnswers  string
	UTMSource          string
	UTMMedium          string
	UTMCampaign        string
	UTMContent         string
	Referrer           string
	Errors             map[string]string
}

type assessmentAnswer struct {
	Score int    `json:"score"`
	Label string `json:"label"`
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
		WorkflowName:       strings.TrimSpace(r.PostForm.Get("workflow_name")),
		AssessmentResult:   strings.TrimSpace(r.PostForm.Get("assessment_result")),
		AssessmentScore:    strings.TrimSpace(r.PostForm.Get("assessment_score")),
		AssessmentAnswers:  strings.TrimSpace(r.PostForm.Get("assessment_answers")),
		UTMSource:          strings.TrimSpace(r.PostForm.Get("utm_source")),
		UTMMedium:          strings.TrimSpace(r.PostForm.Get("utm_medium")),
		UTMCampaign:        strings.TrimSpace(r.PostForm.Get("utm_campaign")),
		UTMContent:         strings.TrimSpace(r.PostForm.Get("utm_content")),
		Referrer:           strings.TrimSpace(r.PostForm.Get("referrer")),
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
		WorkflowName:       form.WorkflowName,
		AssessmentResult:   form.AssessmentResult,
		AssessmentScore:    form.AssessmentScore,
		AssessmentAnswers:  form.AssessmentAnswers,
		UtmSource:          form.UTMSource,
		UtmMedium:          form.UTMMedium,
		UtmCampaign:        form.UTMCampaign,
		UtmContent:         form.UTMContent,
		Referrer:           form.Referrer,
	})
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	redirect := "/contact?submitted=true"
	if form.AssessmentResult != "" {
		redirect += "&lead_source=automation-assessment"
	}
	http.Redirect(w, r, redirect, http.StatusSeeOther)
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
	validateOptionalLength(form.Errors, "workflow_name", form.WorkflowName, 200)
	validateOptionalLength(form.Errors, "assessment_answers", form.AssessmentAnswers, 3000)
	validateOptionalLength(form.Errors, "utm_source", form.UTMSource, 200)
	validateOptionalLength(form.Errors, "utm_medium", form.UTMMedium, 200)
	validateOptionalLength(form.Errors, "utm_campaign", form.UTMCampaign, 200)
	validateOptionalLength(form.Errors, "utm_content", form.UTMContent, 200)
	validateOptionalLength(form.Errors, "referrer", form.Referrer, 2048)

	if form.AssessmentResult != "" {
		switch form.AssessmentResult {
		case "foundation", "emerging", "strong":
		default:
			form.Errors["assessment_result"] = "Invalid assessment result."
		}

		score, err := strconv.Atoi(form.AssessmentScore)
		if err != nil || score < 0 || score > 10 {
			form.Errors["assessment_score"] = "Invalid assessment score."
		}
		answerScore, oversight, answersValid := parseAssessmentAnswers(form.AssessmentAnswers)
		if !answersValid {
			form.Errors["assessment_answers"] = "Invalid assessment answers."
		} else if err == nil && (answerScore != score || assessmentResult(answerScore, oversight) != form.AssessmentResult) {
			form.Errors["assessment_answers"] = "Assessment answers do not match the result."
		}
	} else if form.AssessmentScore != "" || form.AssessmentAnswers != "" {
		form.Errors["assessment_result"] = "Assessment result is required."
	}

	if form.Referrer != "" {
		referrer, err := url.Parse(form.Referrer)
		if err != nil || referrer.Host == "" || (referrer.Scheme != "http" && referrer.Scheme != "https") {
			form.Errors["referrer"] = "Invalid referring URL."
		}
	}
}

func parseAssessmentAnswers(value string) (int, int, bool) {
	if value == "" {
		return 0, 0, false
	}
	var answers map[string]assessmentAnswer
	if err := json.Unmarshal([]byte(value), &answers); err != nil {
		return 0, 0, false
	}
	score := 0
	for _, field := range []string{"frequency", "consistency", "systems", "impact", "oversight"} {
		answer, ok := answers[field]
		if !ok || answer.Score < 0 || answer.Score > 2 || answer.Label == "" || len(answer.Label) > 200 {
			return 0, 0, false
		}
		score += answer.Score
	}
	if len(answers) != 5 {
		return 0, 0, false
	}
	return score, answers["oversight"].Score, true
}

func assessmentResult(score, oversight int) string {
	if score >= 8 && oversight > 0 {
		return "strong"
	}
	if score >= 4 {
		return "emerging"
	}
	return "foundation"
}

func (form ContactForm) AssessmentLabel() string {
	switch form.AssessmentResult {
	case "strong":
		return "Strong automation candidate"
	case "emerging":
		return "Emerging automation candidate"
	case "foundation":
		return "Foundation-building recommended"
	default:
		return "Automation readiness assessment"
	}
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

func validateOptionalLength(errors map[string]string, field, value string, maxLength int) {
	if len(value) > maxLength {
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
		CanonicalURL:         "https://tabirianalytics.com/contact",
		CurrentPage:          "contact",
		CurrentYear:          time.Now().Year(),
		FormSubmitted:        r.URL.Query().Get("submitted") == "true",
		LeadSource:           contactLeadSource(r.URL.Query().Get("lead_source"), form),
		ContactForm:          form,
	}
	app.renderContactTemplate(w, r, status, data, "./ui/html/pages/contact.tmpl")
}

func contactLeadSource(querySource string, form ContactForm) string {
	if querySource == "automation-assessment" || form.AssessmentResult != "" {
		return "automation-assessment"
	}
	return "direct"
}

func contactFormFromQuery(query url.Values) ContactForm {
	workflowName := strings.TrimSpace(query.Get("workflow_name"))
	workflowBottleneck := ""
	if workflowName != "" {
		workflowBottleneck = workflowName
	}
	return ContactForm{
		WorkflowBottleneck: workflowBottleneck,
		WorkflowName:       workflowName,
		AssessmentResult:   strings.TrimSpace(query.Get("assessment_result")),
		AssessmentScore:    strings.TrimSpace(query.Get("assessment_score")),
		AssessmentAnswers:  strings.TrimSpace(query.Get("assessment_answers")),
		UTMSource:          strings.TrimSpace(query.Get("utm_source")),
		UTMMedium:          strings.TrimSpace(query.Get("utm_medium")),
		UTMCampaign:        strings.TrimSpace(query.Get("utm_campaign")),
		UTMContent:         strings.TrimSpace(query.Get("utm_content")),
		Referrer:           strings.TrimSpace(query.Get("referrer")),
		Errors:             make(map[string]string),
	}
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
