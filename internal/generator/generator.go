package generator

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/German4341374/http-repro-lab/internal/model"
)

type Output struct{ Language, FileName, Content string }

func All(request model.RequestSpec) ([]Output, error) {
	body, err := request.BodyBytes()
	if err != nil {
		return nil, err
	}
	url := request.URL.String()
	outputs := []Output{
		{Language: "curl", FileName: "request.sh", Content: curl(request, string(body), url)},
		{Language: "javascript", FileName: "request.mjs", Content: javascript(request, string(body), url, false)},
		{Language: "typescript", FileName: "request.ts", Content: javascript(request, string(body), url, true)},
		{Language: "python", FileName: "request.py", Content: python(request, string(body), url)},
		{Language: "go", FileName: "main.go", Content: goClient(request, string(body), url)},
		{Language: "java", FileName: "Main.java", Content: javaClient(request, string(body), url)},
		{Language: "csharp", FileName: "Program.cs", Content: csharp(request, string(body), url)},
		{Language: "php", FileName: "request.php", Content: php(request, string(body), url)},
	}
	return outputs, nil
}

func quoted(value string) string { raw, _ := json.Marshal(value); return string(raw) }

func safeHeaders(request model.RequestSpec) []model.NameValue {
	result := make([]model.NameValue, 0, len(request.Headers))
	for _, header := range request.Headers {
		if !strings.EqualFold(header.Name, "content-length") && !strings.EqualFold(header.Name, "host") {
			result = append(result, header)
		}
	}
	return result
}

func curl(request model.RequestSpec, body, url string) string {
	var builder strings.Builder
	builder.WriteString("#!/usr/bin/env sh\nset -eu\n\ncurl --fail-with-body --silent --show-error")
	builder.WriteString(" --max-time " + strconv.Itoa(max(1, request.TimeoutMS/1000)))
	builder.WriteString(" --request " + request.Method)
	for _, header := range safeHeaders(request) {
		builder.WriteString(" \\\n  --header " + shellQuote(header.Name+": "+header.Value))
	}
	if body != "" {
		builder.WriteString(" \\\n  --data-raw " + shellQuote(body))
	}
	builder.WriteString(" \\\n  " + shellQuote(url) + "\n")
	return builder.String()
}

