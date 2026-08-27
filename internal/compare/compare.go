package compare

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/German4341374/http-repro-lab/internal/model"
)

func Responses(targetA, targetB string, a, b model.Response) model.Comparison {
	comparison := model.Comparison{TargetA: targetA, TargetB: targetB, ResponseA: a, ResponseB: b}
	add := func(field string, left, right any, interpretation, verify string) {
		if reflect.DeepEqual(left, right) {
			return
		}
		comparison.Differences = append(comparison.Differences, model.Difference{Field: field, EnvironmentA: left, EnvironmentB: right, Evidence: fmt.Sprintf("Observed %s differs between the two captured responses", field), PossibleInterpretation: interpretation, SuggestedVerification: verify})
	}
	add("status", a.StatusCode, b.StatusCode, "Routing, authentication, authorization, or application behavior may differ.", "Inspect server-side request IDs and authentication policy.")
	for _, name := range []string{"Content-Type", "Location", "WWW-Authenticate", "Retry-After", "Cache-Control", "ETag", "Strict-Transport-Security", "Content-Security-Policy"} {
		add("header."+strings.ToLower(name), header(a.Headers, name), header(b.Headers, name), "The selected response header is configured differently.", "Compare proxy and application configuration for this header.")
	}
	add("responseSize", a.SizeBytes, b.SizeBytes, "The response representations may differ.", "Compare sanitized response bodies or their schemas.")
	if a.TLS != nil && b.TLS != nil {
		add("tls.version", a.TLS.Version, b.TLS.Version, "TLS termination or policy may differ.", "Inspect the load balancer and certificate configuration.")
		add("tls.issuer", a.TLS.Issuer, b.TLS.Issuer, "The environments present certificates from different issuers.", "Verify the intended certificate chain for each host.")
	}
	var jsonA, jsonB any
	if json.Unmarshal([]byte(a.Body), &jsonA) == nil && json.Unmarshal([]byte(b.Body), &jsonB) == nil {
		structuralDiff("$", jsonA, jsonB, &comparison.Differences)
	}
	sort.SliceStable(comparison.Differences, func(i, j int) bool { return comparison.Differences[i].Field < comparison.Differences[j].Field })
	return comparison
}

func header(headers map[string]string, name string) string {
	for key, value := range headers {
		if strings.EqualFold(key, name) {
			return value
		}
	}
	return ""
}

func structuralDiff(path string, left, right any, differences *[]model.Difference) {
	leftType, rightType := jsonType(left), jsonType(right)
	if leftType != rightType {
		*differences = append(*differences, model.Difference{Field: "json." + path, EnvironmentA: leftType, EnvironmentB: rightType, Evidence: "Response JSON types differ at " + path, PossibleInterpretation: "The response contract may differ between environments.", SuggestedVerification: "Compare deployed API schemas and serializers."})
		return
	}
	lm, lok := left.(map[string]any)
	rm, rok := right.(map[string]any)
	if lok && rok {
		keys := map[string]bool{}
		for key := range lm {
			keys[key] = true
		}
		for key := range rm {
			keys[key] = true
		}
		ordered := make([]string, 0, len(keys))
		for key := range keys {
			ordered = append(ordered, key)
		}
		sort.Strings(ordered)
		for _, key := range ordered {
			structuralDiff(path+"."+key, lm[key], rm[key], differences)
		}
	}
}

func jsonType(value any) string {
	if value == nil {
		return "missing-or-null"
	}
	switch value.(type) {
	case string:
		return "string"
	case float64:
		return "number"
	case bool:
		return "boolean"
	case []any:
		return "array"
	case map[string]any:
		return "object"
	default:
		return fmt.Sprintf("%T", value)
	}
}
