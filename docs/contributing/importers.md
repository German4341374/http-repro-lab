# Importer contract

An importer declares whether it recognizes an input, parses it without executing content, and normalizes it to RequestSpec v1. It must preserve semantically meaningful ordering, attach provenance and source SHA-256, return stable error codes, enforce input limits, and have malformed/adversarial fixtures.

Importers never reproduce requests directly. This separation guarantees that sanitization and network policy run for every input format.

