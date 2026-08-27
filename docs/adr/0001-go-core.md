# ADR 0001: Use Go for the core engine

Status: Accepted

Go provides a single cross-platform binary, explicit networking primitives, bounded concurrency, and a strong standard library. TypeScript remains the local control plane, while language-specific components must implement real behavior and tests.

