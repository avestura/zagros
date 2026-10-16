---
status: accepted
date: 2026-10-16
---
# ADR 0001: Record decisions as ADRs

## Context
We build fast, mostly unattended, and need to see later why each feature exists and whether it shipped.

## Decision
Architecture decisions live in docs/adr/NNNN-slug.md with front-matter status (proposed, accepted, superseded). Delivery is tracked per workstream in docs/workstreams/, each linking the ADRs it implements. The website renders both.

## Consequences
Reviewers can trace a feature to a decision and a workstream. Changing a decision means a new ADR that supersedes the old one.
