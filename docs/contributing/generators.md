# Generator contract

A generator declares one language, consumes a sanitized RequestSpec, emits minimal source, and provides a semantic verification step. New generators must set an explicit timeout, implement visible error handling, avoid hidden credentials, disable automatic redirects, and document runtime-specific behavior.

CI must compile or execute the result against the echo server and compare method, path, query, selected headers, content type, and body.

