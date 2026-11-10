-- name: CreateLeadSubmission :one
INSERT INTO lead_submissions (
    submission_type,
    email,
    first_name,
    full_name,
    agency_name,
    workflow_bottleneck,
    workflow_name,
    assessment_result,
    assessment_score,
    assessment_answers,
    client_volume,
    primary_vertical,
    delivery_constraint,
    consent_given,
    consent_recorded_date,
    utm_source,
    utm_medium,
    utm_campaign,
    utm_content,
    referrer,
    spam_status,
    spam_reasons,
    lead_status,
    turnstile_verified_at
) VALUES (
    sqlc.arg(submission_type),
    sqlc.arg(email),
    sqlc.arg(first_name),
    sqlc.arg(full_name),
    sqlc.arg(agency_name),
    sqlc.arg(workflow_bottleneck),
    sqlc.arg(workflow_name),
    sqlc.arg(assessment_result),
    sqlc.arg(assessment_score),
    sqlc.arg(assessment_answers),
    sqlc.arg(client_volume),
    sqlc.arg(primary_vertical),
    sqlc.arg(delivery_constraint),
    sqlc.arg(consent_given),
    CASE
        WHEN sqlc.arg(consent_given) = 1 THEN CURRENT_TIMESTAMP
        ELSE NULL
    END,
    sqlc.arg(utm_source),
    sqlc.arg(utm_medium),
    sqlc.arg(utm_campaign),
    sqlc.arg(utm_content),
    sqlc.arg(referrer),
    sqlc.arg(spam_status),
    sqlc.arg(spam_reasons),
    sqlc.arg(lead_status),
    CURRENT_TIMESTAMP
)
RETURNING *;

-- name: ClaimLeadSubmission :one
UPDATE lead_submissions
SET
    sync_status = 'processing',
    sync_attempts = sync_attempts + 1,
    last_attempt_date = CURRENT_TIMESTAMP,
    next_attempt_date = NULL
WHERE id = (
    SELECT id
    FROM lead_submissions
    WHERE
        sync_status = 'pending'
        OR (
            sync_status = 'failed'
            AND next_attempt_date IS NOT NULL
            AND next_attempt_date <= CURRENT_TIMESTAMP
        )
        OR (
            sync_status = 'processing'
            AND last_attempt_date <= datetime(CURRENT_TIMESTAMP, '-5 minutes')
        )
    ORDER BY submitted_date, id
    LIMIT 1
)
RETURNING *;

-- name: MarkLeadSubmissionSynced :one
UPDATE lead_submissions
SET
    sync_status = 'synced',
    next_attempt_date = NULL,
    last_sync_error = '',
    hubspot_contact_id = ?,
    synced_date = CURRENT_TIMESTAMP
WHERE id = ? AND sync_status = 'processing'
RETURNING id;

-- name: MarkLeadSubmissionFailed :one
UPDATE lead_submissions
SET
    sync_status = 'failed',
    next_attempt_date = sqlc.narg(next_attempt_date),
    last_sync_error = sqlc.arg(last_sync_error),
    synced_date = NULL
WHERE id = sqlc.arg(id) AND sync_status = 'processing'
RETURNING id;
