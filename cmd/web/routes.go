package main

import (
	"net/http"
	"path/filepath"
)

// The routes() method returns a servemux containing our application routes.
func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(neuteredFileSystem{http.Dir(app.cfg.staticDir)})
	mux.Handle("GET /static", http.NotFoundHandler())
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	mux.Handle("GET /robots.txt", app.discoveryFile("robots.txt", "text/plain; charset=utf-8", "public, max-age=3600"))
	mux.Handle("GET /sitemap.xml", app.discoveryFile("sitemap.xml", "application/xml; charset=utf-8", "public, max-age=3600"))
	mux.Handle("GET /.well-known/security.txt", app.discoveryFile(".well-known/security.txt", "text/plain; charset=utf-8", "public, max-age=3600"))
	mux.Handle("GET /healthz", app.discoveryFile("healthz", "text/plain; charset=utf-8", "no-store"))

	mux.HandleFunc("GET /{$}", app.home)
	mux.HandleFunc("GET /digital-employees", app.digitalEmployees)
	mux.HandleFunc("GET /operational-assessment", app.operationalAssessment)
	mux.HandleFunc("GET /penetration-testing", app.penetrationTesting)
	mux.HandleFunc("GET /managed-cybersecurity-services", app.managedCybersecurityServices)
	mux.HandleFunc("GET /cybersecurity-insurance", app.cybersecurityInsurance)
	mux.HandleFunc("GET /data-protection", app.dataProtection)
	mux.HandleFunc("GET /case-studies", app.caseStudies)
	mux.HandleFunc("GET /channel-partners", app.channelPartners)
	mux.HandleFunc("GET /capability-statement", app.capabilityStatement)
	mux.HandleFunc("GET /blog", app.blog)
	mux.HandleFunc("GET /blog/{id}", app.blogPost)
	mux.HandleFunc("GET /about", app.about)
	mux.HandleFunc("GET /privacy-policy", app.privacyPolicy)
	mux.HandleFunc("GET /terms-of-service", app.termsOfService)
	mux.HandleFunc("GET /security-disclosure", app.securityDisclosure)
	mux.HandleFunc("GET /contact", app.contact)
	mux.Handle(
		"POST /contact",
		app.protectLeadSubmission(
			"contact",
			"contact_submit",
			app.renderContactProtectionError,
			http.HandlerFunc(app.contactSubmit),
		),
	)
	mux.HandleFunc("GET /partner-contact", app.partnerContact)
	mux.Handle(
		"POST /partner-contact",
		app.protectLeadSubmission(
			"partner_contact",
			"partner_contact_submit",
			app.renderPartnerContactProtectionError,
			http.HandlerFunc(app.partnerContactSubmit),
		),
	)
	mux.HandleFunc("GET /newsletter", app.newsletter)
	mux.Handle(
		"POST /newsletter",
		app.protectLeadSubmission(
			"newsletter",
			"newsletter_submit",
			app.renderNewsletterProtectionError,
			http.HandlerFunc(app.newsletterSubmit),
		),
	)
	mux.HandleFunc("GET /automation-assessment", app.automationAssessment)
	mux.HandleFunc("GET /", app.notFound)

	dispatch := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, pattern := mux.Handler(r)
		r.Pattern = pattern
		mux.ServeHTTP(w, r)
	})
	return app.middleware(dispatch)
}

type neuteredFileSystem struct {
	fs http.FileSystem
}

func (nfs neuteredFileSystem) Open(path string) (http.File, error) {
	f, err := nfs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if s.IsDir() {
		index := filepath.Join(path, "index.html")
		if _, err := nfs.fs.Open(index); err != nil {
			closeErr := f.Close()
			if closeErr != nil {
				return nil, closeErr
			}

			return nil, err
		}
	}

	return f, nil
}
