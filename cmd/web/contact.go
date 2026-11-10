package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/mail"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	sqlc "website.tabirianalytics.com/internal/database/sqlc"
)

const maxFormBodyBytes = 64 << 10
const conversionCookieName = "tabiri-conversion"
const conversionCookieMaxAge = 10 * time.Minute

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

func contactFormFromPost(values url.Values) ContactForm {
	return ContactForm{
		FullName:           strings.TrimSpace(values.Get("full_name")),
		CorporateEmail:     strings.TrimSpace(values.Get("corporate_email")),
		WorkflowBottleneck: strings.TrimSpace(values.Get("workflow_bottleneck")),
		WorkflowName:       strings.TrimSpace(values.Get("workflow_name")),
		AssessmentResult:   strings.TrimSpace(values.Get("assessment_result")),
		AssessmentScore:    strings.TrimSpace(values.Get("assessment_score")),
		AssessmentAnswers:  strings.TrimSpace(values.Get("assessment_answers")),
		UTMSource:          strings.TrimSpace(values.Get("utm_source")),
		UTMMedium:          strings.TrimSpace(values.Get("utm_medium")),
		UTMCampaign:        strings.TrimSpace(values.Get("utm_campaign")),
		UTMContent:         strings.TrimSpace(values.Get("utm_content")),
		Referrer:           strings.TrimSpace(values.Get("referrer")),
		Errors:             make(map[string]string),
	}
}

func partnerContactFormFromPost(values url.Values) PartnerContactForm {
	return PartnerContactForm{
		AgencyName:         strings.TrimSpace(values.Get("agency_name")),
		CorporateEmail:     strings.TrimSpace(values.Get("corporate_email")),
		ClientVolume:       strings.TrimSpace(values.Get("client_volume")),
		PrimaryVertical:    strings.TrimSpace(values.Get("primary_vertical")),
		DeliveryConstraint: strings.TrimSpace(values.Get("delivery_constraint")),
		Errors:             make(map[string]string),
	}
}

type assessmentAnswer struct {
	Score int    `json:"score"`
	Label string `json:"label"`
}

func (app *application) renderContactProtectionError(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	message string,
) {
	form := contactFormFromPost(r.PostForm)
	form.Errors["verification"] = message
	app.renderContact(w, r, status, form)
}

func (app *application) renderPartnerContactProtectionError(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	message string,
) {
	form := partnerContactFormFromPost(r.PostForm)
	form.Errors["verification"] = message
	app.renderPartnerContact(w, r, status, form)
}

type PartnerContactForm struct {
	AgencyName         string
	CorporateEmail     string
	ClientVolume       string
	PrimaryVertical    string
	DeliveryConstraint string
	Errors             map[string]string
}

type NewsletterForm struct {
	FirstName string
	Email     string
	Consent   bool
	Errors    map[string]string
}

func newsletterFormFromPost(values url.Values) NewsletterForm {
	return NewsletterForm{
		FirstName: strings.TrimSpace(values.Get("first_name")),
		Email:     strings.TrimSpace(values.Get("email")),
		Consent:   values.Get("consent") == "on",
		Errors:    make(map[string]string),
	}
}

func (app *application) renderNewsletterProtectionError(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	message string,
) {
	form := newsletterFormFromPost(r.PostForm)
	form.Errors["verification"] = message
	app.renderNewsletter(w, r, status, form)
}

