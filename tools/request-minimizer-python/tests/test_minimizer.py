from __future__ import annotations

import pytest

from http_repro_minimizer import Budget, BudgetExceeded, minimize_json


def test_minimizes_top_level_fields_deterministically() -> None:
    value = {"customer": {"name": "Demo"}, "items": [1, 2], "unused": True}
    result = minimize_json(
        value, lambda candidate: isinstance(candidate, dict) and "items" in candidate
    )
    assert result.value == {"items": []}
    assert result.removed_nodes == 4


def test_rejects_unsatisfied_initial_predicate() -> None:
    with pytest.raises(ValueError, match="does not satisfy"):
        minimize_json({"ok": True}, lambda _candidate: False)


def test_enforces_request_budget() -> None:
    with pytest.raises(BudgetExceeded, match="budget"):
        minimize_json({"a": 1, "b": 2}, lambda _candidate: True, Budget(max_requests=1))


def test_does_not_mutate_input() -> None:
    value = {"required": [1, 2, 3], "unused": "x"}
    original = {"required": [1, 2, 3], "unused": "x"}
    minimize_json(value, lambda candidate: isinstance(candidate, dict) and "required" in candidate)
    assert value == original


def test_stable_key_order_gives_repeatable_result() -> None:
    value = {"z": 1, "required": True, "a": 2}

    def predicate(candidate: object) -> bool:
        return isinstance(candidate, dict) and "required" in candidate

    first = minimize_json(value, predicate)
    second = minimize_json(value, predicate)
    assert first == second


def test_keeps_required_list_element() -> None:
    value = ["noise", "marker", "other"]
    result = minimize_json(
        value, lambda candidate: isinstance(candidate, list) and "marker" in candidate
    )
    assert result.value == ["marker"]


@pytest.mark.parametrize("value", [None, True, 7, 2.5, "text"])
def test_scalar_is_already_minimal(value: object) -> None:
    result = minimize_json(value, lambda _candidate: True)  # type: ignore[arg-type]
    assert result.value == value
    assert result.removed_nodes == 0
