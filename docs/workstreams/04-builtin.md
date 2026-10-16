---
status: partial
adrs: 0008
---
# WS04: Built-in providers

- [x] Apps (Start menu, Store apps)
- [x] Everything search
- [x] Calculator, units, text tools, passwords
- [x] System commands, windows, processes
- [x] Web search and aliases, clipboard history, emoji, help

Notes:
- Tests pass for calc, units, passwords and text tools. Apps ran on a dev machine (360 Start Menu entries found).
- Everything integration is untested against a real Everything install; only the "not found" path is verified.
- Clipboard history capture is untested against real clipboard changes.
- Executor must honour `Extra["confirm"]` and `Extra["nohistory"]`, and run `Func` actions.
