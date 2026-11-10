package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	sqlc "website.tabirianalytics.com/internal/database/sqlc"
)

const (
	hubspotEngagementProperty = "tabiri_engagement_types"
	hubspotResponseLimit      = 1 << 20
)

type hubspotClient struct {
	httpClient               *http.Client
	baseURL                  string
	serviceKey               string
	newsletterSubscriptionID string
}

type hubspotAPIError struct {
	statusCode int
	requestID  string
	retryAfter time.Duration
}

func (e *hubspotAPIError) Error() string {
	if e.requestID == "" {
		return fmt.Sprintf("HubSpot API returned status %d", e.statusCode)
	}
	return fmt.Sprintf("HubSpot API returned status %d (request ID %s)", e.statusCode, e.requestID)
}

type hubspotSearchResponse struct {
	Results []hubspotContact `json:"results"`
}

type hubspotContact struct {
	ID         string            `json:"id"`
	Properties map[string]string `json:"properties"`
}

type hubspotContactRequest struct {
	Properties map[string]string `json:"properties"`
}

type hubspotPreferenceRequest struct {
	EmailAddress          string `json:"emailAddress"`
	SubscriptionID        string `json:"subscriptionId"`
	LegalBasis            string `json:"legalBasis"`
	LegalBasisExplanation string `json:"legalBasisExplanation"`
}

func (c *hubspotClient) Sync(
	ctx context.Context,
	submission sqlc.LeadSubmission,
) (string, error) {
	email := strings.ToLower(strings.TrimSpace(submission.Email))
	existing, err := c.findContact(ctx, email)
	if err != nil {
		return "", err
	}

	properties := hubspotProperties(submission, existing)
	var contact hubspotContact
	if existing == nil {
		err = c.doJSON(ctx, http.MethodPost, "/crm/v3/objects/contacts", hubspotContactRequest{
			Properties: properties,
		}, &contact)
	} else {
		path := "/crm/v3/objects/contacts/" + url.PathEscape(existing.ID)
		err = c.doJSON(ctx, http.MethodPatch, path, hubspotContactRequest{
			Properties: properties,
		}, &contact)
	}
	if err != nil {
		return "", err
	}
	if contact.ID == "" && existing != nil {
		contact.ID = existing.ID
	}
	if contact.ID == "" {
		return "", fmt.Errorf("HubSpot contact response did not include an ID")
	}

	if submission.SubmissionType == "newsletter" && submission.ConsentGiven == 1 {
		err = c.doJSON(ctx, http.MethodPost, "/communication-preferences/v3/subscribe", hubspotPreferenceRequest{
			EmailAddress:          email,
			SubscriptionID:        c.newsletterSubscriptionID,
			LegalBasis:            "CONSENT_WITH_NOTICE",
			LegalBasisExplanation: "Explicit consent through the Tabiri Analytics newsletter form.",
		}, nil)
		if err != nil {
			return "", err
		}
	}

	return contact.ID, nil
}

func (c *hubspotClient) findContact(ctx context.Context, email string) (*hubspotContact, error) {
	request := map[string]any{
		"filterGroups": []any{
			map[string]any{
				"filters": []any{
					map[string]string{
						"propertyName": "email",
						"operator":     "EQ",
						"value":        email,
					},
				},
			},
		},
		"properties": []string{hubspotEngagementProperty},
		"limit":      1,
	}
	var response hubspotSearchResponse
	if err := c.doJSON(
		ctx,
		http.MethodPost,
		"/crm/v3/objects/contacts/search",
		request,
		&response,
	); err != nil {
		return nil, err
	}
	if len(response.Results) == 0 {
		return nil, nil
	}
	return &response.Results[0], nil
}

