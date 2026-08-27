package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/German4341374/http-repro-lab/internal/analyze"
	"github.com/German4341374/http-repro-lab/internal/compare"
	"github.com/German4341374/http-repro-lab/internal/engine"
	"github.com/German4341374/http-repro-lab/internal/generator"
	"github.com/German4341374/http-repro-lab/internal/importer"
	"github.com/German4341374/http-repro-lab/internal/model"
	"github.com/German4341374/http-repro-lab/internal/pack"
	"github.com/German4341374/http-repro-lab/internal/privacy"
	"github.com/German4341374/http-repro-lab/internal/report"
)

const version = "0.1.0"

type codedError struct {
	code int
	err  error
}

func (e codedError) Error() string { return e.err.Error() }
func (e codedError) Unwrap() error { return e.err }

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		var coded codedError
		if errors.As(err, &coded) {
			os.Exit(coded.code)
		}
		os.Exit(5)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		usage(stdout)
		return codedError{2, fmt.Errorf("command is required")}
	}
	switch args[0] {
	case "analyze":
		return analyzeCommand(args[1:], stdout, stderr)
	case "sanitize":
		return sanitizeCommand(args[1:], stdout, stderr)
	case "reproduce":
		return reproduceCommand(args[1:], stdout, stderr)
	case "compare", "diff":
		return compareCommand(args[1:], stdout, stderr)
	case "generate":
		return generateCommand(args[1:], stdout, stderr)
	case "pack":
		return packCommand(args[1:], stdout, stderr)
	case "import":
		return importCommand(args[1:], stdout, stderr)
	case "validate":
		return validateCommand(args[1:], stdout, stderr)
	case "har":
		if len(args) > 1 && args[1] == "summary" {
			return summaryCommand(args[2:], stdout, stderr)
		}
		return codedError{2, fmt.Errorf("expected har summary")}
	case "version":
		fmt.Fprintf(stdout, "http-repro %s\n", version)
		return nil
	case "help", "--help", "-h":
		usage(stdout)
		return nil
	default:
		return codedError{2, fmt.Errorf("unknown command %q", args[0])}
	}
}

func usage(writer io.Writer) {
	fmt.Fprintln(writer, `HTTP Repro Lab - privacy-first HTTP reproduction

Usage:
  http-repro analyze <capture.har> [--output report] [--strict] [--json]
  http-repro sanitize <capture.har> --request 1 --output request.json
  http-repro reproduce <capture.har|request.json> --request 1 --target URL [--allow-private] [--allow-write]
  http-repro compare <capture.har|request.json> --request 1 --target-a URL --target-b URL
  http-repro generate <capture.har|request.json> --request 1 --output generated
  http-repro pack <capture.har|request.json> --request 1 --output incident.repro.zip
  http-repro import --curl "curl ..." --output request.json
  http-repro har summary <capture.har>
  http-repro validate <request.json>
  http-repro version`)
}

func analyzeCommand(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("analyze", flag.ContinueOnError)
	fs.SetOutput(stderr)
	output := fs.String("output", "report", "report directory")
	strict := fs.Bool("strict", false, "enable PII detection")
	jsonOutput := fs.Bool("json", false, "write JSON to stdout")
	if err := fs.Parse(interspersed(args, map[string]bool{"--output": true})); err != nil {
		return codedError{2, err}
	}
	if fs.NArg() != 1 {
		return codedError{2, fmt.Errorf("analyze requires one HAR file")}
	}
	result, err := readHAR(fs.Arg(0))
	if err != nil {
		return codedError{2, err}
	}
	analysis, err := analyze.HAR(result, *strict)
	if err != nil {
		return err
	}
	generated := map[string]string{}
	if len(analysis.Requests) > 0 {
		outputs, genErr := generator.All(analysis.Requests[0])
		if genErr != nil {
			return genErr
		}
		for _, item := range outputs {
			generated[item.Language] = item.Content
		}
	}
	if err := report.Write(*output, report.Data{Analysis: analysis, Generated: generated}); err != nil {
		return err
	}
	if *jsonOutput {
		raw, _ := json.MarshalIndent(analysis, "", "  ")
		fmt.Fprintln(stdout, string(raw))
		return nil
	}
	fmt.Fprintf(stdout, "%d HTTP requests analyzed\n%d findings\n%d potentially sensitive values detected\nOffline report: %s\n", len(analysis.Requests), len(analysis.Findings), len(analysis.Sensitive), filepath.Join(*output, "index.html"))
	return nil
}

func sanitizeCommand(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("sanitize", flag.ContinueOnError)
	fs.SetOutput(stderr)
	index := fs.Int("request", 1, "one-based request index")
	output := fs.String("output", "request.json", "output path")
	strict := fs.Bool("strict", false, "enable PII detection")
	if err := fs.Parse(interspersed(args, map[string]bool{"--request": true, "--output": true})); err != nil {
		return codedError{2, err}
	}
	if fs.NArg() != 1 {
		return codedError{2, fmt.Errorf("sanitize requires an input file")}
	}
	request, err := loadRequest(fs.Arg(0), *index)
	if err != nil {
		return codedError{2, err}
	}
	sanitized, detections, err := privacy.Sanitize(request, *strict)
	if err != nil {
		return err
	}
	if err := writeJSON(*output, sanitized); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Sanitized request written to %s (%d detections)\n", *output, len(detections))
	return nil
}

