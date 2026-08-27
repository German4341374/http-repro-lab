package model

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const SchemaVersion = "1"

type NameValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type URLSpec struct {
	Scheme string      `json:"scheme"`
	Host   string      `json:"host"`
	Port   int         `json:"port,omitempty"`
	Path   string      `json:"path"`
	Query  []NameValue `json:"query"`
}

func URLFromString(raw string) (URLSpec, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return URLSpec{}, fmt.Errorf("parse URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return URLSpec{}, fmt.Errorf("unsupported URL scheme %q", u.Scheme)
	}
	if u.Hostname() == "" || u.User != nil {
		return URLSpec{}, fmt.Errorf("URL must include a host and must not contain userinfo")
	}
	port := 0
	if u.Port() != "" {
		port, err = strconv.Atoi(u.Port())
		if err != nil || port < 1 || port > 65535 {
			return URLSpec{}, fmt.Errorf("invalid URL port")
		}
	}
	path := u.EscapedPath()
	if path == "" {
		path = "/"
	}
	query := make([]NameValue, 0)
	for _, pair := range strings.Split(u.RawQuery, "&") {
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, "=", 2)
		name, decodeErr := url.QueryUnescape(parts[0])
		if decodeErr != nil {
			return URLSpec{}, fmt.Errorf("decode query name: %w", decodeErr)
		}
		value := ""
		if len(parts) == 2 {
			value, decodeErr = url.QueryUnescape(parts[1])
			if decodeErr != nil {
				return URLSpec{}, fmt.Errorf("decode query value: %w", decodeErr)
			}
		}
		query = append(query, NameValue{Name: name, Value: value})
	}
	return URLSpec{Scheme: strings.ToLower(u.Scheme), Host: strings.ToLower(u.Hostname()), Port: port, Path: path, Query: query}, nil
}

func (u URLSpec) String() string {
	host := u.Host
	if strings.Contains(host, ":") {
		host = "[" + host + "]"
	}
	if u.Port != 0 && !((u.Scheme == "http" && u.Port == 80) || (u.Scheme == "https" && u.Port == 443)) {
		host = net.JoinHostPort(u.Host, strconv.Itoa(u.Port))
	}
	values := make([]string, 0, len(u.Query))
	for _, item := range u.Query {
		values = append(values, url.QueryEscape(item.Name)+"="+url.QueryEscape(item.Value))
	}
	result := u.Scheme + "://" + host + u.Path
	if len(values) > 0 {
		result += "?" + strings.Join(values, "&")
	}
	return result
}

type Body struct {
	Type    string `json:"type"`
	Content any    `json:"content,omitempty"`
}

type RequestSpec struct {
	SchemaVersion string            `json:"schemaVersion"`
	Method        string            `json:"method"`
	URL           URLSpec           `json:"url"`
	Headers       []NameValue       `json:"headers"`
	Body          Body              `json:"body"`
	TimeoutMS     int               `json:"timeoutMs"`
	Provenance    map[string]string `json:"provenance,omitempty"`
}

func (r RequestSpec) Validate() error {
	if r.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported schema version %q", r.SchemaVersion)
	}
	if r.Method == "" || strings.ToUpper(r.Method) != r.Method {
		return fmt.Errorf("method must be uppercase")
	}
	if _, err := URLFromString(r.URL.String()); err != nil {
		return err
	}
	if r.TimeoutMS < 1 || r.TimeoutMS > 300000 {
		return fmt.Errorf("timeoutMs must be between 1 and 300000")
	}
	validBody := map[string]bool{"none": true, "text": true, "json": true, "form": true, "multipart": true, "binary-metadata": true}
	if !validBody[r.Body.Type] {
		return fmt.Errorf("unsupported body type %q", r.Body.Type)
	}
	return nil
}

func (r RequestSpec) BodyBytes() ([]byte, error) {
	if r.Body.Content == nil || r.Body.Type == "none" {
		return nil, nil
	}
	if text, ok := r.Body.Content.(string); ok {
		return []byte(text), nil
	}
	return json.Marshal(r.Body.Content)
}

type SensitiveValue struct {
	Type            string `json:"type"`
	Location        string `json:"location"`
	Confidence      string `json:"confidence"`
	Preview         string `json:"preview"`
	SuggestedAction string `json:"suggestedAction"`
}

type Evidence struct {
	Source string `json:"source"`
	Field  string `json:"field"`
	Value  string `json:"value"`
}

type Finding struct {
	ID         string     `json:"id"`
	RuleID     string     `json:"ruleId"`
	Severity   string     `json:"severity"`
	Category   string     `json:"category"`
	Title      string     `json:"title"`
	Summary    string     `json:"summary"`
	Evidence   []Evidence `json:"evidence"`
	Confidence string     `json:"confidence"`
	NextSteps  []string   `json:"nextSteps"`
}

type Timing struct {
	DNSMS      float64 `json:"dnsMs,omitempty"`
	ConnectMS  float64 `json:"connectMs,omitempty"`
	TLSMS      float64 `json:"tlsMs,omitempty"`
	TTFBMS     float64 `json:"ttfbMs,omitempty"`
	DownloadMS float64 `json:"downloadMs,omitempty"`
	TotalMS    float64 `json:"totalMs"`
}

type TLSMetadata struct {
	Version            string    `json:"version,omitempty"`
	Cipher             string    `json:"cipher,omitempty"`
	Subject            string    `json:"subject,omitempty"`
	Issuer             string    `json:"issuer,omitempty"`
	DNSNames           []string  `json:"dnsNames,omitempty"`
	ValidFrom          time.Time `json:"validFrom,omitempty"`
	ValidUntil         time.Time `json:"validUntil,omitempty"`
	ChainLength        int       `json:"chainLength,omitempty"`
	HostnameValidation bool      `json:"hostnameValidation,omitempty"`
}

type RedirectStep struct {
	Status int    `json:"status"`
	Method string `json:"method"`
	URL    string `json:"url"`
}

type Response struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body,omitempty"`
	BodySHA256 string            `json:"bodySha256"`
	SizeBytes  int64             `json:"sizeBytes"`
	Truncated  bool              `json:"truncated"`
	Timing     Timing            `json:"timing"`
	TLS        *TLSMetadata      `json:"tls,omitempty"`
	Redirects  []RedirectStep    `json:"redirects"`
	ResolvedIP []string          `json:"resolvedIp"`
}

type Analysis struct {
	SourceSHA256   string           `json:"sourceSha256"`
	Requests       []RequestSpec    `json:"requests"`
	Sensitive      []SensitiveValue `json:"sensitiveValues"`
	Findings       []Finding        `json:"findings"`
	GeneratedAtUTC time.Time        `json:"generatedAtUtc"`
}

type Difference struct {
	Field                  string `json:"field"`
	EnvironmentA           any    `json:"environmentA"`
	EnvironmentB           any    `json:"environmentB"`
	Evidence               string `json:"evidence"`
	PossibleInterpretation string `json:"possibleInterpretation"`
	SuggestedVerification  string `json:"suggestedVerification"`
}

type Comparison struct {
	TargetA     string       `json:"targetA"`
	TargetB     string       `json:"targetB"`
	ResponseA   Response     `json:"responseA"`
	ResponseB   Response     `json:"responseB"`
	Differences []Difference `json:"differences"`
}
