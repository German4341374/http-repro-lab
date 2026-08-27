package dev.httprepro;

import static org.junit.jupiter.api.Assertions.*;
import com.sun.net.httpserver.HttpServer;
import java.net.InetSocketAddress;
import java.net.URI;
import java.net.http.HttpClient;
import java.nio.charset.StandardCharsets;
import java.time.Duration;
import java.util.Map;
import org.junit.jupiter.api.Test;

final class ReproRunnerTest {
    @Test void executesAgainstLocalServer() throws Exception {
        var server = HttpServer.create(new InetSocketAddress("127.0.0.1", 0), 0);
        server.createContext("/echo", exchange -> { var body = exchange.getRequestMethod().getBytes(StandardCharsets.UTF_8); exchange.sendResponseHeaders(200, body.length); exchange.getResponseBody().write(body); exchange.close(); });
        server.start();
        try {
            var result = new ReproRunner(HttpClient.newBuilder().followRedirects(HttpClient.Redirect.NEVER).build()).execute("POST", URI.create("http://127.0.0.1:" + server.getAddress().getPort() + "/echo"), Map.of("X-Demo", "synthetic"), "{}", Duration.ofSeconds(2));
            assertEquals(200, result.statusCode()); assertEquals("POST", result.body());
        } finally { server.stop(0); }
    }
    @Test void rejectsUnsupportedScheme() { assertThrows(IllegalArgumentException.class, () -> new ReproRunner(HttpClient.newHttpClient()).execute("GET", URI.create("ftp://example.invalid/file"), Map.of(), null, Duration.ofSeconds(1))); }
}