func shellQuote(value string) string { return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'" }

func javascript(request model.RequestSpec, body, url string, typed bool) string {
	var builder strings.Builder
	if typed {
		builder.WriteString("interface ReproResult { status: number; headers: Record<string, string>; body: string }\n\n")
	}
	resultType := ""
	if typed {
		resultType = ": Promise<ReproResult>"
	}
	builder.WriteString("async function reproduce()" + resultType + " {\n")
	builder.WriteString("  const controller = new AbortController();\n  const timeout = setTimeout(() => controller.abort(), " + strconv.Itoa(request.TimeoutMS) + ");\n")
	builder.WriteString("  try {\n    const response = await fetch(" + quoted(url) + ", {\n      method: " + quoted(request.Method) + ",\n      headers: {")
	for _, header := range safeHeaders(request) {
		builder.WriteString("\n        " + quoted(header.Name) + ": " + quoted(header.Value) + ",")
	}
	builder.WriteString("\n      },")
	if body != "" {
		builder.WriteString("\n      body: " + quoted(body) + ",")
	}
	builder.WriteString("\n      redirect: \"manual\",\n      signal: controller.signal,\n    });\n    const text = await response.text();\n    return { status: response.status, headers: Object.fromEntries(response.headers), body: text };\n  } finally { clearTimeout(timeout); }\n}\n\n")
	builder.WriteString("reproduce().then((result) => { console.log(JSON.stringify(result)); }).catch((error) => { console.error(error); process.exitCode = 1; });\n")
	return builder.String()
}

func python(request model.RequestSpec, body, url string) string {
	var headers strings.Builder
	headers.WriteString("{")
	for index, header := range safeHeaders(request) {
		if index > 0 {
			headers.WriteString(", ")
		}
		headers.WriteString(quoted(header.Name) + ": " + quoted(header.Value))
	}
	headers.WriteString("}")
	bodyExpression := "None"
	if body != "" {
		bodyExpression = quoted(body) + ".encode(\"utf-8\")"
	}
	return "#!/usr/bin/env python3\nimport json\nimport urllib.request\n\nrequest = urllib.request.Request(" + quoted(url) + ", data=" + bodyExpression + ", headers=" + headers.String() + ", method=" + quoted(request.Method) + ")\ntry:\n    with urllib.request.urlopen(request, timeout=" + fmt.Sprintf("%.3f", float64(request.TimeoutMS)/1000) + ") as response:\n        print(json.dumps({\"status\": response.status, \"headers\": dict(response.headers), \"body\": response.read().decode(\"utf-8\", errors=\"replace\")}))\nexcept Exception as error:\n    raise SystemExit(f\"request failed: {error}\") from error\n"
}

func goClient(request model.RequestSpec, body, url string) string {
	reader := "nil"
	if body != "" {
		reader = "strings.NewReader(" + quoted(body) + ")"
	}
	var headers strings.Builder
	for _, header := range safeHeaders(request) {
		headers.WriteString("\treq.Header.Add(" + quoted(header.Name) + ", " + quoted(header.Value) + ")\n")
	}
	stringsImport := ""
	if body != "" {
		stringsImport = "\n\t\"strings\""
	}
	return "package main\n\nimport (\n\t\"context\"\n\t\"fmt\"\n\t\"io\"\n\t\"net/http\"" + stringsImport + "\n\t\"time\"\n)\n\nfunc main() {\n\tctx, cancel := context.WithTimeout(context.Background(), " + strconv.Itoa(request.TimeoutMS) + "*time.Millisecond)\n\tdefer cancel()\n\treq, err := http.NewRequestWithContext(ctx, " + quoted(request.Method) + ", " + quoted(url) + ", " + reader + ")\n\tif err != nil { panic(err) }\n" + headers.String() + "\tclient := &http.Client{CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}\n\tresp, err := client.Do(req)\n\tif err != nil { panic(err) }\n\tdefer resp.Body.Close()\n\tbody, err := io.ReadAll(resp.Body)\n\tif err != nil { panic(err) }\n\tfmt.Printf(\"status=%d body=%s\\n\", resp.StatusCode, body)\n}\n"
}

func javaClient(request model.RequestSpec, body, url string) string {
	methodBody := "noBody()"
	if body != "" {
		methodBody = "ofString(" + quoted(body) + ")"
	}
	var headers strings.Builder
	for _, header := range safeHeaders(request) {
		headers.WriteString("\n        .header(" + quoted(header.Name) + ", " + quoted(header.Value) + ")")
	}
	return "import java.net.URI;\nimport java.net.http.HttpClient;\nimport java.net.http.HttpRequest;\nimport java.net.http.HttpResponse;\nimport java.time.Duration;\n\npublic final class Main {\n  public static void main(String[] args) throws Exception {\n    HttpRequest request = HttpRequest.newBuilder(URI.create(" + quoted(url) + "))\n        .timeout(Duration.ofMillis(" + strconv.Itoa(request.TimeoutMS) + "))" + headers.String() + "\n        .method(" + quoted(request.Method) + ", HttpRequest.BodyPublishers." + methodBody + ")\n        .build();\n    HttpClient client = HttpClient.newBuilder().followRedirects(HttpClient.Redirect.NEVER).build();\n    HttpResponse<String> response = client.send(request, HttpResponse.BodyHandlers.ofString());\n    System.out.printf(\"status=%d body=%s%n\", response.statusCode(), response.body());\n  }\n}\n"
}

func csharp(request model.RequestSpec, body, url string) string {
	var headers strings.Builder
	for _, header := range safeHeaders(request) {
		headers.WriteString("request.Headers.TryAddWithoutValidation(" + quoted(header.Name) + ", " + quoted(header.Value) + ");\n")
	}
	content := ""
	if body != "" {
		content = "request.Content = new StringContent(" + quoted(body) + ", Encoding.UTF8);\n"
	}
	return "using System.Text;\n\nusing var cancellation = new CancellationTokenSource(TimeSpan.FromMilliseconds(" + strconv.Itoa(request.TimeoutMS) + "));\nusing var handler = new HttpClientHandler { AllowAutoRedirect = false };\nusing var client = new HttpClient(handler);\nusing var request = new HttpRequestMessage(new HttpMethod(" + quoted(request.Method) + "), " + quoted(url) + ");\n" + headers.String() + content + "using var response = await client.SendAsync(request, cancellation.Token);\nvar body = await response.Content.ReadAsStringAsync(cancellation.Token);\nConsole.WriteLine($\"status={(int)response.StatusCode} body={body}\");\n"
}

func php(request model.RequestSpec, body, url string) string {
	headers := make([]string, 0)
	for _, header := range safeHeaders(request) {
		headers = append(headers, quoted(header.Name+": "+header.Value))
	}
	post := ""
	if body != "" {
		post = "\ncurl_setopt($handle, CURLOPT_POSTFIELDS, " + quoted(body) + ");"
	}
	return "<?php\ndeclare(strict_types=1);\n\n$handle = curl_init(" + quoted(url) + ");\ncurl_setopt_array($handle, [\n    CURLOPT_CUSTOMREQUEST => " + quoted(request.Method) + ",\n    CURLOPT_HTTPHEADER => [" + strings.Join(headers, ", ") + "],\n    CURLOPT_RETURNTRANSFER => true,\n    CURLOPT_FOLLOWLOCATION => false,\n    CURLOPT_TIMEOUT_MS => " + strconv.Itoa(request.TimeoutMS) + ",\n    CURLOPT_SSL_VERIFYPEER => true,\n    CURLOPT_SSL_VERIFYHOST => 2,\n]);" + post + "\n$body = curl_exec($handle);\nif ($body === false) { throw new RuntimeException(curl_error($handle)); }\n$status = curl_getinfo($handle, CURLINFO_RESPONSE_CODE);\ncurl_close($handle);\nfwrite(STDOUT, json_encode([\"status\" => $status, \"body\" => $body], JSON_THROW_ON_ERROR) . PHP_EOL);\n"
}
