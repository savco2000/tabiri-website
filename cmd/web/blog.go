package main

import (
	"html/template"
	"net/http"
	"time"
)

// blogPost holds the per-article metadata needed to render a published post
// and its page-level SEO tags.
type blogPost struct {
	Template             string
	PageTitle            string
	PageDescription      string
	OpenGraphDescription string
	OpenGraphImage       string
}

// blogPosts maps the slug used in the /blog/{id} URL to its article. Adding a
// post here is all that is required to publish it at /blog/{slug}.
var blogPosts = map[string]blogPost{
	"why-secure-ai-teams-outperform-generic-automation-tools": {
		Template:             "./ui/html/pages/blog-post-1.tmpl",
		PageTitle:            "Why Secure AI Teams Outperform Generic Automation Tools | Tabiri Analytics",
		PageDescription:      "Why governance, controlled isolation, and human oversight matter more than raw model performance in secure enterprise AI workflows.",
		OpenGraphDescription: "The difference between a useful AI workflow and an operational risk is governance, packaging, and the way work is isolated inside trusted environments.",
		OpenGraphImage:       "https://tabirianalytics.com/static/images/tabiri_office_environment.avif",
	},
	"inside-a-regulated-fintech-agentic-secops-penetration-test": {
		Template:             "./ui/html/pages/blog-post-2.tmpl",
		PageTitle:            "Inside a Regulated Fintech's Agentic SecOps Penetration Test | Tabiri Analytics",
		PageDescription:      "How Tabiri Analytics combined machine-speed coverage with human judgment to assess live payment infrastructure and expose material risk quickly.",
		OpenGraphDescription: "A practical look at assessing live payment infrastructure with Agentic SecOps, and the remediation path it gave a regulated fintech.",
		OpenGraphImage:       "https://tabirianalytics.com/static/images/fintech-compliance.avif",
	},
	"from-alert-queues-to-agentic-defense": {
		Template:             "./ui/html/pages/blog-post-3.tmpl",
		PageTitle:            "From Alert Queues to Agentic Defense | Tabiri Analytics",
		PageDescription:      "How CDP-MapSync's continuous monitoring evolved from endpoint telemetry and engineer-led investigation into human-governed Agentic SecOps.",
		OpenGraphDescription: "How a continuous-monitoring partnership grew from engineer-led log analysis into a human-governed Agentic SecOps architecture.",
		OpenGraphImage:       "https://tabirianalytics.com/static/images/municipal-agentic-secops.avif",
	},
}

func (app *application) blogPost(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("id")

	post, ok := blogPosts[slug]
	if !ok {
		app.notFound(w, r)
		return
	}

	w.Header().Add("Server", "Go")

	data := PageData{
		PageTitle:            post.PageTitle,
		PageDescription:      post.PageDescription,
		OpenGraphDescription: post.OpenGraphDescription,
		CanonicalURL:         "https://tabirianalytics.com/blog/" + slug,
		OpenGraphImage:       post.OpenGraphImage,
		CurrentPage:          "blog",
		CurrentYear:          time.Now().Year(),
	}

	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		post.Template,
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
