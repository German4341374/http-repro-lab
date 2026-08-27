package generator

import (
	"github.com/German4341374/http-repro-lab/internal/model"
	"strings"
	"testing"
)

func TestAllGeneratorsAreProduced(t *testing.T) {
	request := model.RequestSpec{SchemaVersion: "1", Method: "POST", URL: model.URLSpec{Scheme: "http", Host: "127.0.0.1", Port: 9090, Path: "/echo", Query: []model.NameValue{{Name: "q", Value: "demo"}}}, Headers: []model.NameValue{{Name: "Content-Type", Value: "application/json"}}, Body: model.Body{Type: "json", Content: map[string]any{"name": "demo"}}, TimeoutMS: 3000}
	outputs, err := All(request)
	if err != nil {
		t.Fatal(err)
	}
	if len(outputs) != 8 {
		t.Fatalf("expected 8, got %d", len(outputs))
	}
	for _, output := range outputs {
		if output.Content == "" || !strings.Contains(output.Content, "127.0.0.1") {
			t.Fatalf("invalid %s output", output.Language)
		}
	}
}
func TestGeneratorsDoNotEmitHostHeader(t *testing.T) {
	request := model.RequestSpec{SchemaVersion: "1", Method: "GET", URL: model.URLSpec{Scheme: "https", Host: "example.invalid", Path: "/", Query: []model.NameValue{}}, Headers: []model.NameValue{{Name: "Host", Value: "evil.invalid"}}, Body: model.Body{Type: "none"}, TimeoutMS: 1000}
	outputs, err := All(request)
	if err != nil {
		t.Fatal(err)
	}
	for _, output := range outputs {
		if strings.Contains(output.Content, "evil.invalid") {
			t.Fatalf("unsafe Host header in %s", output.Language)
		}
	}
}
