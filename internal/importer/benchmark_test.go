package importer

import (
	"bytes"
	"fmt"
	"testing"
)

func syntheticHAR(entries int) []byte {
	var buffer bytes.Buffer
	buffer.Grow(entries * 260)
	buffer.WriteString(`{"log":{"entries":[`)
	for index := 0; index < entries; index++ {
		if index > 0 {
			buffer.WriteByte(',')
		}
		fmt.Fprintf(&buffer, `{"time":%d,"request":{"method":"GET","url":"https://api.example.invalid/items/%d?q=synthetic","headers":[{"name":"Accept","value":"application/json"}],"queryString":[{"name":"q","value":"synthetic"}]},"response":{"status":200,"headers":[]}}`, index%3000, index)
	}
	buffer.WriteString(`]}}`)
	return buffer.Bytes()
}

func benchmarkParseHAR(b *testing.B, entries int) {
	raw := syntheticHAR(entries)
	b.ReportMetric(float64(len(raw)), "input-bytes")
	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		result, err := ParseHAR(bytes.NewReader(raw))
		if err != nil {
			b.Fatal(err)
		}
		if len(result.Requests) != entries {
			b.Fatalf("got %d entries", len(result.Requests))
		}
	}
}

func BenchmarkParseHAR10k(b *testing.B)  { benchmarkParseHAR(b, 10_000) }
func BenchmarkParseHAR50k(b *testing.B)  { benchmarkParseHAR(b, 50_000) }
func BenchmarkParseHAR100k(b *testing.B) { benchmarkParseHAR(b, 100_000) }