func (c *hubspotClient) doJSON(
	ctx context.Context,
	method string,
	path string,
	body any,
	result any,
) error {
	var requestBody io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode HubSpot request: %w", err)
		}
		requestBody = bytes.NewReader(encoded)
	}

	request, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, requestBody)
	if err != nil {
		return fmt.Errorf("create HubSpot request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+c.serviceKey)
	request.Header.Set("Content-Type", "application/json")
	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("send HubSpot request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, hubspotResponseLimit))
		return &hubspotAPIError{
			statusCode: response.StatusCode,
			requestID:  response.Header.Get("X-HubSpot-Correlation-Id"),
			retryAfter: parseRetryAfter(response.Header.Get("Retry-After"), time.Now()),
		}
	}
	if result == nil || response.StatusCode == http.StatusNoContent {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, hubspotResponseLimit))
		return nil
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, hubspotResponseLimit)).Decode(result); err != nil {
		return fmt.Errorf("decode HubSpot response: %w", err)
	}
	return nil
}

func hubspotProperties(
	submission sqlc.LeadSubmission,
	existing *hubspotContact,
) map[string]string {
	properties := map[string]string{
		"email":                       strings.ToLower(strings.TrimSpace(submission.Email)),
		hubspotEngagementProperty:     mergedEngagements(existing, submission.SubmissionType),
		"tabiri_latest_submission_at": hubspotDate(submission.SubmittedDate),
		"tabiri_latest_lead_source":   hubspotLeadSource(submission),
	}
	switch submission.SubmissionType {
	case "commercial_lead":
		firstName, lastName := splitName(submission.FullName)
		addProperty(properties, "firstname", firstName)
		addProperty(properties, "lastname", lastName)
		addProperty(properties, "tabiri_workflow_bottleneck", submission.WorkflowBottleneck)
		addProperty(properties, "tabiri_workflow_name", submission.WorkflowName)
		addProperty(properties, "tabiri_assessment_result", submission.AssessmentResult)
		addProperty(properties, "tabiri_assessment_score", submission.AssessmentScore)
	case "partner":
		addProperty(properties, "company", submission.AgencyName)
		addProperty(properties, "tabiri_partner_client_volume", submission.ClientVolume)
		addProperty(properties, "tabiri_partner_primary_vertical", submission.PrimaryVertical)
		addProperty(properties, "tabiri_partner_delivery_constraint", submission.DeliveryConstraint)
	case "newsletter":
		addProperty(properties, "firstname", submission.FirstName)
	}
	return properties
}

func mergedEngagements(existing *hubspotContact, submissionType string) string {
	values := make(map[string]struct{})
	if existing != nil {
		for _, value := range strings.Split(existing.Properties[hubspotEngagementProperty], ";") {
			if value = strings.TrimSpace(value); value != "" {
				values[value] = struct{}{}
			}
		}
	}
	values[submissionType] = struct{}{}

	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return strings.Join(result, ";")
}

func hubspotLeadSource(submission sqlc.LeadSubmission) string {
	switch submission.SubmissionType {
	case "partner":
		return "Partner form"
	case "newsletter":
		return "Newsletter signup"
	default:
		if submission.AssessmentResult != "" || submission.AssessmentScore != "" {
			return "Automation assessment"
		}
		return "Contact form"
	}
}

func hubspotDate(value string) string {
	for _, layout := range []string{
		time.RFC3339Nano,
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
	} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return strconv.FormatInt(parsed.UTC().UnixMilli(), 10)
		}
	}
	return value
}

func splitName(value string) (string, string) {
	parts := strings.Fields(value)
	if len(parts) == 0 {
		return "", ""
	}
	return parts[0], strings.Join(parts[1:], " ")
}

func addProperty(properties map[string]string, name, value string) {
	if value = strings.TrimSpace(value); value != "" {
		properties[name] = value
	}
}

func parseRetryAfter(value string, now time.Time) time.Duration {
	if seconds, err := strconv.Atoi(strings.TrimSpace(value)); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	if retryAt, err := http.ParseTime(value); err == nil && retryAt.After(now) {
		return retryAt.Sub(now)
	}
	return 0
}
