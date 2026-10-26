package main

import (
	"net/http"
)

type PageData struct {
	PageTitle            string
	PageDescription      string
	OpenGraphDescription string
	CanonicalURL         string
	OpenGraphImage       string
	CurrentPage          string
	CurrentYear          int
	FormSubmitted        bool
	LeadSource           string
	ContactForm          ContactForm
	PartnerContactForm   PartnerContactForm
	NewsletterForm       NewsletterForm
	TurnstileSiteKey     string
	FormProof            string
}

type pageSpec struct {
	template             string
	title                string
	description          string
	openGraphDescription string
	canonicalURL         string
	openGraphImage       string
	currentPage          string
}

var pageSpecs = map[string]pageSpec{
	"home": {
		template:             "home",
		title:                "Secure AI-Powered Digital Employees | Tabiri Analytics",
		description:          "Tabiri Analytics builds secure AI-powered Digital Employees that automate enterprise workflows while protecting sensitive data.",
		openGraphDescription: "Deploy secure AI-powered Digital Employees for operations, service, and compliance-intensive environments while keeping people in control.",
		canonicalURL:         "https://tabirianalytics.com/",
		currentPage:          "home",
	},
	"digital-employees": {
		template:             "digital-employees",
		title:                "Digital Employees for Enterprise Automation | Tabiri Analytics",
		description:          "Return capacity to your team with secure Digital Employees that handle repetitive workflows under human oversight, policy guardrails, and auditable controls.",
		openGraphDescription: "Deploy secure, agentic AI teammates for operations, service, and compliance-intensive environments.",
		canonicalURL:         "https://tabirianalytics.com/digital-employees",
		currentPage:          "digital-employees",
	},
	"operational-assessment": {
		template:             "operational-assessment",
		title:                "Operational Assessment | Tabiri Analytics",
		description:          "Turn cybersecurity uncertainty into a prioritized improvement roadmap with an adaptive operational assessment across six essential security domains.",
		openGraphDescription: "Turn structured responses across six cybersecurity domains into a clear view of operational maturity with Tabiri Analytics Management Systems.",
		canonicalURL:         "https://tabirianalytics.com/operational-assessment",
		currentPage:          "operational-assessment",
	},
	"penetration-testing": {
		template:             "penetration-testing",
		title:                "Penetration Testing Services | Tabiri Analytics",
		description:          "Find exploitable weaknesses faster with engineer-led, agent-assisted penetration testing, validated findings, same-day escalation, and compliance-ready evidence.",
		openGraphDescription: "Get evidence-backed findings, practical remediation guidance, and compliance-ready reporting from an engineer-led penetration test.",
		canonicalURL:         "https://tabirianalytics.com/penetration-testing",
		currentPage:          "penetration-testing",
	},
	"managed-cybersecurity-services": {
		template:             "managed-cybersecurity-services",
		title:                "Managed Cybersecurity Services | Tabiri Analytics",
		description:          "Protect networks and endpoints with 24/7 telemetry, AI-assisted threat triage, controlled containment, and engineer-led managed cybersecurity services.",
		openGraphDescription: "Turn continuous security telemetry into prioritized investigation, controlled response, and evidence your team can act on.",
		canonicalURL:         "https://tabirianalytics.com/managed-cybersecurity-services",
		currentPage:          "managed-cybersecurity-services",
	},
	"cybersecurity-insurance": {
		template:             "cybersecurity-insurance",
		title:                "Cybersecurity Insurance | Tabiri Analytics",
		description:          "Prepare for cyber insurance with Tabiri's penetration testing, continuous monitoring, and practical risk remediation services.",
		openGraphDescription: "Combine cyber risk transfer with evidence-led security testing and continuous monitoring designed to reduce the likelihood and impact of a breach.",
		canonicalURL:         "https://tabirianalytics.com/cybersecurity-insurance",
		currentPage:          "cybersecurity-insurance",
	},
	"data-protection": {
		template:             "data-protection",
		title:                "Data Protection | Tabiri Analytics",
		description:          "Explore Tabiri's sovereign data architecture, isolated compute, and encryption model for secure enterprise AI operations.",
		openGraphDescription: "See how Tabiri protects sensitive data with private nodes, strict isolation, and enterprise-grade controls.",
		canonicalURL:         "https://tabirianalytics.com/data-protection",
		currentPage:          "data-protection",
	},
	"case-studies": {
		template:             "case-studies",
		title:                "Case Studies | Secure AI Outcomes | Tabiri Analytics",
		description:          "Explore real-world Tabiri deployments delivering secure AI automation and cybersecurity outcomes across legal, financial, and high-trust operations.",
		openGraphDescription: "See measurable results from secure AI deployments and engineer-led cybersecurity engagements in high-trust environments.",
		canonicalURL:         "https://tabirianalytics.com/case-studies",
		currentPage:          "case-studies",
	},
	"channel-partners": {
		template:             "channel-partners",
		title:                "Channel Partner Program | Tabiri Analytics",
		description:          "Join Tabiri's channel partner program to deliver secure AI automation services at scale.",
		openGraphDescription: "Apply to become a regional partner and co-deliver sovereign AI solutions for enterprise clients.",
		canonicalURL:         "https://tabirianalytics.com/channel-partners",
		currentPage:          "channel-partners",
	},
	"capability-statement": {
		template:             "capability-statement",
		title:                "Capability Statement | Tabiri Analytics",
		description:          "Review Tabiri Analytics capabilities in secure AI automation, cyber operations, and compliance-focused delivery.",
		openGraphDescription: "Explore core competencies, past performance, and differentiators for secure AI and cybersecurity services.",
		canonicalURL:         "https://tabirianalytics.com/capability-statement",
		currentPage:          "capability-statement",
	},
	"blog": {
		template:             "blog",
		title:                "Blog | Tabiri Analytics",
		description:          "Read the latest insights and updates from Tabiri Analytics on secure AI automation, cyber operations, and compliance-focused delivery.",
		openGraphDescription: "Stay informed with expert articles, case studies, and news from Tabiri Analytics on secure AI and cybersecurity services.",
		canonicalURL:         "https://tabirianalytics.com/blog",
		currentPage:          "blog",
	},
	"about": {
		template:             "about",
		title:                "About Us | Tabiri Analytics",
		description:          "Meet the Tabiri Analytics team building secure, sovereign AI systems for high-trust enterprise and public-sector environments.",
		openGraphDescription: "Learn about the engineers and operators behind Tabiri's secure AI automation platform.",
		canonicalURL:         "https://tabirianalytics.com/about",
		currentPage:          "about",
	},
	"contact": {
		template: "contact",
	},
	"partner-contact": {
		template: "partner-contact",
	},
	"newsletter": {
		template:             "newsletter",
		title:                "Newsletter | Tabiri Analytics",
		description:          "Subscribe to Tabiri Analytics for practical guidance on secure AI operations, governance, and digital workforce design.",
		openGraphDescription: "Get practical insights for building secure, accountable AI systems delivered to your inbox.",
		currentPage:          "newsletter",
	},
	"automation-assessment": {
		template:             "automation-assessment",
		title:                "Automation Readiness Assessment | Tabiri Analytics",
		description:          "Take Tabiri Analytics' five-question assessment to see whether your business workflows are ready for secure AI automation.",
		openGraphDescription: "Answer five quick questions to discover whether your business is a good candidate for secure AI automation.",
		canonicalURL:         "https://tabirianalytics.com/automation-assessment",
		openGraphImage:       "https://tabirianalytics.com/static/images/automation-assessment-social.png",
		currentPage:          "automation-assessment",
	},
	"privacy-policy": {
		template:             "privacy-policy",
		title:                "Privacy Policy (Draft) | Tabiri Analytics",
		description:          "Draft privacy policy for the Tabiri Analytics website. This draft requires review and approval before publication.",
		openGraphDescription: "Draft privacy policy for the Tabiri Analytics website.",
		canonicalURL:         "https://tabirianalytics.com/privacy-policy",
		currentPage:          "privacy-policy",
	},
	"terms-of-service": {
		template:             "terms-of-service",
		title:                "Terms of Service (Draft) | Tabiri Analytics",
		description:          "Draft terms of service for the Tabiri Analytics website. This draft requires review and approval before publication.",
		openGraphDescription: "Draft terms of service for the Tabiri Analytics website.",
		canonicalURL:         "https://tabirianalytics.com/terms-of-service",
		currentPage:          "terms-of-service",
	},
	"security-disclosure": {
		template:             "security-disclosure",
		title:                "Security Disclosure (Draft) | Tabiri Analytics",
		description:          "Draft vulnerability disclosure guidance for the Tabiri Analytics website. This draft requires review and approval before publication.",
		openGraphDescription: "Draft security disclosure guidance for the Tabiri Analytics website.",
		canonicalURL:         "https://tabirianalytics.com/security-disclosure",
		currentPage:          "security-disclosure",
	},
	"not-found": {
		template:             "404",
		title:                "Page Not Found | Tabiri Analytics",
		description:          "The requested page could not be found. Return to the Tabiri Analytics homepage to explore our secure AI and cybersecurity services.",
		openGraphDescription: "The requested Tabiri Analytics page could not be found.",
	},
}

