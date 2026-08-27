package privacy

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/German4341374/http-repro-lab/internal/model"
)

var (
	secretNames       = regexp.MustCompile(`(?i)(authorization|cookie|set-cookie|x-api-key|api[-_]?key|client[-_]?secret|password|passwd|session[-_]?id|csrf|access[-_]?token|refresh[-_]?token|connection[-_]?string)`)
	bearerPattern     = regexp.MustCompile(`(?i)\bbearer\s+[A-Za-z0-9._~+/=-]{8,}`)
	jwtPattern        = regexp.MustCompile(`\beyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\b`)
	emailPattern      = regexp.MustCompile(`(?i)\b[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}\b`)
	phonePattern      = regexp.MustCompile(`(?:\+?[0-9][0-9 ()-]{7,}[0-9])`)
	privateKeyPattern = regexp.MustCompile(`-----BEGIN (?:RSA |EC |OPENSSH )?PRIVATE KEY-----`)
)

func preview(value string) string {
	runes := []rune(value)
	if len(runes) <= 8 {
		return "[redacted]"
	}
	return string(runes[:4]) + "..." + string(runes[len(runes)-3:])
}

func Detect(request model.RequestSpec, strict bool) []model.SensitiveValue {
	values := make([]model.SensitiveValue, 0)
	for _, header := range request.Headers {
		kind := ""
		if secretNames.MatchString(header.Name) {
			kind = classifyName(header.Name, header.Value)
		}
		if kind == "" && bearerPattern.MatchString(header.Value) {
			kind = "Bearer token"
		}
		if kind == "" && jwtPattern.MatchString(header.Value) {
			kind = "JWT"
		}
		if kind != "" {
			values = append(values, model.SensitiveValue{Type: kind, Location: "Header: " + header.Name, Confidence: "exact", Preview: preview(header.Value), SuggestedAction: "replace with " + placeholder(kind)})
		}
	}
	for _, query := range request.URL.Query {
		if secretNames.MatchString(query.Name) {
			kind := classifyName(query.Name, query.Value)
			values = append(values, model.SensitiveValue{Type: kind, Location: "Query: " + query.Name, Confidence: "strong", Preview: preview(query.Value), SuggestedAction: "replace with " + placeholder(kind)})
		}
	}
	body, _ := request.BodyBytes()
	text := string(body)
	if bearerPattern.MatchString(text) {
		values = append(values, model.SensitiveValue{Type: "Bearer token", Location: "Body", Confidence: "strong", Preview: "[redacted]", SuggestedAction: "replace with ${AUTH_TOKEN}"})
	}
	if privateKeyPattern.MatchString(text) {
		values = append(values, model.SensitiveValue{Type: "Private key", Location: "Body", Confidence: "exact", Preview: "[private key]", SuggestedAction: "remove the key and rotate it"})
	}
	if strict {
		for _, match := range emailPattern.FindAllString(text, -1) {
			values = append(values, model.SensitiveValue{Type: "Email", Location: "Body", Confidence: "strong", Preview: preview(match), SuggestedAction: "replace with a stable pseudonym"})
		}
		for _, match := range phonePattern.FindAllString(text, -1) {
			values = append(values, model.SensitiveValue{Type: "Phone", Location: "Body", Confidence: "heuristic", Preview: preview(match), SuggestedAction: "replace with a stable pseudonym"})
		}
	}
	sort.SliceStable(values, func(i, j int) bool { return values[i].Location+values[i].Type < values[j].Location+values[j].Type })
	return values
}

func classifyName(name, value string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.Contains(lower, "authorization") && strings.HasPrefix(strings.ToLower(value), "basic "):
		return "Basic credential"
	case strings.Contains(lower, "authorization"):
		return "Authorization"
	case strings.Contains(lower, "cookie"):
		return "Cookie"
	case strings.Contains(lower, "password") || strings.Contains(lower, "passwd"):
		return "Password"
	case strings.Contains(lower, "csrf"):
		return "CSRF token"
	case strings.Contains(lower, "session"):
		return "Session ID"
	case strings.Contains(lower, "refresh"):
		return "OAuth refresh token"
	case strings.Contains(lower, "token"):
		return "OAuth access token"
	case strings.Contains(lower, "secret"):
		return "Client secret"
	case strings.Contains(lower, "key"):
		return "API key"
	case strings.Contains(lower, "connection"):
		return "Connection string"
	default:
		return "Sensitive value"
	}
}

func placeholder(kind string) string {
	switch kind {
	case "Authorization", "Bearer token", "OAuth access token":
		return "${AUTH_TOKEN}"
	case "Basic credential":
		return "${BASIC_CREDENTIALS}"
	case "Cookie", "Session ID":
		return "${SESSION_VALUE}"
	case "Password":
		return "${PASSWORD}"
	case "API key":
		return "${API_KEY}"
	case "Client secret":
		return "${CLIENT_SECRET}"
	case "CSRF token":
		return "${CSRF_TOKEN}"
	default:
		return "${SECRET_VALUE}"
	}
}

func Sanitize(request model.RequestSpec, strict bool) (model.RequestSpec, []model.SensitiveValue, error) {
	result := request
	result.Headers = append([]model.NameValue(nil), request.Headers...)
	result.URL.Query = append([]model.NameValue(nil), request.URL.Query...)
	detections := Detect(request, strict)
	for index, header := range result.Headers {
		if secretNames.MatchString(header.Name) || bearerPattern.MatchString(header.Value) || jwtPattern.MatchString(header.Value) {
			result.Headers[index].Value = placeholder(classifyName(header.Name, header.Value))
		}
	}
	for index, query := range result.URL.Query {
		if secretNames.MatchString(query.Name) {
			result.URL.Query[index].Value = placeholder(classifyName(query.Name, query.Value))
		}
	}
	if request.Body.Content != nil {
		raw, err := json.Marshal(request.Body.Content)
		if err != nil {
			return model.RequestSpec{}, nil, fmt.Errorf("marshal body: %w", err)
		}
		text := bearerPattern.ReplaceAllString(string(raw), "${AUTH_TOKEN}")
		text = jwtPattern.ReplaceAllString(text, "${JWT}")
		if strict {
			text = stableReplace(text, emailPattern, "EMAIL")
			text = stableReplace(text, phonePattern, "PHONE")
		}
		var value any
		if err := json.Unmarshal([]byte(text), &value); err == nil {
			result.Body.Content = value
		} else {
			result.Body.Content = strings.Trim(text, `"`)
		}
	}
	if result.Provenance == nil {
		result.Provenance = map[string]string{}
	}
	result.Provenance["sanitizationProfile"] = map[bool]string{true: "strict", false: "standard"}[strict]
	return result, detections, nil
}

func stableReplace(input string, pattern *regexp.Regexp, prefix string) string {
	ids := map[string]string{}
	next := 1
	return pattern.ReplaceAllStringFunc(input, func(value string) string {
		if existing, ok := ids[value]; ok {
			return existing
		}
		replacement := fmt.Sprintf("%s_%03d", prefix, next)
		next++
		ids[value] = replacement
		return replacement
	})
}
