package analyze

import (
	"fmt"
	"time"

	"github.com/German4341374/http-repro-lab/internal/importer"
	"github.com/German4341374/http-repro-lab/internal/model"
	"github.com/German4341374/http-repro-lab/internal/privacy"
)

func HAR(result importer.HARResult, strict bool) (model.Analysis, error) {
	analysis := model.Analysis{SourceSHA256: result.SourceSHA256, GeneratedAtUTC: time.Now().UTC(), Requests: make([]model.RequestSpec, 0, len(result.Requests))}
	for index, request := range result.Requests {
		sanitized, sensitive, err := privacy.Sanitize(request, strict)
		if err != nil {
			return model.Analysis{}, err
		}
		analysis.Requests = append(analysis.Requests, sanitized)
		analysis.Sensitive = append(analysis.Sensitive, sensitive...)
		status := result.Statuses[index]
		source := fmt.Sprintf("HAR entry %d", index)
		switch {
		case status == 401:
			analysis.Findings = append(analysis.Findings, finding(index, "HTTP_AUTH_401", "high", "authentication", "Authentication challenge observed", "The captured response returned HTTP 401.", source, "Response status", "401", "exact", "Inspect WWW-Authenticate and the credential forwarding policy.", "Verify token issuer, audience, expiry, and target environment."))
		case status == 403:
			analysis.Findings = append(analysis.Findings, finding(index, "HTTP_AUTH_403", "high", "authorization", "Authorization failure observed", "The captured response returned HTTP 403.", source, "Response status", "403", "exact", "Confirm the authenticated principal and required permissions."))
		case status == 429:
			analysis.Findings = append(analysis.Findings, finding(index, "HTTP_RATE_LIMIT", "medium", "reliability", "Rate limit evidence detected", "The captured response returned HTTP 429.", source, "Response status", "429", "exact", "Inspect Retry-After and provider rate-limit headers."))
		case status >= 500:
			analysis.Findings = append(analysis.Findings, finding(index, "HTTP_SERVER_ERROR", "high", "server", "Server error observed", fmt.Sprintf("The captured response returned HTTP %d.", status), source, "Response status", fmt.Sprintf("%d", status), "exact", "Correlate the request with sanitized server logs and request IDs."))
		}
		if result.DurationsMS[index] > 2000 {
			analysis.Findings = append(analysis.Findings, finding(index, "HTTP_SLOW", "medium", "performance", "Slow captured request", "The HAR reports a duration above two seconds.", source, "Total duration", fmt.Sprintf("%.1f ms", result.DurationsMS[index]), "exact", "Inspect the HAR timing waterfall and compare multiple samples."))
		}
		if len(sensitive) > 0 {
			analysis.Findings = append(analysis.Findings, finding(index, "SECRET_DETECTED", "high", "privacy", "Sensitive values require review", fmt.Sprintf("%d potentially sensitive values were detected and replaced in the sanitized request.", len(sensitive)), source, "Privacy detections", fmt.Sprintf("%d", len(sensitive)), "strong", "Review placeholders and rotate any credential that was shared outside its intended boundary."))
		}
	}
	return analysis, nil
}

func finding(index int, rule, severity, category, title, summary, source, field, value, confidence string, next ...string) model.Finding {
	return model.Finding{ID: fmt.Sprintf("finding-%03d-%s", index+1, rule), RuleID: rule, Severity: severity, Category: category, Title: title, Summary: summary, Evidence: []model.Evidence{{Source: source, Field: field, Value: value}}, Confidence: confidence, NextSteps: next}
}
