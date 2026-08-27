"""Bounded delta-debugging for JSON-compatible request bodies."""

from __future__ import annotations

from collections.abc import Callable
from copy import deepcopy
from time import monotonic

from pydantic import BaseModel, ConfigDict, Field

type JSONValue = None | bool | int | float | str | list[JSONValue] | dict[str, JSONValue]
Predicate = Callable[[JSONValue], bool]


class Budget(BaseModel):
    """Hard limits for predicate executions."""

    model_config = ConfigDict(frozen=True)
    max_requests: int = Field(default=100, ge=1, le=10_000)
    max_duration_seconds: float = Field(default=30.0, gt=0, le=3_600)


class BudgetExceeded(RuntimeError):
    """Raised when deterministic minimization exhausts its budget."""


class MinimizeResult(BaseModel):
    model_config = ConfigDict(frozen=True)
    value: JSONValue
    attempts: int
    removed_nodes: int


class _Counter:
    def __init__(self, budget: Budget) -> None:
        self.budget = budget
        self.attempts = 0
        self.started = monotonic()

    def check(self, value: JSONValue, predicate: Predicate) -> bool:
        if self.attempts >= self.budget.max_requests:
            raise BudgetExceeded("MINIMIZATION_BUDGET_EXCEEDED: request budget exhausted")
        if monotonic() - self.started >= self.budget.max_duration_seconds:
            raise BudgetExceeded("MINIMIZATION_BUDGET_EXCEEDED: duration budget exhausted")
        self.attempts += 1
        return predicate(deepcopy(value))


def minimize_json(
    value: JSONValue, predicate: Predicate, budget: Budget | None = None
) -> MinimizeResult:
    """Minimize a JSON value in stable key/index order.

    The predicate must be deterministic and side-effect safe. Network predicates remain the caller's
    responsibility because write authorization and target policy belong to the Go execution engine.
    """

    counter = _Counter(budget or Budget())
    original = deepcopy(value)
    if not counter.check(original, predicate):
        raise ValueError("the original value does not satisfy the reproduction predicate")
    minimized, removed = _minimize_node(original, predicate, counter)
    return MinimizeResult(value=minimized, attempts=counter.attempts, removed_nodes=removed)


def _minimize_node(
    value: JSONValue, predicate: Predicate, counter: _Counter
) -> tuple[JSONValue, int]:
    removed = 0
    if isinstance(value, dict):
        current = deepcopy(value)
        for key in sorted(list(current)):
            candidate_map = deepcopy(current)
            del candidate_map[key]
            if counter.check(candidate_map, predicate):
                current = candidate_map
                removed += 1
        for key in sorted(current):
            child = current[key]

            def child_predicate(candidate_child: JSONValue, *, current_key: str = key) -> bool:
                candidate_parent = deepcopy(current)
                candidate_parent[current_key] = candidate_child
                return predicate(candidate_parent)

            reduced, child_removed = _minimize_node(child, child_predicate, counter)
            current[key] = reduced
            removed += child_removed
        return current, removed
    if isinstance(value, list):
        current_list = deepcopy(value)
        index = 0
        while index < len(current_list):
            candidate_list = current_list[:index] + current_list[index + 1 :]
            if counter.check(candidate_list, predicate):
                current_list = candidate_list
                removed += 1
            else:
                index += 1
        return current_list, removed
    return value, removed