func reproduceCommand(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("reproduce", flag.ContinueOnError)
	fs.SetOutput(stderr)
	index := fs.Int("request", 1, "one-based request index")
	target := fs.String("target", "", "target base URL")
	allowPrivate := fs.Bool("allow-private", false, "allow private targets for local testing")
	allowWrite := fs.Bool("allow-write", false, "allow mutating methods")
	output := fs.String("output", "", "response JSON")
	if err := fs.Parse(interspersed(args, map[string]bool{"--request": true, "--target": true, "--output": true})); err != nil {
		return codedError{2, err}
	}
	if fs.NArg() != 1 {
		return codedError{2, fmt.Errorf("reproduce requires an input file")}
	}
	request, err := loadRequest(fs.Arg(0), *index)
	if err != nil {
		return codedError{2, err}
	}
	sanitized, _, err := privacy.Sanitize(request, false)
	if err != nil {
		return err
	}
	response, err := engine.Execute(context.Background(), sanitized, *target, engine.Options{Policy: engine.Policy{AllowPrivate: *allowPrivate, AllowWrite: *allowWrite}, MaxResponseCaptureBytes: 10 << 20})
	if err != nil {
		if strings.Contains(err.Error(), "blocked by safety policy") || strings.Contains(err.Error(), "TARGET_BLOCKED") {
			return codedError{3, err}
		}
		return codedError{4, err}
	}
	if *output != "" {
		if err := writeJSON(*output, response); err != nil {
			return err
		}
	}
	raw, _ := json.MarshalIndent(response, "", "  ")
	fmt.Fprintln(stdout, string(raw))
	return nil
}

func compareCommand(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("compare", flag.ContinueOnError)
	fs.SetOutput(stderr)
	index := fs.Int("request", 1, "one-based request index")
	targetA := fs.String("target-a", "", "first target")
	targetB := fs.String("target-b", "", "second target")
	allowPrivate := fs.Bool("allow-private", false, "allow private targets for local testing")
	output := fs.String("output", "comparison.json", "output path")
	if err := fs.Parse(interspersed(args, map[string]bool{"--request": true, "--target-a": true, "--target-b": true, "--output": true})); err != nil {
		return codedError{2, err}
	}
	if fs.NArg() != 1 || *targetA == "" || *targetB == "" {
		return codedError{2, fmt.Errorf("compare requires input, --target-a, and --target-b")}
	}
	request, err := loadRequest(fs.Arg(0), *index)
	if err != nil {
		return codedError{2, err}
	}
	request, _, err = privacy.Sanitize(request, false)
	if err != nil {
		return err
	}
	options := engine.Options{Policy: engine.Policy{AllowPrivate: *allowPrivate}, MaxResponseCaptureBytes: 10 << 20}
	a, err := engine.Execute(context.Background(), request, *targetA, options)
	if err != nil {
		return codedError{4, fmt.Errorf("target A: %w", err)}
	}
	b, err := engine.Execute(context.Background(), request, *targetB, options)
	if err != nil {
		return codedError{4, fmt.Errorf("target B: %w", err)}
	}
	result := compare.Responses(*targetA, *targetB, a, b)
	if err := writeJSON(*output, result); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "%d observed differences written to %s\n", len(result.Differences), *output)
	if len(result.Differences) > 0 {
		return codedError{1, fmt.Errorf("reproduction mismatch")}
	}
	return nil
}

func generateCommand(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("generate", flag.ContinueOnError)
	fs.SetOutput(stderr)
	index := fs.Int("request", 1, "one-based request index")
	output := fs.String("output", "generated", "output directory")
	language := fs.String("language", "all", "language or all")
	if err := fs.Parse(interspersed(args, map[string]bool{"--request": true, "--output": true, "--language": true})); err != nil {
		return codedError{2, err}
	}
	if fs.NArg() != 1 {
		return codedError{2, fmt.Errorf("generate requires input")}
	}
	request, err := loadRequest(fs.Arg(0), *index)
	if err != nil {
		return codedError{2, err}
	}
	request, _, err = privacy.Sanitize(request, false)
	if err != nil {
		return err
	}
	outputs, err := generator.All(request)
	if err != nil {
		return err
	}
	count := 0
	for _, item := range outputs {
		if *language != "all" && item.Language != *language {
			continue
		}
		dir := filepath.Join(*output, item.Language)
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, item.FileName), []byte(item.Content), 0o640); err != nil {
			return err
		}
		count++
	}
	if count == 0 {
		return codedError{2, fmt.Errorf("unsupported language %q", *language)}
	}
	fmt.Fprintf(stdout, "Generated %d clients in %s\n", count, *output)
	return nil
}

