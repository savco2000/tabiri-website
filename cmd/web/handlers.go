package main

import (
	"bytes"
	"html/template"
	"net/http"
	"time"
)

type PageData struct {
	PageTitle            string
	PageDescription      string
	OpenGraphDescription string
	CurrentPage          string
	CurrentYear          int
	FormSubmitted        bool
	ContactForm          ContactForm
	PartnerContactForm   PartnerContactForm
}

// Change the signature of the home handler so it is defined as a method against
// *application.
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Server", "Go")

	data := PageData{
		PageTitle:            "Home - Secure AI-Powered Employees | Tabiri Analytics",
		PageDescription:      "Tabiri Analytics builds secure AI employees that automate enterprise workflows while protecting sensitive data.",
		OpenGraphDescription: "Deploy secure, agentic AI teammates for operations, service, and compliance-intensive environments.",
		CurrentPage:          "home",
		CurrentYear:          time.Now().Year(),
	}

	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/pages/home.tmpl",
		"./ui/html/partials/footer.tmpl",
	}

	ts, err := template.ParseFiles(files...)

	if err != nil {
		app.serverError(w, r, err)
		return
	}
	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) digitalEmployees(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Server", "Go")

	data := PageData{
		PageTitle:            "Digital Employees for Enterprise Automation | Tabiri Analytics",
		PageDescription:      "Return capacity to your team with secure Digital Employees that handle repetitive workflows under human oversight, policy guardrails, and auditable controls.",
		OpenGraphDescription: "Deploy secure, agentic AI teammates for operations, service, and compliance-intensive environments.",
		CurrentPage:          "digital-employees",
		CurrentYear:          time.Now().Year(),
	}

	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/pages/digital-employees.tmpl",
		"./ui/html/partials/footer.tmpl",
	}

	ts, err := template.ParseFiles(files...)

	if err != nil {
		app.serverError(w, r, err)
		return
	}
	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) operationalAssessment(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Server", "Go")

	data := PageData{
		PageTitle:            "Operational Assessment | Tabiri Analytics",
		PageDescription:      "Turn cybersecurity uncertainty into a prioritized improvement roadmap with an adaptive operational assessment across six essential security domains.",
		OpenGraphDescription: "Turn structured responses across six cybersecurity domains into a clear view of operational maturity with Tabiri Analytics Management Systems.",
		CurrentPage:          "operational-assessment",
		CurrentYear:          time.Now().Year(),
	}

	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/pages/operational-assessment.tmpl",
		"./ui/html/partials/footer.tmpl",
	}

	ts, err := template.ParseFiles(files...)

	if err != nil {
		app.serverError(w, r, err)
		return
	}
	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) penetrationTesting(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Server", "Go")

	data := PageData{
		PageTitle:            "Penetration Testing Services | Tabiri Analytics",
		PageDescription:      "Find exploitable weaknesses faster with engineer-led, agent-assisted penetration testing, validated findings, same-day escalation, and compliance-ready evidence.",
		OpenGraphDescription: "Get evidence-backed findings, practical remediation guidance, and compliance-ready reporting from an engineer-led penetration test.",
		CurrentPage:          "penetration-testing",
		CurrentYear:          time.Now().Year(),
	}

	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/pages/penetration-testing.tmpl",
		"./ui/html/partials/footer.tmpl",
	}

	ts, err := template.ParseFiles(files...)

	if err != nil {
		app.serverError(w, r, err)
		return
	}
	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) managedCybersecurityServices(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Server", "Go")

	data := PageData{
		PageTitle:            "Managed Cybersecurity Services | Tabiri Analytics",
		PageDescription:      "Protect networks and endpoints with 24/7 telemetry, AI-assisted threat triage, controlled containment, and engineer-led managed cybersecurity services.",
		OpenGraphDescription: "Turn continuous security telemetry into prioritized investigation, controlled response, and evidence your team can act on.",
		CurrentPage:          "managed-cybersecurity-services",
		CurrentYear:          time.Now().Year(),
	}

	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/pages/managed-cybersecurity-services.tmpl",
		"./ui/html/partials/footer.tmpl",
	}

	ts, err := template.ParseFiles(files...)

	if err != nil {
		app.serverError(w, r, err)
		return
	}
	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) cybersecurityInsurance(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Server", "Go")

	data := PageData{
		PageTitle:            "Cybersecurity Insurance | Tabiri Analytics",
		PageDescription:      "Prepare for cyber insurance with Tabiri's penetration testing, continuous monitoring, and practical risk remediation services.",
		OpenGraphDescription: "Combine cyber risk transfer with evidence-led security testing and continuous monitoring designed to reduce the likelihood and impact of a breach.",
		CurrentPage:          "cybersecurity-insurance",
		CurrentYear:          time.Now().Year(),
	}

	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/pages/cybersecurity-insurance.tmpl",
		"./ui/html/partials/footer.tmpl",
	}

	ts, err := template.ParseFiles(files...)

	if err != nil {
		app.serverError(w, r, err)
		return
	}
	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) dataProtection(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Server", "Go")

	data := PageData{
		PageTitle:            "Data Protection | Tabiri Analytics",
		PageDescription:      "Explore Tabiri's sovereign data architecture, isolated compute, and encryption model for secure enterprise AI operations.",
		OpenGraphDescription: "See how Tabiri protects sensitive data with private nodes, strict isolation, and enterprise-grade controls.",
		CurrentPage:          "data-protection",
		CurrentYear:          time.Now().Year(),
	}

	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/pages/data-protection.tmpl",
		"./ui/html/partials/footer.tmpl",
	}

	ts, err := template.ParseFiles(files...)

	if err != nil {
		app.serverError(w, r, err)
		return
	}
	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) caseStudies(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Server", "Go")

	data := PageData{
		PageTitle:            "Case Studies | Secure AI Outcomes | Tabiri Analytics",
		PageDescription:      "Explore real-world Tabiri deployments delivering secure AI automation and cybersecurity outcomes across legal, financial, and high-trust operations.",
		OpenGraphDescription: "See measurable results from secure AI deployments and engineer-led cybersecurity engagements in high-trust environments.",
		CurrentPage:          "case-studies",
		CurrentYear:          time.Now().Year(),
	}

	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/pages/case-studies.tmpl",
		"./ui/html/partials/footer.tmpl",
	}

	ts, err := template.ParseFiles(files...)

	if err != nil {
		app.serverError(w, r, err)
		return
	}
	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) channelPartners(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Server", "Go")

	data := PageData{
		PageTitle:            "Channel Partner Program | Tabiri Analytics",
		PageDescription:      "Join Tabiri's channel partner program to deliver secure AI automation services at scale.",
		OpenGraphDescription: "Apply to become a regional partner and co-deliver sovereign AI solutions for enterprise clients.",
		CurrentPage:          "channel-partners",
		CurrentYear:          time.Now().Year(),
	}

	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/pages/channel-partners.tmpl",
		"./ui/html/partials/footer.tmpl",
	}

	ts, err := template.ParseFiles(files...)

	if err != nil {
		app.serverError(w, r, err)
		return
	}
	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) capabilityStatement(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Server", "Go")

	data := PageData{
		PageTitle:            "Capability Statement | Tabiri Analytics",
		PageDescription:      "Review Tabiri Analytics capabilities in secure AI automation, cyber operations, and compliance-focused delivery.",
		OpenGraphDescription: "Explore core competencies, past performance, and differentiators for secure AI and cybersecurity services.",
		CurrentPage:          "capability-statement",
		CurrentYear:          time.Now().Year(),
	}

	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/pages/capability-statement.tmpl",
		"./ui/html/partials/footer.tmpl",
	}

	ts, err := template.ParseFiles(files...)

	if err != nil {
		app.serverError(w, r, err)
		return
	}
	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) blog(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Server", "Go")

	data := PageData{
		PageTitle:            "Blog | Tabiri Analytics",
		PageDescription:      "Read the latest insights and updates from Tabiri Analytics on secure AI automation, cyber operations, and compliance-focused delivery.",
		OpenGraphDescription: "Stay informed with expert articles, case studies, and news from Tabiri Analytics on secure AI and cybersecurity services.",
		CurrentPage:          "blog",
		CurrentYear:          time.Now().Year(),
	}

	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/pages/blog.tmpl",
		"./ui/html/partials/footer.tmpl",
	}

	ts, err := template.ParseFiles(files...)

	if err != nil {
		app.serverError(w, r, err)
		return
	}
	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) about(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Server", "Go")

	data := PageData{
		PageTitle:            "About Us | Tabiri Analytics",
		PageDescription:      "Meet the Tabiri Analytics team building secure, sovereign AI systems for high-trust enterprise and public-sector environments.",
		OpenGraphDescription: "Learn about the engineers and operators behind Tabiri's secure AI automation platform.",
		CurrentPage:          "about",
		CurrentYear:          time.Now().Year(),
	}

	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/pages/about.tmpl",
		"./ui/html/partials/footer.tmpl",
	}

	ts, err := template.ParseFiles(files...)

	if err != nil {
		app.serverError(w, r, err)
		return
	}
	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) contact(w http.ResponseWriter, r *http.Request) {
	app.renderContact(w, r, http.StatusOK, ContactForm{})
}

