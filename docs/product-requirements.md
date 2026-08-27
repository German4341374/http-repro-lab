# Product requirements

HTTP Repro Lab turns untrusted HTTP evidence into a safe and repeatable diagnostic artifact. The primary user is an engineer who receives a HAR, cURL command, or raw request and needs to determine what happened without leaking credentials or executing input as shell code.

## Product questions

1. What request actually happened?
2. Which values are sensitive?
3. Can the request be reproduced safely?
4. What differs between two targets?
5. What is the smallest useful reproduction?

## v1 acceptance path

The first production path imports a HAR, normalizes its entries, detects and replaces secrets, reproduces an explicitly selected safe request against a permitted target, compares responses, emits generated clients, and writes a self-contained offline report. Findings contain evidence and avoid unsupported causal claims.

## Non-goals

The project is not an API gateway, load-testing system, browser automation platform, hosted collaboration product, or general vulnerability scanner.

