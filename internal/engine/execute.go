package engine

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"strings"
	"time"

	"github.com/German4341374/http-repro-lab/internal/model"
)

type Options struct {
	Policy                  Policy
	MaxResponseCaptureBytes int64
	MaxRedirects            int
}

func Execute(ctx context.Context, request model.RequestSpec, target string, options Options) (model.Response, error) {
	if err := request.Validate(); err != nil {
		return model.Response{}, fmt.Errorf("INPUT_INVALID: %w", err)
	}
	if err := options.Policy.ValidateMethod(request.Method, request.URL.Path); err != nil {
		return model.Response{}, err
	}
	if target == "" {
		target = request.URL.String()
	} else {
		base := strings.TrimRight(target, "/")
		query := ""
		if index := strings.Index(request.URL.String(), "?"); index >= 0 {
			query = request.URL.String()[index:]
		}
		target = base + request.URL.Path + query
	}
	resolved, err := options.Policy.ValidateURL(ctx, target)
	if err != nil {
		return model.Response{}, err
	}
	body, err := request.BodyBytes()
	if err != nil {
		return model.Response{}, fmt.Errorf("INPUT_INVALID: encode body: %w", err)
	}
	timeout := time.Duration(request.TimeoutMS) * time.Millisecond
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	httpRequest, err := http.NewRequestWithContext(ctx, request.Method, target, bytes.NewReader(body))
	if err != nil {
		return model.Response{}, fmt.Errorf("INPUT_INVALID: %w", err)
	}
	for _, header := range request.Headers {
		if strings.Contains(header.Value, "${") {
			continue
		}
		httpRequest.Header.Add(header.Name, header.Value)
	}

	var dnsStart, connectStart, tlsStart, wroteAt, firstByte time.Time
	var timing model.Timing
	trace := &httptrace.ClientTrace{
		DNSStart: func(httptrace.DNSStartInfo) { dnsStart = time.Now() },
		DNSDone: func(httptrace.DNSDoneInfo) {
			if !dnsStart.IsZero() {
				timing.DNSMS = milliseconds(time.Since(dnsStart))
			}
		},
		ConnectStart: func(_, _ string) { connectStart = time.Now() },
		ConnectDone: func(_, _ string, _ error) {
			if !connectStart.IsZero() {
				timing.ConnectMS = milliseconds(time.Since(connectStart))
			}
		},
		TLSHandshakeStart: func() { tlsStart = time.Now() },
		TLSHandshakeDone: func(tls.ConnectionState, error) {
			if !tlsStart.IsZero() {
				timing.TLSMS = milliseconds(time.Since(tlsStart))
			}
		},
		WroteRequest: func(httptrace.WroteRequestInfo) { wroteAt = time.Now() },
		GotFirstResponseByte: func() {
			firstByte = time.Now()
			if !wroteAt.IsZero() {
				timing.TTFBMS = milliseconds(firstByte.Sub(wroteAt))
			}
		},
	}
	httpRequest = httpRequest.WithContext(httptrace.WithClientTrace(httpRequest.Context(), trace))
	redirects := make([]model.RedirectStep, 0)
	maxRedirects := options.MaxRedirects
	if maxRedirects <= 0 {
		maxRedirects = 10
	}
	transport := &http.Transport{Proxy: http.ProxyFromEnvironment, ForceAttemptHTTP2: true, TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12}}
	client := &http.Client{Transport: transport, Timeout: timeout}
	client.CheckRedirect = func(next *http.Request, via []*http.Request) error {
		if len(via) >= maxRedirects {
			return fmt.Errorf("redirect limit exceeded")
		}
		if _, policyErr := options.Policy.ValidateURL(next.Context(), next.URL.String()); policyErr != nil {
			return policyErr
		}
		if len(via) > 0 && !strings.EqualFold(via[len(via)-1].URL.Hostname(), next.URL.Hostname()) {
			next.Header.Del("Authorization")
			next.Header.Del("Cookie")
		}
		redirects = append(redirects, model.RedirectStep{Status: 0, Method: next.Method, URL: next.URL.String()})
		return nil
	}
	started := time.Now()
	response, err := client.Do(httpRequest)
	if err != nil {
		if ctx.Err() != nil {
			return model.Response{}, fmt.Errorf("HTTP_TIMEOUT: %w", ctx.Err())
		}
		return model.Response{}, fmt.Errorf("network request failed: %w", err)
	}
	defer response.Body.Close()
	limit := options.MaxResponseCaptureBytes
	if limit <= 0 {
		limit = 10 << 20
	}
	captured, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil {
		return model.Response{}, fmt.Errorf("read response: %w", err)
	}
	total := time.Since(started)
	timing.TotalMS = milliseconds(total)
	if !firstByte.IsZero() {
		timing.DownloadMS = milliseconds(time.Since(firstByte))
	}
	truncated := int64(len(captured)) > limit
	if truncated {
		captured = captured[:limit]
	}
	sum := sha256.Sum256(captured)
	headers := make(map[string]string, len(response.Header))
	for name, values := range response.Header {
		headers[name] = strings.Join(values, ", ")
	}
	result := model.Response{StatusCode: response.StatusCode, Headers: headers, BodySHA256: hex.EncodeToString(sum[:]), SizeBytes: response.ContentLength, Truncated: truncated, Timing: timing, Redirects: redirects, ResolvedIP: resolved}
	if response.ContentLength < 0 {
		result.SizeBytes = int64(len(captured))
	}
	contentType := strings.ToLower(response.Header.Get("Content-Type"))
	if strings.HasPrefix(contentType, "text/") || strings.Contains(contentType, "json") || strings.Contains(contentType, "xml") || contentType == "" {
		result.Body = string(captured)
	}
	if response.TLS != nil {
		result.TLS = tlsMetadata(response.TLS)
	}
	return result, nil
}

func tlsMetadata(state *tls.ConnectionState) *model.TLSMetadata {
	metadata := &model.TLSMetadata{Version: tlsVersion(state.Version), Cipher: tls.CipherSuiteName(state.CipherSuite), HostnameValidation: len(state.VerifiedChains) > 0}
	if len(state.PeerCertificates) > 0 {
		certificate := state.PeerCertificates[0]
		metadata.Subject = certificate.Subject.String()
		metadata.Issuer = certificate.Issuer.String()
		metadata.DNSNames = certificate.DNSNames
		metadata.ValidFrom = certificate.NotBefore
		metadata.ValidUntil = certificate.NotAfter
	}
	if len(state.VerifiedChains) > 0 {
		metadata.ChainLength = len(state.VerifiedChains[0])
	}
	return metadata
}

func tlsVersion(version uint16) string {
	switch version {
	case tls.VersionTLS13:
		return "TLS 1.3"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS10:
		return "TLS 1.0"
	default:
		return "unknown"
	}
}

func milliseconds(duration time.Duration) float64 { return float64(duration.Microseconds()) / 1000 }
