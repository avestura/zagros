---
status: accepted
date: 2026-10-16
---
# ADR 0009: Git-based marketplace with direct GitHub references

## Context
No budget or wish for a backend server. People should be able to install straight from a repository.

## Decision
A registry lives in the repo (registry/extensions/*.toml, added by pull request). A generator builds registry.json, published on GitHub Pages. Clients also install from github:owner/repo[@tag], a repo URL or a direct .zagros URL. Downloads are https only, host allowlisted, size capped, hashed.

## Consequences
Zero hosting cost and auditable listings. Rate limits on the GitHub API apply, so a token is optional but supported. Trust comes from the permission prompt, with a verified flag for reviewed entries.
