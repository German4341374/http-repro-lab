package importer

import "testing"

func TestParseCURL(t *testing.T) {
	request, err := ParseCURL(`curl -X POST -H "Content-Type: application/json" --data-raw '{"name":"demo"}' https://example.invalid/items`)
	if err != nil {
		t.Fatal(err)
	}
	if request.Method != "POST" || request.URL.Path != "/items" || len(request.Headers) != 1 {
		t.Fatalf("unexpected request: %#v", request)
	}
}

func TestParseCURLRejectsCommandSubstitution(t *testing.T) {
	for _, input := range []string{`curl "$(cat token)"`, `curl https://example.invalid && rm -rf /`, `curl ` + "`whoami`"} {
		if _, err := ParseCURL(input); err == nil {
			t.Fatalf("accepted unsafe input %q", input)
		}
	}
}

func TestParseCURLRejectsUnknownFlag(t *testing.T) {
	if _, err := ParseCURL(`curl --config file https://example.invalid`); err == nil {
		t.Fatal("expected rejection")
	}
}
