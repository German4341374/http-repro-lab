package privacy

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/German4341374/http-repro-lab/internal/model"
)

func sensitiveRequest() model.RequestSpec {
	return model.RequestSpec{SchemaVersion: "1", Method: "POST", URL: model.URLSpec{Scheme: "https", Host: "example.invalid", Path: "/", Query: []model.NameValue{{Name: "api_key", Value: "synthetic-key-value"}}}, Headers: []model.NameValue{{Name: "Authorization", Value: "Bearer synthetic-token-value"}, {Name: "Cookie", Value: "session=synthetic"}}, Body: model.Body{Type: "json", Content: map[string]any{"email": "test.user@example.invalid", "note": "hello"}}, TimeoutMS: 1000}
}

func TestSanitizeReplacesSecrets(t *testing.T) {
	result, detections, err := Sanitize(sensitiveRequest(), true)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(result)
	text := string(raw)
	for _, forbidden := range []string{"synthetic-token-value", "synthetic-key-value", "test.user@example.invalid"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("secret remained: %s", forbidden)
		}
	}
	if len(detections) < 4 {
		t.Fatalf("expected detections, got %d", len(detections))
	}
}
func TestSanitizeIsIdempotent(t *testing.T) {
	first, _, err := Sanitize(sensitiveRequest(), true)
	if err != nil {
		t.Fatal(err)
	}
	second, _, err := Sanitize(first, true)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("sanitization changed on second pass\n%#v\n%#v", first, second)
	}
}
func TestStablePseudonyms(t *testing.T) {
	request := sensitiveRequest()
	request.Body.Content = map[string]any{"a": "same@example.invalid", "b": "same@example.invalid", "c": "other@example.invalid"}
	result, _, err := Sanitize(request, true)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(result.Body.Content)
	text := string(raw)
	if strings.Count(text, "EMAIL_001") != 2 || strings.Count(text, "EMAIL_002") != 1 {
		t.Fatalf("unexpected pseudonyms: %s", text)
	}
}
