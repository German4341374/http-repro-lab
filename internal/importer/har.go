package importer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/German4341374/http-repro-lab/internal/model"
)

const MaxHARBytes = 100 << 20

type harDocument struct {
	Log struct {
		Entries []harEntry `json:"entries"`
	} `json:"log"`
}

type harEntry struct {
	Request struct {
		Method      string            `json:"method"`
		URL         string            `json:"url"`
		Headers     []model.NameValue `json:"headers"`
		QueryString []model.NameValue `json:"queryString"`
		PostData    *struct {
			MimeType string `json:"mimeType"`
			Text     string `json:"text"`
		} `json:"postData"`
	} `json:"request"`
	Response struct {
		Status  int               `json:"status"`
		Headers []model.NameValue `json:"headers"`
	} `json:"response"`
	Time float64 `json:"time"`
}

type HARResult struct {
	SourceSHA256 string
	Requests     []model.RequestSpec
	Statuses     []int
	DurationsMS  []float64
}

func ParseHAR(reader io.Reader) (HARResult, error) {
	limited := io.LimitReader(reader, MaxHARBytes+1)
	raw, err := io.ReadAll(limited)
	if err != nil {
		return HARResult{}, fmt.Errorf("read HAR: %w", err)
	}
	if len(raw) > MaxHARBytes {
		return HARResult{}, fmt.Errorf("HAR exceeds %d byte limit", MaxHARBytes)
	}
	var document harDocument
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.UseNumber()
	if err := decoder.Decode(&document); err != nil {
		return HARResult{}, fmt.Errorf("HAR_MALFORMED: %w", err)
	}
	if len(document.Log.Entries) == 0 {
		return HARResult{}, fmt.Errorf("HAR contains no entries")
	}
	result := HARResult{Requests: make([]model.RequestSpec, 0, len(document.Log.Entries))}
	sum := sha256.Sum256(raw)
	result.SourceSHA256 = hex.EncodeToString(sum[:])
	for index, entry := range document.Log.Entries {
		urlSpec, parseErr := model.URLFromString(entry.Request.URL)
		if parseErr != nil {
			return HARResult{}, fmt.Errorf("entry %d URL: %w", index, parseErr)
		}
		if len(entry.Request.QueryString) > 0 {
			urlSpec.Query = append([]model.NameValue(nil), entry.Request.QueryString...)
		}
		body := model.Body{Type: "none"}
		if entry.Request.PostData != nil {
			body.Type = "text"
			body.Content = entry.Request.PostData.Text
			if strings.Contains(strings.ToLower(entry.Request.PostData.MimeType), "json") {
				var value any
				if json.Unmarshal([]byte(entry.Request.PostData.Text), &value) == nil {
					body.Type = "json"
					body.Content = value
				}
			}
		}
		request := model.RequestSpec{
			SchemaVersion: model.SchemaVersion,
			Method:        strings.ToUpper(entry.Request.Method),
			URL:           urlSpec,
			Headers:       append([]model.NameValue(nil), entry.Request.Headers...),
			Body:          body,
			TimeoutMS:     30000,
			Provenance: map[string]string{
				"source":           fmt.Sprintf("HAR entry %d", index),
				"sourceSha256":     result.SourceSHA256,
				"normalization":    "v1",
				"originalStatus":   fmt.Sprintf("%d", entry.Response.Status),
				"originalDuration": fmt.Sprintf("%.3fms", entry.Time),
			},
		}
		if request.Method == "" {
			request.Method = http.MethodGet
		}
		if err := request.Validate(); err != nil {
			return HARResult{}, fmt.Errorf("entry %d: %w", index, err)
		}
		result.Requests = append(result.Requests, request)
		result.Statuses = append(result.Statuses, entry.Response.Status)
		result.DurationsMS = append(result.DurationsMS, entry.Time)
	}
	return result, nil
}
