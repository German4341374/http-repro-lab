package compare

import (
	"github.com/German4341374/http-repro-lab/internal/model"
	"testing"
)

func TestIdenticalDiffIsEmpty(t *testing.T) {
	response := model.Response{StatusCode: 200, Headers: map[string]string{"Content-Type": "application/json"}, Body: `{"id":10}`, SizeBytes: 9}
	result := Responses("a", "b", response, response)
	if len(result.Differences) != 0 {
		t.Fatalf("unexpected differences: %#v", result.Differences)
	}
}
func TestStructuralTypeDifference(t *testing.T) {
	a := model.Response{StatusCode: 200, Headers: map[string]string{}, Body: `{"id":10}`}
	b := model.Response{StatusCode: 200, Headers: map[string]string{}, Body: `{"id":"10"}`}
	result := Responses("a", "b", a, b)
	found := false
	for _, difference := range result.Differences {
		if difference.Field == "json.$.id" {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing structural diff: %#v", result.Differences)
	}
}
