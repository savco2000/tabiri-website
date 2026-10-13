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
    referrer
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: CreatePartnerContactSubmission :one
INSERT INTO partner_contact_submissions (
    agency_name,
    corporate_email,
    client_volume,
    primary_vertical,
    delivery_constraint
) VALUES (?, ?, ?, ?, ?)
RETURNING *;
