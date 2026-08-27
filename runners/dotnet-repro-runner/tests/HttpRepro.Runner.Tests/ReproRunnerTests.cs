using System.Net;
using HttpRepro.Runner;
using Xunit;

namespace HttpRepro.Runner.Tests;

public sealed class ReproRunnerTests
{
    [Fact]
    public async Task ExecuteAsyncPreservesMethodAndCapturesResponse()
    {
        var handler = new RecordingHandler();
        var request = new ReproRequest(HttpMethod.Post, new Uri("https://example.invalid/echo"), new Dictionary<string, string> { ["X-Demo"] = "synthetic" }, "{}", TimeSpan.FromSeconds(2));
        var result = await ReproRunner.ExecuteAsync(request, handler, TestContext.Current.CancellationToken);
        Assert.Equal(201, result.StatusCode);
        Assert.Equal("accepted", result.Body);
        Assert.Equal(HttpMethod.Post, handler.Method);
    }

    [Fact]
    public async Task ExecuteAsyncRejectsUnsupportedSchemes()
    {
        var request = new ReproRequest(HttpMethod.Get, new Uri("ftp://example.invalid/file"), new Dictionary<string, string>(), null, TimeSpan.FromSeconds(1));
        await Assert.ThrowsAsync<ArgumentException>(() => ReproRunner.ExecuteAsync(request, cancellationToken: TestContext.Current.CancellationToken));
    }

    private sealed class RecordingHandler : HttpMessageHandler
    {
        public HttpMethod? Method { get; private set; }
        protected override Task<HttpResponseMessage> SendAsync(HttpRequestMessage request, CancellationToken cancellationToken)
        {
            Method = request.Method;
            return Task.FromResult(new HttpResponseMessage(HttpStatusCode.Created) { Content = new StringContent("accepted") });
        }
    }
}