func (app *application) contactSubmit(w http.ResponseWriter, r *http.Request) {
	if _, err := requireLeadVerification(r.Context()); err != nil {
		app.serverError(w, r, err)
		return
	}

	form := contactFormFromPost(r.PostForm)
	form.validate()

	if len(form.Errors) > 0 {
		app.renderContact(w, r, http.StatusUnprocessableEntity, form)
		return
	}

	err := app.databaseOperation(r.Context(), "create_lead_submission", func(ctx context.Context) error {
		_, err := app.queries.CreateLeadSubmission(ctx, sqlc.CreateLeadSubmissionParams{
			SubmissionType:     "commercial_lead",
			Email:              form.CorporateEmail,
			FirstName:          "",
			FullName:           form.FullName,
			AgencyName:         "",
			WorkflowBottleneck: form.WorkflowBottleneck,
			WorkflowName:       form.WorkflowName,
			AssessmentResult:   form.AssessmentResult,
			AssessmentScore:    form.AssessmentScore,
			AssessmentAnswers:  form.AssessmentAnswers,
			ClientVolume:       "",
			PrimaryVertical:    "",
			DeliveryConstraint: "",
			ConsentGiven:       0,
			UtmSource:          form.UTMSource,
			UtmMedium:          form.UTMMedium,
			UtmCampaign:        form.UTMCampaign,
			UtmContent:         form.UTMContent,
			Referrer:           form.Referrer,
			SpamStatus:         string(spamStatusAccepted),
			SpamReasons:        "[]",
			LeadStatus:         string(leadStatusNew),
		})
		return err
	})
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	app.telemetry.recordLeadOutcome(r.Context(), "contact", "accepted")
	app.logger.InfoContext(
		r.Context(),
		"lead submission accepted",
		"event", "lead.submission.accepted",
		"form", "contact",
		"request_id", requestIDFromContext(r.Context()),
	)
	app.setConversionCookie(w, r, "contact")

	redirect := "/contact?submitted=true"
	if form.AssessmentResult != "" {
		redirect += "&lead_source=automation-assessment"
	}
	http.Redirect(w, r, redirect, http.StatusSeeOther)
}

func (app *application) partnerContactSubmit(w http.ResponseWriter, r *http.Request) {
	if _, err := requireLeadVerification(r.Context()); err != nil {
		app.serverError(w, r, err)
		return
	}

	form := partnerContactFormFromPost(r.PostForm)
	form.validate()

	if len(form.Errors) > 0 {
		app.renderPartnerContact(w, r, http.StatusUnprocessableEntity, form)
		return
	}

	err := app.databaseOperation(r.Context(), "create_lead_submission", func(ctx context.Context) error {
		_, err := app.queries.CreateLeadSubmission(ctx, sqlc.CreateLeadSubmissionParams{
			SubmissionType:     "partner",
			Email:              form.CorporateEmail,
			FirstName:          "",
			FullName:           "",
			AgencyName:         form.AgencyName,
			WorkflowBottleneck: "",
			WorkflowName:       "",
			AssessmentResult:   "",
			AssessmentScore:    "",
			AssessmentAnswers:  "",
			ClientVolume:       form.ClientVolume,
			PrimaryVertical:    form.PrimaryVertical,
			DeliveryConstraint: form.DeliveryConstraint,
			ConsentGiven:       0,
			UtmSource:          "",
			UtmMedium:          "",
			UtmCampaign:        "",
			UtmContent:         "",
			Referrer:           "",
			SpamStatus:         string(spamStatusAccepted),
			SpamReasons:        "[]",
			LeadStatus:         string(leadStatusNew),
		})
		return err
	})
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	app.telemetry.recordLeadOutcome(r.Context(), "partner_contact", "accepted")
	app.logger.InfoContext(
		r.Context(),
		"lead submission accepted",
		"event", "lead.submission.accepted",
		"form", "partner_contact",
		"request_id", requestIDFromContext(r.Context()),
	)
	app.setConversionCookie(w, r, "partner_contact")

	http.Redirect(w, r, "/partner-contact?submitted=true", http.StatusSeeOther)
}

