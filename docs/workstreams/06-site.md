---
status: partial
adrs: 0010
---
# WS06: Website, docs and CI

- [x] Boxy-styled landing, docs, marketplace, decisions and roadmap pages
- [x] Pages, CI and release workflows
- [x] README and LICENSE

Notes:
- Site builds locally (25 pages, 0 broken internal links, Boxy lint 0 errors). The three workflows were written but never run on GitHub.
- Fallback docs under `site/content/` contain claims to verify against the finished UI and built-ins (Alt+Space default, Tab opens actions menu, default web searches).
- `release.yml` sets `-X main.version`; `cmd/zagros` needs a `version` variable.
