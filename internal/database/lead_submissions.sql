CREATE TABLE IF NOT EXISTS lead_submissions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    submission_type TEXT NOT NULL
        CHECK (submission_type IN ('commercial_lead', 'partner', 'newsletter')),
    email TEXT NOT NULL,
    first_name TEXT NOT NULL DEFAULT '',
    full_name TEXT NOT NULL DEFAULT '',
    agency_name TEXT NOT NULL DEFAULT '',
    workflow_bottleneck TEXT NOT NULL DEFAULT '',
    workflow_name TEXT NOT NULL DEFAULT '',
    assessment_result TEXT NOT NULL DEFAULT '',
    assessment_score TEXT NOT NULL DEFAULT '',
    assessment_answers TEXT NOT NULL DEFAULT '',
    client_volume TEXT NOT NULL DEFAULT ''
        CHECK (client_volume IN ('', '1-10', '11-50', '50+')),
    primary_vertical TEXT NOT NULL DEFAULT '',
    delivery_constraint TEXT NOT NULL DEFAULT '',
    consent_given INTEGER NOT NULL DEFAULT 0
        CHECK (consent_given IN (0, 1)),
    consent_recorded_date TEXT,
    utm_source TEXT NOT NULL DEFAULT '',
    utm_medium TEXT NOT NULL DEFAULT '',
    utm_campaign TEXT NOT NULL DEFAULT '',
    utm_content TEXT NOT NULL DEFAULT '',
    referrer TEXT NOT NULL DEFAULT '',
    spam_status TEXT NOT NULL DEFAULT 'review'
        CHECK (spam_status IN ('accepted', 'review', 'rejected')),
    spam_reasons TEXT NOT NULL DEFAULT '[]',
    lead_status TEXT NOT NULL DEFAULT 'new'
        CHECK (lead_status IN ('new', 'contacted', 'qualified', 'disqualified', 'archived')),
    turnstile_verified_at TEXT NOT NULL DEFAULT '',
    sync_status TEXT NOT NULL DEFAULT 'pending'
        CHECK (sync_status IN ('pending', 'processing', 'synced', 'failed')),
    sync_attempts INTEGER NOT NULL DEFAULT 0 CHECK (sync_attempts >= 0),
    next_attempt_date TEXT,
    last_attempt_date TEXT,
    last_sync_error TEXT NOT NULL DEFAULT '',
    hubspot_contact_id TEXT NOT NULL DEFAULT '',
    synced_date TEXT,
    submitted_date TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_lead_submissions_sync_queue
    ON lead_submissions (sync_status, next_attempt_date, submitted_date);
