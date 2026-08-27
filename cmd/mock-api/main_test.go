package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEnvironmentVariantsExposeMeaningfulDifferences(t *testing.T) {
	staging := httptest.NewRecorder()
	routes("staging").ServeHTTP(staging, httptest.NewRequest(http.MethodGet, "/environment", nil))
	production := httptest.NewRecorder()
	routes("production").ServeHTTP(production, httptest.NewRequest(http.MethodGet, "/environment", nil))
	if staging.Header().Get("Content-Type") != "application/json" || production.Header().Get("Content-Type") != "text/html" {
		t.Fatal("content types did not differ")
	}
	if !strings.Contains(staging.Body.String(), `"id":10`) || !strings.Contains(production.Body.String(), `"id":"10"`) {
		t.Fatal("JSON types did not differ")
	}
}
