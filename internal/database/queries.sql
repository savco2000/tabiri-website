-- name: CreateContactSubmission :one
INSERT INTO contact_submissions (
    full_name,
    corporate_email,
    workflow_bottleneck
) VALUES (?, ?, ?)
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
