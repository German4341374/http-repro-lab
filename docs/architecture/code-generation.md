# Code generation

Generators consume the sanitized canonical model. They never read the original capture and never invent credentials. Every output includes an explicit timeout, visible error handling, disabled automatic redirects, and a body representation consistent with RequestSpec.

The generated-client workflow starts one deterministic echo server, then compiles or executes cURL, JavaScript, TypeScript, Python, Go, Java, C#, and PHP outputs. It asserts method and query parity. The next parity expansion is exact body, content type, and normalized header comparison through a machine-readable receiver log.

Runtime defaults can still differ: proxy environment variables, automatic compression, default user agents, certificate stores, and URL encoding are language-specific. Those differences are documented rather than hidden.