func packCommand(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("pack", flag.ContinueOnError)
	fs.SetOutput(stderr)
	index := fs.Int("request", 1, "one-based request index")
	output := fs.String("output", "incident.repro.zip", "output archive")
	if err := fs.Parse(interspersed(args, map[string]bool{"--request": true, "--output": true})); err != nil {
		return codedError{2, err}
	}
	if fs.NArg() != 1 {
		return codedError{2, fmt.Errorf("pack requires input")}
	}
	request, err := loadRequest(fs.Arg(0), *index)
	if err != nil {
		return codedError{2, err}
	}
	request, _, err = privacy.Sanitize(request, true)
	if err != nil {
		return err
	}
	outputs, err := generator.All(request)
	if err != nil {
		return err
	}
	if err := pack.Write(*output, request, outputs); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Sanitized reproduction pack written to %s\n", *output)
	return nil
}

func importCommand(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("import", flag.ContinueOnError)
	fs.SetOutput(stderr)
	curlInput := fs.String("curl", "", "cURL command treated as data")
	output := fs.String("output", "request.json", "output path")
	if err := fs.Parse(args); err != nil {
		return codedError{2, err}
	}
	if *curlInput == "" {
		return codedError{2, fmt.Errorf("--curl is required")}
	}
	request, err := importer.ParseCURL(*curlInput)
	if err != nil {
		return codedError{2, err}
	}
	request, _, err = privacy.Sanitize(request, false)
	if err != nil {
		return err
	}
	if err := writeJSON(*output, request); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Normalized request written to %s\n", *output)
	return nil
}

func validateCommand(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	fs.SetOutput(stderr)
	if err := fs.Parse(args); err != nil {
		return codedError{2, err}
	}
	if fs.NArg() != 1 {
		return codedError{2, fmt.Errorf("validate requires request.json")}
	}
	request, err := loadRequest(fs.Arg(0), 1)
	if err != nil {
		return codedError{2, err}
	}
	if err := request.Validate(); err != nil {
		return codedError{2, err}
	}
	fmt.Fprintln(stdout, "RequestSpec is valid")
	return nil
}

func summaryCommand(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("summary", flag.ContinueOnError)
	fs.SetOutput(stderr)
	if err := fs.Parse(args); err != nil {
		return codedError{2, err}
	}
	if fs.NArg() != 1 {
		return codedError{2, fmt.Errorf("summary requires HAR")}
	}
	result, err := readHAR(fs.Arg(0))
	if err != nil {
		return codedError{2, err}
	}
	domains := map[string]bool{}
	failed, redirects, clientErrors, serverErrors, slow := 0, 0, 0, 0, 0
	for i, r := range result.Requests {
		domains[r.URL.Host] = true
		s := result.Statuses[i]
		if s >= 400 {
			failed++
		}
		if s >= 300 && s < 400 {
			redirects++
		}
		if s >= 400 && s < 500 {
			clientErrors++
		}
		if s >= 500 {
			serverErrors++
		}
		if result.DurationsMS[i] > 2000 {
			slow++
		}
	}
	sensitive := 0
	for _, r := range result.Requests {
		sensitive += len(privacy.Detect(r, false))
	}
	fmt.Fprintf(stdout, "Requests: %d\nDomains: %d\nFailed: %d\nRedirects: %d\n4xx: %d\n5xx: %d\nSlow > 2s: %d\nSensitive values: %d\n", len(result.Requests), len(domains), failed, redirects, clientErrors, serverErrors, slow, sensitive)
	return nil
}

func readHAR(path string) (importer.HARResult, error) {
	file, err := os.Open(path)
	if err != nil {
		return importer.HARResult{}, err
	}
	defer file.Close()
	return importer.ParseHAR(file)
}
func loadRequest(path string, index int) (model.RequestSpec, error) {
	if index < 1 {
		return model.RequestSpec{}, fmt.Errorf("request index starts at 1")
	}
	if strings.HasSuffix(strings.ToLower(path), ".har") {
		result, err := readHAR(path)
		if err != nil {
			return model.RequestSpec{}, err
		}
		if index > len(result.Requests) {
			return model.RequestSpec{}, fmt.Errorf("request %d not found", index)
		}
		return result.Requests[index-1], nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return model.RequestSpec{}, err
	}
	var request model.RequestSpec
	if err := json.Unmarshal(raw, &request); err != nil {
		return model.RequestSpec{}, err
	}
	return request, request.Validate()
}
func writeJSON(path string, value any) error {
	if directory := filepath.Dir(path); directory != "." {
		if err := os.MkdirAll(directory, 0o750); err != nil {
			return err
		}
	}
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0o640)
}

func interspersed(args []string, valued map[string]bool) []string {
	flags := []string{}
	positionals := []string{}
	for index := 0; index < len(args); index++ {
		arg := args[index]
		if strings.HasPrefix(arg, "--") {
			flags = append(flags, arg)
			name := strings.SplitN(arg, "=", 2)[0]
			if valued[name] && !strings.Contains(arg, "=") {
				if index+1 < len(args) {
					index++
					flags = append(flags, args[index])
				}
			}
		} else {
			positionals = append(positionals, arg)
		}
	}
	return append(flags, positionals...)
}

func atoi(value string) int { number, _ := strconv.Atoi(value); return number }