func (app *application) newsletterSubmit(w http.ResponseWriter, r *http.Request) {
	if _, err := requireLeadVerification(r.Context()); err != nil {
		app.serverError(w, r, err)
		return
	}

	form := newsletterFormFromPost(r.PostForm)
	form.validate()

	if len(form.Errors) > 0 {
		app.renderNewsletter(w, r, http.StatusUnprocessableEntity, form)
		return
	}

	err := app.databaseOperation(r.Context(), "create_lead_submission", func(ctx context.Context) error {
		_, err := app.queries.CreateLeadSubmission(ctx, sqlc.CreateLeadSubmissionParams{
			SubmissionType:     "newsletter",
			Email:              form.Email,
			FirstName:          form.FirstName,
			FullName:           "",
			AgencyName:         "",
			WorkflowBottleneck: "",
			WorkflowName:       "",
			AssessmentResult:   "",
			AssessmentScore:    "",
			AssessmentAnswers:  "",
			ClientVolume:       "",
			PrimaryVertical:    "",
			DeliveryConstraint: "",
			ConsentGiven:       1,
			UtmSource:          "",
			UtmMedium:          "",
			UtmCampaign:        "",
			UtmContent:         "",
			Referrer:           "",
			SpamStatus:         string(spamStatusAccepted),
			SpamReasons:        "[]",
			LeadStatus:         string(leadStatusNew),
		})
		return err
	})
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	app.telemetry.recordLeadOutcome(r.Context(), "newsletter", "accepted")
	app.logger.InfoContext(
		r.Context(),
		"lead submission accepted",
		"event", "lead.submission.accepted",
		"form", "newsletter",
		"request_id", requestIDFromContext(r.Context()),
	)
	app.setConversionCookie(w, r, "newsletter")

	http.Redirect(w, r, "/newsletter?submitted=true", http.StatusSeeOther)
}

func (app *application) setConversionCookie(w http.ResponseWriter, r *http.Request, conversion string) {
	token, err := app.formProof.Create("conversion:" + conversion)
	if err != nil {
		app.logger.ErrorContext(
			r.Context(),
			"conversion cookie creation failed",
			"event", "analytics.conversion_cookie.failed",
			"conversion", conversion,
			"request_id", requestIDFromContext(r.Context()),
			"error", err,
		)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     conversionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(conversionCookieMaxAge.Seconds()),
		HttpOnly: true,
		Secure:   app.cfg.environment == "production",
		SameSite: http.SameSiteLaxMode,
	})
}

