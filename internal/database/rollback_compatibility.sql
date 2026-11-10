CREATE TRIGGER IF NOT EXISTS mirror_legacy_commercial_lead
AFTER INSERT ON contact_submissions
BEGIN
    INSERT INTO lead_submissions (
        submission_type,
        email,
        full_name,
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
        turnstile_verified_at,
        submitted_date
    ) VALUES (
        'commercial_lead',
        NEW.corporate_email,
        NEW.full_name,
        NEW.workflow_bottleneck,
        NEW.workflow_name,
        NEW.assessment_result,
        NEW.assessment_score,
        NEW.assessment_answers,
        NEW.utm_source,
        NEW.utm_medium,
        NEW.utm_campaign,
        NEW.utm_content,
        NEW.referrer,
        NEW.spam_status,
        NEW.spam_reasons,
        NEW.lead_status,
        NEW.turnstile_verified_at,
        NEW.created_date
    );
END;

CREATE TRIGGER IF NOT EXISTS mirror_legacy_partner_lead
AFTER INSERT ON partner_contact_submissions
BEGIN
    INSERT INTO lead_submissions (
        submission_type,
        email,
        agency_name,
        client_volume,
        primary_vertical,
        delivery_constraint,
        spam_status,
        spam_reasons,
        lead_status,
        turnstile_verified_at,
        submitted_date
    ) VALUES (
        'partner',
        NEW.corporate_email,
        NEW.agency_name,
        NEW.client_volume,
        NEW.primary_vertical,
        NEW.delivery_constraint,
        NEW.spam_status,
        NEW.spam_reasons,
        NEW.lead_status,
        NEW.turnstile_verified_at,
        NEW.created_date
    );
END;

CREATE TRIGGER IF NOT EXISTS mirror_legacy_newsletter
AFTER INSERT ON newsletter_subscriptions
BEGIN
    INSERT INTO lead_submissions (
        submission_type,
        email,
        first_name,
        consent_given,
        consent_recorded_date,
        spam_status,
        spam_reasons,
        lead_status,
        turnstile_verified_at,
        submitted_date
    ) VALUES (
        'newsletter',
        NEW.email,
        NEW.first_name,
        NEW.consent_given,
        NEW.created_date,
        NEW.spam_status,
        NEW.spam_reasons,
        NEW.lead_status,
        NEW.turnstile_verified_at,
        NEW.created_date
    );
END;
