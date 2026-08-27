# Canonical request model

Every importer produces `RequestSpec` version 1. The model preserves ordered headers and query pairs, a typed body, timeout, and provenance. It deliberately does not store runtime credentials outside placeholder form in sanitized artifacts.

Normalization is deterministic: methods and schemes are normalized by case, empty paths become `/`, default ports are omitted when formatting, and the original order of query pairs is preserved because reordering can change server behavior.

