---
status: accepted
date: 2026-10-18
---
# ADR 0017: Remembered picks per query and per Tab scope

## Context
Frecency (ADR 0001 era, `history.go`) boosts an item by how often it ran, whatever was typed. Users expect more: after typing `fir` and running Firefox, `fir` should put Firefox first, and the tool run in the String scope should lead that scope next time. Tool rows carry the user's text, so recording what was typed is unsafe for them.

## Decision
`core.Choices` (`choices.go`, stored in `choices.json`) keeps two maps of item picks: by normalized typed query, and by Tab scope id. The catalog adds a score bonus (above any fuzzy score, below the keyword-routing bonus) to items picked for the exact query, half of it for queries that extend a learned one (`fire` after `fir`), and reorders scope lists so picked items lead and the rest keep the provider's order. Scope picks never store text, and rows marked `nohistory` are only learned inside scopes.

The store is bounded and self-cleaning: 300 contexts (LRU), 6 items per context, queries of at most 40 characters, a 30 day half-life with deletion below a floor, deletion after three consecutive misses for the exact query (only when every provider answered, was ready, and none timed out), immediate deletion when an action fails, and no storage for items whose IDs are not stable (clipboard, calc, windows, processes, password, runner). An empty store removes its file.

Separately, the tools provider offers Encode and Hash (and Network, Date, ID or Lines when the input looks like one) as Tab scope rows for arbitrary input, next to Number and String. The catalog caps fallback rows at four and gives each provider its first row before any gets a second, so they never crowd out more than half of a short list.

## Consequences
Extensions get both behaviours for free through their item `id`, so the stub and agent guide ask for stable ids. Wrongly learned picks fade by themselves. The cost is one map lookup per keystroke (at most 40 prefix probes) and one small JSON file.