func (app *application) pageData(spec pageSpec) PageData {
	return PageData{
		PageTitle:            spec.title,
		PageDescription:      spec.description,
		OpenGraphDescription: spec.openGraphDescription,
		CanonicalURL:         spec.canonicalURL,
		OpenGraphImage:       spec.openGraphImage,
		CurrentPage:          spec.currentPage,
		CurrentYear:          app.now().Year(),
	}
}

func (app *application) renderNamedPage(w http.ResponseWriter, r *http.Request, name string, status int) {
	spec := pageSpecs[name]
	app.render(w, r, status, spec.template, app.pageData(spec))
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	app.renderNamedPage(w, r, "home", http.StatusOK)
}

func (app *application) digitalEmployees(w http.ResponseWriter, r *http.Request) {
	app.renderNamedPage(w, r, "digital-employees", http.StatusOK)
}

func (app *application) operationalAssessment(w http.ResponseWriter, r *http.Request) {
	app.renderNamedPage(w, r, "operational-assessment", http.StatusOK)
}

func (app *application) penetrationTesting(w http.ResponseWriter, r *http.Request) {
	app.renderNamedPage(w, r, "penetration-testing", http.StatusOK)
}

func (app *application) managedCybersecurityServices(w http.ResponseWriter, r *http.Request) {
	app.renderNamedPage(w, r, "managed-cybersecurity-services", http.StatusOK)
}

