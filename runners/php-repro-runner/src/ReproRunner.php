<?php

declare(strict_types=1);

namespace HttpReproLab;

final class ReproRunner
{
    /**
     * @param array<string, string> $headers
     * @return array{status: int, body: string, durationMs: float}
     */
    public function execute(string $method, string $url, array $headers, ?string $body, int $timeoutMs): array
    {
        if ($url === '') {
            throw new \InvalidArgumentException('URL must not be empty.');
        }
        /** @var non-empty-string $validatedUrl */
        $validatedUrl = $url;
        $scheme = parse_url($validatedUrl, PHP_URL_SCHEME);
        if (!is_string($scheme) || !in_array($scheme, ['http', 'https'], true)) {
            throw new \InvalidArgumentException('Only HTTP and HTTPS URLs are supported.');
        }
        if ($timeoutMs < 1) {
            throw new \InvalidArgumentException('Timeout must be positive.');
        }
        $safeHeaders = [];
        foreach ($headers as $name => $value) {
            if (!in_array(strtolower($name), ['host', 'content-length'], true)) {
                $safeHeaders[] = $name . ': ' . $value;
            }
        }
        $handle = curl_init($validatedUrl);
        if ($handle === false) {
            throw new \RuntimeException('Unable to initialize cURL.');
        }
        curl_setopt_array($handle, [CURLOPT_CUSTOMREQUEST => strtoupper($method), CURLOPT_HTTPHEADER => $safeHeaders, CURLOPT_RETURNTRANSFER => true, CURLOPT_FOLLOWLOCATION => false, CURLOPT_TIMEOUT_MS => $timeoutMs, CURLOPT_SSL_VERIFYPEER => true, CURLOPT_SSL_VERIFYHOST => 2]);
        if ($body !== null) { curl_setopt($handle, CURLOPT_POSTFIELDS, $body); }
        $started = hrtime(true);
        $responseBody = curl_exec($handle);
        if ($responseBody === false) { $message = curl_error($handle); curl_close($handle); throw new \RuntimeException($message); }
        if (!is_string($responseBody)) { curl_close($handle); throw new \RuntimeException('cURL did not return a response body.'); }
        $status = curl_getinfo($handle, CURLINFO_RESPONSE_CODE); curl_close($handle);
        return ['status' => $status, 'body' => $responseBody, 'durationMs' => (hrtime(true) - $started) / 1_000_000.0];
    }
}
