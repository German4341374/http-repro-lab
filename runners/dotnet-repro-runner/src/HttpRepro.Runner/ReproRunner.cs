using System.Net.Security;
using System.Security.Cryptography.X509Certificates;
using System.Text;

namespace HttpRepro.Runner;

public sealed record ReproRequest(
    HttpMethod Method,
    Uri Uri,
    IReadOnlyDictionary<string, string> Headers,
    string? Body,
    TimeSpan Timeout);

public sealed record ReproResult(int StatusCode, string Body, TimeSpan Duration);

public static class ReproRunner
{
    public static async Task<ReproResult> ExecuteAsync(
        ReproRequest request,
        HttpMessageHandler? handler = null,
        CancellationToken cancellationToken = default)
    {
        ArgumentNullException.ThrowIfNull(request);
        if (request.Uri.Scheme is not ("http" or "https"))
        {
            throw new ArgumentException("Only HTTP and HTTPS URLs are supported.", nameof(request));
        }

        handler ??= new SocketsHttpHandler
        {
            AllowAutoRedirect = false,
            SslOptions = new SslClientAuthenticationOptions
            {
                CertificateRevocationCheckMode = X509RevocationMode.Online,
            },
        };
        using var client = new HttpClient(handler, disposeHandler: true) { Timeout = request.Timeout };
        using var message = new HttpRequestMessage(request.Method, request.Uri);
        string? contentType = null;
        foreach (var header in request.Headers)
        {
            if (header.Key.Equals("Host", StringComparison.OrdinalIgnoreCase) ||
                header.Key.Equals("Content-Length", StringComparison.OrdinalIgnoreCase))
            {
                continue;
            }
            if (header.Key.Equals("Content-Type", StringComparison.OrdinalIgnoreCase))
            {
                contentType = header.Value;
                continue;
            }
            message.Headers.TryAddWithoutValidation(header.Key, header.Value);
        }
        if (request.Body is not null)
        {
            message.Content = new StringContent(request.Body, Encoding.UTF8);
            if (contentType is not null)
            {
                message.Content.Headers.Remove("Content-Type");
                message.Content.Headers.TryAddWithoutValidation("Content-Type", contentType);
            }
        }

        var started = TimeProvider.System.GetTimestamp();
        using var response = await client.SendAsync(message, HttpCompletionOption.ResponseHeadersRead, cancellationToken);
        var body = await response.Content.ReadAsStringAsync(cancellationToken);
        return new ReproResult((int)response.StatusCode, body, TimeProvider.System.GetElapsedTime(started));
    }
}