func (app *application) consumeConversionCookie(
	w http.ResponseWriter,
	r *http.Request,
	conversion string,
) bool {
	cookie, err := r.Cookie(conversionCookieName)
	if err != nil {
		return false
	}
	http.SetCookie(w, &http.Cookie{
		Name:     conversionCookieName,
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   app.cfg.environment == "production",
		SameSite: http.SameSiteLaxMode,
	})
	return app.formProof.VerifyRecent(
		cookie.Value,
		"conversion:"+conversion,
		conversionCookieMaxAge,
	) == nil
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

func (form *NewsletterForm) validate() {
	validateOptionalLength(form.Errors, "first_name", form.FirstName, 200)
	validateEmailField(form.Errors, "email", form.Email, "Enter your email address.")
	if !form.Consent {
		form.Errors["consent"] = "Agree to receive the newsletter before subscribing."
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
	validateEmailField(errors, "corporate_email", email, "Enter your corporate email.")
}

func validateEmailField(errors map[string]string, field, email, requiredMessage string) {
	if email == "" {
		errors[field] = requiredMessage
		return
	}
	if len(email) > 320 {
		errors[field] = "Enter a valid email address."
		return
	}

	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		errors[field] = "Enter a valid email address."
	}
}

func (app *application) renderContact(w http.ResponseWriter, r *http.Request, status int, form ContactForm) {
	formSubmitted := r.URL.Query().Get("submitted") == "true"
	formProof := ""
	if !formSubmitted {
		var err error
		formProof, err = app.formProof.Create("contact")
		if err != nil {
			app.serverError(w, r, err)
			return
		}
	}
	analyticsConversion := ""
	if formSubmitted && app.consumeConversionCookie(w, r, "contact") {
		analyticsConversion = "generate_lead"
	}
	data := PageData{
		PageTitle:            "Commercial Contact | Tabiri Analytics",
		PageDescription:      "Contact Tabiri's commercial team to plan secure AI deployment, workflow modernization, and data-protected operations for your organization.",
		OpenGraphDescription: "Connect with Tabiri's commercial team to evaluate secure AI solutions for business operations.",
		CanonicalURL:         "https://tabirianalytics.com/contact",
		CurrentPage:          "contact",
		CurrentYear:          app.now().Year(),
		FormSubmitted:        formSubmitted,
		LeadSource:           contactLeadSource(r.URL.Query().Get("lead_source"), form),
		AnalyticsConversion:  analyticsConversion,
		AnalyticsLeadType:    "commercial_contact",
		ContactForm:          form,
		TurnstileSiteKey:     app.cfg.turnstileSiteKey,
		FormProof:            formProof,
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
	formSubmitted := r.URL.Query().Get("submitted") == "true"
	formProof := ""
	if !formSubmitted {
		var err error
		formProof, err = app.formProof.Create("partner_contact")
		if err != nil {
			app.serverError(w, r, err)
			return
		}
	}
	analyticsConversion := ""
	if formSubmitted && app.consumeConversionCookie(w, r, "partner_contact") {
		analyticsConversion = "generate_lead"
	}
	data := PageData{
		PageTitle:            "Partner Contact | Channel Partner Application | Tabiri",
		PageDescription:      "Apply to become a Tabiri Analytics channel partner and co-deliver secure AI solutions for enterprise and public-sector clients.",
		OpenGraphDescription: "Submit your channel partner application and tell us how your team delivers secure AI outcomes at scale.",
		CanonicalURL:         "https://tabirianalytics.com/partner-contact",
		CurrentPage:          "partner-contact",
		CurrentYear:          app.now().Year(),
		FormSubmitted:        formSubmitted,
		LeadSource:           "partner",
		AnalyticsConversion:  analyticsConversion,
		AnalyticsLeadType:    "partner_contact",
		PartnerContactForm:   form,
		TurnstileSiteKey:     app.cfg.turnstileSiteKey,
		FormProof:            formProof,
	}
	app.renderContactTemplate(w, r, status, data, "./ui/html/pages/partner-contact.tmpl")
}

func (app *application) renderNewsletter(w http.ResponseWriter, r *http.Request, status int, form NewsletterForm) {
	formSubmitted := r.URL.Query().Get("submitted") == "true"
	formProof := ""
	if !formSubmitted {
		var err error
		formProof, err = app.formProof.Create("newsletter")
		if err != nil {
			app.serverError(w, r, err)
			return
		}
	}
	analyticsConversion := ""
	if formSubmitted && app.consumeConversionCookie(w, r, "newsletter") {
		analyticsConversion = "sign_up"
	}
	data := PageData{
		PageTitle:            "Newsletter | Tabiri Analytics",
		PageDescription:      "Get practical analysis on secure AI adoption, governance, operations, and digital workforce design from Tabiri Analytics.",
		OpenGraphDescription: "Practical ideas for leading secure AI adoption, delivered to your inbox.",
		CanonicalURL:         "https://tabirianalytics.com/newsletter",
		CurrentPage:          "newsletter",
		CurrentYear:          app.now().Year(),
		FormSubmitted:        formSubmitted,
		AnalyticsConversion:  analyticsConversion,
		AnalyticsLeadType:    "newsletter",
		NewsletterForm:       form,
		TurnstileSiteKey:     app.cfg.turnstileSiteKey,
		FormProof:            formProof,
	}
	app.renderContactTemplate(w, r, status, data, "./ui/html/pages/newsletter.tmpl")
}

func (app *application) renderContactTemplate(w http.ResponseWriter, r *http.Request, status int, data PageData, page string) {
	name := trimTemplateExtension(filepath.Base(page))
	app.render(w, r, status, name, data)
}
