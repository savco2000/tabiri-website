package main

import (
	"net/http"
	"path/filepath"
)

// The routes() method returns a servemux containing our application routes.
func (app *application) routes() *http.ServeMux {
	mux := http.NewServeMux()

	fileServer := http.FileServer(neuteredFileSystem{http.Dir(app.cfg.staticDir)})
	mux.Handle("GET /static", http.NotFoundHandler())
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

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
	//mux.HandleFunc("GET /blog/{id}", app.blogPost)
	mux.HandleFunc("GET /about", app.about)
	mux.HandleFunc("GET /contact", app.contact)
	mux.HandleFunc("GET /partner-contact", app.partnerContact)
	mux.HandleFunc("GET /newsletter", app.newsletter)
	mux.HandleFunc("GET /automation-assessment", app.automationAssessment)
	mux.HandleFunc("GET /", app.notFound)

	return mux
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
