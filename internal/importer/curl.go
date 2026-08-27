package importer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/German4341374/http-repro-lab/internal/model"
)

func ParseCURL(input string) (model.RequestSpec, error) {
	if strings.Contains(input, "$(") || strings.Contains(input, "`") || strings.Contains(input, "&&") || strings.Contains(input, "||") || strings.Contains(input, ";") {
		return model.RequestSpec{}, fmt.Errorf("CURL_UNSUPPORTED_SYNTAX: shell expansion or command chaining detected")
	}
	tokens, err := tokenize(input)
	if err != nil {
		return model.RequestSpec{}, fmt.Errorf("CURL_UNSUPPORTED_SYNTAX: %w", err)
	}
	if len(tokens) == 0 || (tokens[0] != "curl" && tokens[0] != "curl.exe") {
		return model.RequestSpec{}, fmt.Errorf("input must start with curl")
	}
	request := model.RequestSpec{SchemaVersion: model.SchemaVersion, Method: "GET", Body: model.Body{Type: "none"}, TimeoutMS: 30000, Provenance: map[string]string{"source": "cURL", "normalization": "v1"}}
	var rawURL string
	for i := 1; i < len(tokens); i++ {
		token := tokens[i]
		next := func() (string, error) {
			i++
			if i >= len(tokens) {
				return "", fmt.Errorf("missing value after %s", token)
			}
			return tokens[i], nil
		}
		switch token {
		case "-X", "--request":
			value, nextErr := next()
			if nextErr != nil {
				return model.RequestSpec{}, nextErr
			}
			request.Method = strings.ToUpper(value)
		case "-H", "--header":
			value, nextErr := next()
			if nextErr != nil {
				return model.RequestSpec{}, nextErr
			}
			parts := strings.SplitN(value, ":", 2)
			if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" {
				return model.RequestSpec{}, fmt.Errorf("invalid header %q", value)
			}
			request.Headers = append(request.Headers, model.NameValue{Name: strings.TrimSpace(parts[0]), Value: strings.TrimSpace(parts[1])})
		case "-d", "--data", "--data-raw", "--data-binary":
			value, nextErr := next()
			if nextErr != nil {
				return model.RequestSpec{}, nextErr
			}
			request.Body = model.Body{Type: "text", Content: value}
			if request.Method == "GET" {
				request.Method = "POST"
			}
		case "-u", "--user":
			value, nextErr := next()
			if nextErr != nil {
				return model.RequestSpec{}, nextErr
			}
			request.Headers = append(request.Headers, model.NameValue{Name: "Authorization", Value: "Basic " + value})
		case "-b", "--cookie":
			value, nextErr := next()
			if nextErr != nil {
				return model.RequestSpec{}, nextErr
			}
			request.Headers = append(request.Headers, model.NameValue{Name: "Cookie", Value: value})
		case "--max-time":
			value, nextErr := next()
			if nextErr != nil {
				return model.RequestSpec{}, nextErr
			}
			seconds, parseErr := strconv.ParseFloat(value, 64)
			if parseErr != nil || seconds <= 0 {
				return model.RequestSpec{}, fmt.Errorf("invalid max time")
			}
			request.TimeoutMS = int(seconds * 1000)
		case "--compressed", "--location", "-L":
			// Behavior is represented by the execution policy, not shell flags.
		default:
			if strings.HasPrefix(token, "-") {
				return model.RequestSpec{}, fmt.Errorf("CURL_UNSUPPORTED_SYNTAX: unsupported flag %s", token)
			}
			if rawURL != "" {
				return model.RequestSpec{}, fmt.Errorf("multiple URLs are not supported")
			}
			rawURL = token
		}
	}
	if rawURL == "" {
		return model.RequestSpec{}, fmt.Errorf("cURL command has no URL")
	}
	request.URL, err = model.URLFromString(rawURL)
	if err != nil {
		return model.RequestSpec{}, err
	}
	return request, request.Validate()
}

func tokenize(input string) ([]string, error) {
	var tokens []string
	var current strings.Builder
	var quote rune
	escaped := false
	flush := func() {
		if current.Len() > 0 {
			tokens = append(tokens, current.String())
			current.Reset()
		}
	}
	for _, char := range input {
		if escaped {
			current.WriteRune(char)
			escaped = false
			continue
		}
		if char == '\\' && quote != '\'' {
			escaped = true
			continue
		}
		if quote != 0 {
			if char == quote {
				quote = 0
			} else {
				current.WriteRune(char)
			}
			continue
		}
		switch char {
		case '\'', '"':
			quote = char
		case ' ', '\t', '\r', '\n':
			flush()
		default:
			current.WriteRune(char)
		}
	}
	if escaped || quote != 0 {
		return nil, fmt.Errorf("unterminated escape or quote")
	}
	flush()
	return tokens, nil
}
