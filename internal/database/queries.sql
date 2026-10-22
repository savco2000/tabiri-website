-- name: CreateContactSubmission :one
INSERT INTO contact_submissions (
    full_name,
    corporate_email,
    workflow_bottleneck,
    workflow_name,
    assessment_result,
    assessment_score,
    assessment_answers,
    utm_source,
    utm_medium,
    utm_campaign,
    utm_content,
    referrer,
    spam_status,
    spam_reasons,
    lead_status,
    turnstile_verified_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
RETURNING *;

-- name: CreateNewsletterSubscription :one
INSERT INTO newsletter_subscriptions (
    first_name,
    email,
    consent_given,
    spam_status,
    spam_reasons,
    lead_status,
    turnstile_verified_at
) VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
RETURNING *;

-- name: CreatePartnerContactSubmission :one
INSERT INTO partner_contact_submissions (
    agency_name,
    corporate_email,
    client_volume,
    primary_vertical,
    delivery_constraint,
    spam_status,
    spam_reasons,
    lead_status,
    turnstile_verified_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
RETURNING *;
