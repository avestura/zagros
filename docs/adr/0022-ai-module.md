---
status: accepted
date: 2026-10-19
---
# ADR 0022: A permissioned ai module with provider config, budgets and tool use

## Context
Extensions want to summarize, translate or answer with a language model, and to let the model call their functions. A model call sends text off the machine, spends the user's money and can be steered by whatever text it reads, so it needs the same care as `network` and `secrets`.

## Decision
- Providers are configured by the user in `zagros.toml` (`[[ai.provider]]`: name, kind `anthropic` or `openai`, base_url, model, embed_model). The `openai` kind covers OpenAI and local servers such as Ollama. Extensions pick a provider by name and never see the endpoint credentials.
- API keys are not in the config. `zagros ai key set <name>` stores them in Credential Manager under `zagros-ai|<name>`, a prefix the `secrets` module cannot reach. Errors and logs have the key removed. Plain http is accepted only for local and private hosts, and redirects are never followed, so the key cannot leave for another host. Local servers need no key.
- New permission `ai`, scoped by provider name (`ai:claude`, case-insensitive, unscoped means all). The check is inside every host function, before the provider is looked up. The consent dialog states the daily limit.
- Each extension has a daily token budget, `ai_budget_tokens` in the manifest (default 50000, at most 100 million). The host counts input plus output tokens, keeps the count in the extension's data folder, lowers `max_tokens` to fit the rest and refuses calls once it is spent. A request whose estimated input alone exceeds the rest is refused.
- `ai.run` runs the tool loop on the extension's own thread. Tools are Starlark functions wrapped by `ai.tool`, or pure registry functions through `ai.registry_tools`, so a tool has exactly the extension's permissions and the model only chooses arguments. A raising tool is reported to the model; cancellation aborts the run. `ai.run` cannot nest. Steps are capped at 20, tool results at 64 KiB, replies at `MaxHTTPBytes`, prompts at `MaxFileBytes`.
- Timeouts: the 5 second hook limit would end every model call. The hook clock became pausable (`callClock`, installed by `withTimeout`); the ai module pauses it while a request is in flight. A request may take 60 seconds, an `ai.run` 120 seconds and all waits in one hook call 180 seconds. The runtime core changed only to create the clock and read the cancel cause, so cancellation still reports "context deadline exceeded".
- `ai.embed` works with `openai`-kind providers and raises a clear error for `anthropic`.
- The provider clients live in `internal/ai`, which knows nothing about extensions, so the wire formats are tested against fake servers without Starlark.

## Consequences
Extensions can use a cloud or local model without handling secrets, and the user bounds both who may call which provider and how much. The budget is per extension per day, not per call, so one noisy extension cannot drain another's allowance but can use its own. Prompt injection is not solved: a tool that reads files or the web hands that text to the model. The mitigation is that tools cannot exceed the extension's grants, which the docs say plainly. Settings are the config file (tray, Settings) and the `zagros ai` command; the extension manager window has no form for providers yet.
