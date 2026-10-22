# Deployment and edge configuration

## Hosting boundary

The existing production site uses the `tabiri-public-sites` project,
`tabirianalytics-backend` backend bucket, `www.tabirianalytics.com` Cloud Storage
bucket, and `tabirianalytics-map` HTTPS load balancer URL map. TLS terminates at
the load balancer.

Cloud Storage serves files, not Go handlers or Go templates. This repository's
contact, partner-contact, and newsletter POST handlers, signed form proofs, and
Turnstile verification need a running Go backend. Uploading templates or rendering
forms once does **not** provide working form submissions. A production Go
deployment also needs persistent SQLite storage; an ephemeral instance filesystem
is not sufficient. This change does not provision a backend or export the site
as static HTML.

## Discovery files and health

The files in `cmd/web/public/` are embedded in the Go binary and served at:

| URL | Content type | Cache-Control |
| --- | --- | --- |
| `/robots.txt` | `text/plain; charset=utf-8` | `public, max-age=3600` |
| `/sitemap.xml` | `application/xml; charset=utf-8` | `public, max-age=3600` |
| `/.well-known/security.txt` | `text/plain; charset=utf-8` | `public, max-age=3600` |
| `/healthz` | `text/plain; charset=utf-8` | `no-store` |

For the Go app, `/healthz` is a lightweight HTTP liveness probe returning `200`
and `ok`. It does not check database readiness or external providers. On the
bucket, it is only a delivery probe for a static object, not backend health.

The sitemap lists published canonical page URLs and all blog posts, excluding
the unapproved legal drafts. Update it when publishing or removing pages/posts;
the Go tests fail if it drifts from the available pages. Before uploading it,
verify every listed URL is also published on the static site at the exact
extensionless path. The sitemap does not create bucket routes.

The monitored security reporting contact is
`mailto:savannah.kadima@tabirianalytics.com`. Renew the `Expires` field in
`cmd/web/public/.well-known/security.txt` **before October 1, 2027**, then redeploy
it. No finalized disclosure policy or safe-harbor commitment is implied.

### Upload to the existing bucket (manual, requires approval)

Run from the repository root after reviewing the destination and existing
objects. These commands overwrite only the four named discovery/probe objects;
they do not deploy the HTML, CSS, or JavaScript changes.

```bash
gcloud storage cp cmd/web/public/robots.txt gs://www.tabirianalytics.com/robots.txt \
  --project=tabiri-public-sites --content-type='text/plain; charset=utf-8' --cache-control='public, max-age=3600'
gcloud storage cp cmd/web/public/sitemap.xml gs://www.tabirianalytics.com/sitemap.xml \
  --project=tabiri-public-sites --content-type='application/xml; charset=utf-8' --cache-control='public, max-age=3600'
gcloud storage cp cmd/web/public/.well-known/security.txt gs://www.tabirianalytics.com/.well-known/security.txt \
  --project=tabiri-public-sites --content-type='text/plain; charset=utf-8' --cache-control='public, max-age=3600'
gcloud storage cp cmd/web/public/healthz gs://www.tabirianalytics.com/healthz \
  --project=tabiri-public-sites --content-type='text/plain; charset=utf-8' --cache-control='no-store'
```

Check the backend's cache mode first. `FORCE_CACHE_ALL` can override `no-store`;
use `CACHE_ALL_STATIC` or `USE_ORIGIN_HEADERS` if the probe must not be cached.
Invalidate existing cached responses (including cached 404s) for these four exact
paths using `gcloud compute url-maps invalidate-cdn-cache tabirianalytics-map
--project=tabiri-public-sites --path=PATH`, once per path.

## Security headers

The Go app sends CSP on all responses and HSTS when `APP_ENV=production`. It
intentionally does not infer production HTTPS from user-supplied forwarding
headers. Production must be HTTPS at the edge; development does not emit HSTS.
HSTS uses `max-age=31536000` without `includeSubDomains` or preload, so it does
not assert that unrelated subdomains support HTTPS.

For bucket hosting, Go middleware never runs. Apply headers on the backend
bucket instead. Inspect and save the existing configuration before changing it:

```bash
gcloud compute backend-buckets describe tabirianalytics-backend \
  --project=tabiri-public-sites --format=yaml
```

The following is the full proposed custom-response-header list. Merge any
existing unrelated headers before applying it; do not discard existing
configuration. **Do not run without approval to change production.**

```bash
gcloud compute backend-buckets update tabirianalytics-backend \
  --project=tabiri-public-sites \
  --custom-response-header="Strict-Transport-Security: max-age=31536000" \
  --custom-response-header="Content-Security-Policy: default-src 'self'; base-uri 'self'; object-src 'none'; frame-ancestors 'none'; form-action 'self'; script-src 'self' https://challenges.cloudflare.com https://www.googletagmanager.com; style-src 'self' 'unsafe-inline'; img-src 'self' data: https://*.google-analytics.com https://*.googletagmanager.com; font-src 'self'; connect-src 'self' https://challenges.cloudflare.com https://*.google-analytics.com https://*.analytics.google.com https://*.googletagmanager.com https://*.google.com; frame-src https://challenges.cloudflare.com https://www.youtube-nocookie.com https://www.googletagmanager.com" \
  --custom-response-header="X-Content-Type-Options: nosniff" \
  --custom-response-header="X-Frame-Options: DENY" \
  --custom-response-header="Referrer-Policy: strict-origin-when-cross-origin" \
  --custom-response-header="Permissions-Policy: camera=(), microphone=(), geolocation=()"
```

CSP allows local assets, Cloudflare Turnstile, consent-gated Google Tag
Manager/Google Analytics, and the YouTube privacy-enhanced embed. Inline styles
remain allowed for the existing SVG diagram and style attributes; inline
scripts, `eval`, plugins, framing this site, and cross-origin form submissions
are blocked. CSP is not an analytics-consent mechanism; keep the consent gate.

Review the **published GTM container** before enforcement: custom HTML tags,
custom JavaScript variables, Google Ads, preview/debug tools, and third-party
tags can require additional sources or refactoring. Do not broadly enable
`unsafe-inline` or `unsafe-eval` for scripts. On staging, try the policy first
as `Content-Security-Policy-Report-Only`, inspect browser CSP violations, then
switch to the enforcing header above after testing forms, analytics after
accept/reject, SVG rendering, and video playback. Report-only alone does not
protect the site. The backend bucket policy must be tested against the actual
static HTML as well, since it can differ from these templates.

If later routing to the Go app, avoid conflicting edge/app CSP headers:
browsers enforce all supplied policies. Forward dynamic forms to the app and
do not cache pages containing per-request form proofs or form POST responses.

## Verify after rollout

```bash
curl --fail --silent --show-error https://tabirianalytics.com/robots.txt
curl --fail --silent --show-error https://tabirianalytics.com/sitemap.xml
curl --fail --silent --show-error https://tabirianalytics.com/.well-known/security.txt
curl --fail --silent --show-error -D - https://tabirianalytics.com/healthz
curl --fail --silent --show-error -I https://tabirianalytics.com/
```

Verify content types, expected bodies (not a fallback HTML page), security
headers, health cache behavior, exact sitemap URLs, and that the footer no
longer claims "All Systems Operational". Test GET and HEAD on the Go app.
For rollback, restore the saved custom header list and previous object versions;
removing HSTS does not erase an already cached browser HSTS policy.

References: [backend-bucket header flags](https://cloud.google.com/sdk/gcloud/reference/compute/backend-buckets/update),
[Google Tag Manager CSP guidance](https://developers.google.com/tag-platform/security/guides/csp).
