---
status: accepted
date: 2026-10-17
---
# ADR 0013: Tab enters a scope instead of opening a native menu

## Context
Tab used to open a native Windows popup menu with an item's secondary actions. It looked out of place, only listed actions, and gave extensions no way to offer a second level of choices (for web search, the list of engines).

## Decision
A `core.Scope` is a context the launcher can enter. An item carries `Scope`; Tab enters it. The input box shows a chip with the scope label, the list shows the items of the scope's owner (`ScopeProvider.QueryScope`, or a static list), and Esc or Backspace on an empty input leaves it, restoring the earlier text. Items with only `Alt` actions get a built-in "Actions" scope, so the native menu is gone. Extensions use the same mechanism through `scope(...)` on an item and an `on_scope` hook. Tab keeps two cheaper meanings that win over scopes: `Refresh` (regenerate passwords) and `Complete` (insert a keyword).

## Consequences
One host feature serves built-ins and extensions, and looks the same in every theme. Scopes are one level deep: Tab inside a scope does nothing. Scope answers are not streamed or ranked; the owner returns them in display order.
