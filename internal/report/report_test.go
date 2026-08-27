package report

import (
	"github.com/German4341374/http-repro-lab/internal/model"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteUsesTextOnlyRendering(t *testing.T) {
	directory := t.TempDir()
	analysis := model.Analysis{Requests: []model.RequestSpec{{SchemaVersion: "1", Method: "GET", URL: model.URLSpec{Scheme: "https", Host: "example.invalid", Path: "/<script>alert(1)</script>", Query: []model.NameValue{}}, Headers: []model.NameValue{}, Body: model.Body{Type: "none"}, TimeoutMS: 1000}}}
	if err := Write(directory, Data{Analysis: analysis}); err != nil {
		t.Fatal(err)
	}
	app, err := os.ReadFile(filepath.Join(directory, "app.js"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(app), "textContent") || strings.Contains(string(app), "innerHTML") {
		t.Fatal("report renderer does not enforce text-only DOM writes")
	}
	if _, err := os.Stat(filepath.Join(directory, "index.html")); err != nil {
		t.Fatal(err)
	}
}
