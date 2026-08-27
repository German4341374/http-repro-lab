# ADR 0002: Normalize all inputs into RequestSpec

Status: Accepted

Importers are isolated from execution and generation by a versioned canonical schema. This makes sanitization, comparisons, and semantic parity tests independent of the source format.