func (app *application) cybersecurityInsurance(w http.ResponseWriter, r *http.Request) {
	app.renderNamedPage(w, r, "cybersecurity-insurance", http.StatusOK)
}

func (app *application) dataProtection(w http.ResponseWriter, r *http.Request) {
	app.renderNamedPage(w, r, "data-protection", http.StatusOK)
}

func (app *application) caseStudies(w http.ResponseWriter, r *http.Request) {
	app.renderNamedPage(w, r, "case-studies", http.StatusOK)
}

func (app *application) channelPartners(w http.ResponseWriter, r *http.Request) {
	app.renderNamedPage(w, r, "channel-partners", http.StatusOK)
}

func (app *application) capabilityStatement(w http.ResponseWriter, r *http.Request) {
	app.renderNamedPage(w, r, "capability-statement", http.StatusOK)
}

func (app *application) blog(w http.ResponseWriter, r *http.Request) {
	app.renderNamedPage(w, r, "blog", http.StatusOK)
}

func (app *application) about(w http.ResponseWriter, r *http.Request) {
	app.renderNamedPage(w, r, "about", http.StatusOK)
}

func (app *application) contact(w http.ResponseWriter, r *http.Request) {
	app.renderContact(w, r, http.StatusOK, contactFormFromQuery(r.URL.Query()))
}

func (app *application) partnerContact(w http.ResponseWriter, r *http.Request) {
	app.renderPartnerContact(w, r, http.StatusOK, PartnerContactForm{})
}

func (app *application) newsletter(w http.ResponseWriter, r *http.Request) {
	app.renderNewsletter(w, r, http.StatusOK, NewsletterForm{})
}

func (app *application) automationAssessment(w http.ResponseWriter, r *http.Request) {
	app.renderNamedPage(w, r, "automation-assessment", http.StatusOK)
}

func (app *application) privacyPolicy(w http.ResponseWriter, r *http.Request) {
	app.renderNamedPage(w, r, "privacy-policy", http.StatusOK)
}

func (app *application) termsOfService(w http.ResponseWriter, r *http.Request) {
	app.renderNamedPage(w, r, "terms-of-service", http.StatusOK)
}

func (app *application) securityDisclosure(w http.ResponseWriter, r *http.Request) {
	app.renderNamedPage(w, r, "security-disclosure", http.StatusOK)
}

func (app *application) notFound(w http.ResponseWriter, r *http.Request) {
	app.renderNamedPage(w, r, "not-found", http.StatusNotFound)
}