func (app *application) partnerContact(w http.ResponseWriter, r *http.Request) {
	app.renderPartnerContact(w, r, http.StatusOK, PartnerContactForm{})
}

func (app *application) newsletter(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Server", "Go")

	data := PageData{
		PageTitle:            "Newsletter | Tabiri Analytics",
		PageDescription:      "Subscribe to Tabiri Analytics for practical guidance on secure AI operations, governance, and digital workforce design.",
		OpenGraphDescription: "Get practical insights for building secure, accountable AI systems delivered to your inbox.",
		CurrentPage:          "newsletter",
		CurrentYear:          time.Now().Year(),
	}

	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/pages/newsletter.tmpl",
		"./ui/html/partials/footer.tmpl",
	}

	ts, err := template.ParseFiles(files...)

	if err != nil {
		app.serverError(w, r, err)
		return
	}
	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) automationAssessment(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Server", "Go")

	data := PageData{
		PageTitle:            "Automation Readiness Assessment | Tabiri Analytics",
		PageDescription:      "Take Tabiri Analytics' five-question assessment to see whether your business workflows are ready for secure AI automation.",
		OpenGraphDescription: "Answer five quick questions to discover whether your business is a good candidate for secure AI automation.",
		CurrentPage:          "automation-assessment",
		CurrentYear:          time.Now().Year(),
	}

	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/pages/automation-assessment.tmpl",
		"./ui/html/partials/footer.tmpl",
	}

	ts, err := template.ParseFiles(files...)

	if err != nil {
		app.serverError(w, r, err)
		return
	}
	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) notFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Server", "Go")

	data := PageData{
		PageTitle:            "Page Not Found | Tabiri Analytics",
		PageDescription:      "The requested page could not be found. Return to the Tabiri Analytics homepage to explore our secure AI and cybersecurity services.",
		OpenGraphDescription: "The requested Tabiri Analytics page could not be found.",
		CurrentYear:          time.Now().Year(),
	}

	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/pages/404.tmpl",
		"./ui/html/partials/footer.tmpl",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	var page bytes.Buffer
	if err := ts.ExecuteTemplate(&page, "base", data); err != nil {
		app.serverError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNotFound)
	if _, err := page.WriteTo(w); err != nil {
		app.logger.Error(err.Error(), "method", r.Method, "uri", r.URL.RequestURI())
	}
}
