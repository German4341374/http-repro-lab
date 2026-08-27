"""Small non-network CLI for deterministic minimizer verification."""

from __future__ import annotations

import argparse
import json
from pathlib import Path
from typing import Any

from .minimizer import Budget, JSONValue, minimize_json


def main() -> None:
    parser = argparse.ArgumentParser(
        description="Minimize a JSON body with a deterministic key predicate"
    )
    parser.add_argument("input", type=Path)
    parser.add_argument(
        "--required-key", required=True, help="Keep candidates containing this object key"
    )
    parser.add_argument("--output", type=Path, default=Path("minimized.json"))
    parser.add_argument("--max-requests", type=int, default=100)
    args = parser.parse_args()
    value: Any = json.loads(args.input.read_text(encoding="utf-8"))

    def predicate(candidate: JSONValue) -> bool:
        return isinstance(candidate, dict) and args.required_key in candidate

    result = minimize_json(value, predicate, Budget(max_requests=args.max_requests))
    args.output.write_text(result.model_dump_json(indent=2) + "\n", encoding="utf-8")
    print(f"Minimized body written to {args.output}; attempts={result.attempts}")


if __name__ == "__main__":
    main()
