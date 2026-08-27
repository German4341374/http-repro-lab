package model

import "testing"

func TestURLRoundTripPreservesQueryOrder(t *testing.T) {
	value, err := URLFromString("https://Example.INVALID:8443/api?q=first&q=second&space=a+b")
	if err != nil {
		t.Fatal(err)
	}
	if value.Host != "example.invalid" || value.Port != 8443 || len(value.Query) != 3 {
		t.Fatalf("unexpected URL: %#v", value)
	}
	if value.Query[0].Value != "first" || value.Query[1].Value != "second" {
		t.Fatalf("query order changed: %#v", value.Query)
	}
}

func TestURLRejectsUserInfo(t *testing.T) {
	if _, err := URLFromString("https://user:password@example.invalid/"); err == nil {
		t.Fatal("expected userinfo rejection")
	}
}

func TestValidateIsIdempotent(t *testing.T) {
	request := RequestSpec{SchemaVersion: "1", Method: "GET", URL: URLSpec{Scheme: "https", Host: "example.invalid", Path: "/", Query: []NameValue{}}, Headers: []NameValue{}, Body: Body{Type: "none"}, TimeoutMS: 1000}
	if err := request.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := request.Validate(); err != nil {
		t.Fatal(err)
	}
}
