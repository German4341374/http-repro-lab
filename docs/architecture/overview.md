# Architecture overview

```mermaid
flowchart LR
  I[HAR / cURL / raw HTTP] --> P[Go importers]
  P --> C[Canonical RequestSpec v1]
  C --> S[Secret detection and sanitization]
  S --> E[Policy-gated execution engine]
  E --> D[Comparison and evidence]
  C --> G[Multi-language generators]
  D --> R[Offline report and repro pack]
  A[Fastify local API] --> C
  M[Python minimizer] --> E
  DB[(PostgreSQL server mode)] --> A
```

The Go binary is the trusted orchestration boundary. Imports are data, never executable shell. Original sources are immutable and addressed by SHA-256. Network execution is explicit, bounded, and governed by method and target policies. The browser report is static and renders all evidence with text-only DOM APIs.

