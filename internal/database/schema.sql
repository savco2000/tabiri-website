CREATE TABLE IF NOT EXISTS contact_submissions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    full_name TEXT NOT NULL,
    corporate_email TEXT NOT NULL,
    workflow_bottleneck TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS partner_contact_submissions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    agency_name TEXT NOT NULL,
    corporate_email TEXT NOT NULL,
    client_volume TEXT NOT NULL CHECK (client_volume IN ('1-10', '11-50', '50+')),
    primary_vertical TEXT NOT NULL,
    delivery_constraint TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
