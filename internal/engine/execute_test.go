package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/German4341374/http-repro-lab/internal/model"
)

func testRequest(raw string) model.RequestSpec {
	u, _ := model.URLFromString(raw)
	return model.RequestSpec{SchemaVersion: "1", Method: "GET", URL: u, Headers: []model.NameValue{}, Body: model.Body{Type: "none"}, TimeoutMS: 2000}
}

func TestExecuteLocalWithExplicitPolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()
	response, err := Execute(context.Background(), testRequest(server.URL), "", Options{Policy: Policy{AllowPrivate: true}, MaxResponseCaptureBytes: 1024})
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != 200 || !strings.Contains(response.Body, "ok") || response.Timing.TotalMS <= 0 {
		t.Fatalf("unexpected response: %#v", response)
	}
}

func TestExecuteBlocksPrivateByDefault(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer server.Close()
	_, err := Execute(context.Background(), testRequest(server.URL), "", Options{})
	if err == nil || !strings.Contains(err.Error(), "TARGET_BLOCKED") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecuteRequiresAllowWrite(t *testing.T) {
	request := testRequest("https://example.invalid/items")
	request.Method = "POST"
	_, err := Execute(context.Background(), request, "", Options{})
	if err == nil || !strings.Contains(err.Error(), "allow-write") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResponseTruncation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("x", 100)))
	}))
	defer server.Close()
	response, err := Execute(context.Background(), testRequest(server.URL), "", Options{Policy: Policy{AllowPrivate: true}, MaxResponseCaptureBytes: 10})
	if err != nil {
		t.Fatal(err)
	}
	if !response.Truncated || len(response.Body) != 10 {
		t.Fatalf("not truncated: %#v", response)
	}
}
