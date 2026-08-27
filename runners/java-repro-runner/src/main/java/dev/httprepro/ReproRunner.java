package dev.httprepro;

import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;
import java.util.Map;

public final class ReproRunner {
    private final HttpClient client;
    public ReproRunner(HttpClient client) { this.client = client; }

    public Result execute(String method, URI uri, Map<String, String> headers, String body, Duration timeout) throws Exception {
        if (!(uri.getScheme().equals("http") || uri.getScheme().equals("https"))) throw new IllegalArgumentException("Only HTTP and HTTPS URLs are supported");
        var builder = HttpRequest.newBuilder(uri).timeout(timeout);
        headers.entrySet().stream().filter(entry -> !entry.getKey().equalsIgnoreCase("Host") && !entry.getKey().equalsIgnoreCase("Content-Length")).sorted(Map.Entry.comparingByKey()).forEach(entry -> builder.header(entry.getKey(), entry.getValue()));
        var publisher = body == null ? HttpRequest.BodyPublishers.noBody() : HttpRequest.BodyPublishers.ofString(body);
        var started = System.nanoTime();
        var response = client.send(builder.method(method, publisher).build(), HttpResponse.BodyHandlers.ofString());
        return new Result(response.statusCode(), response.body(), Duration.ofNanos(System.nanoTime() - started));
    }

    public record Result(int statusCode, String body, Duration duration) {}
}

