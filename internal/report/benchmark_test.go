package report

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/German4341374/http-repro-lab/internal/model"
)

func benchmarkReport(b *testing.B, entries int) {
	requests := make([]model.RequestSpec, entries)
	for index := range requests {
		requests[index] = model.RequestSpec{SchemaVersion: "1", Method: "GET", URL: model.URLSpec{Scheme: "https", Host: "api.example.invalid", Path: fmt.Sprintf("/items/%d", index), Query: []model.NameValue{{Name: "q", Value: "synthetic"}}}, Headers: []model.NameValue{{Name: "Accept", Value: "application/json"}}, Body: model.Body{Type: "none"}, TimeoutMS: 30000}
	}
	data := Data{Analysis: model.Analysis{SourceSHA256: "synthetic", Requests: requests, GeneratedAtUTC: time.Unix(0, 0).UTC()}}
	directory := b.TempDir()
	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		if err := Write(filepath.Join(directory, fmt.Sprintf("run-%d", iteration)), data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReport10k(b *testing.B)  { benchmarkReport(b, 10_000) }
func BenchmarkReport50k(b *testing.B)  { benchmarkReport(b, 50_000) }
func BenchmarkReport100k(b *testing.B) { benchmarkReport(b, 100_000) }
