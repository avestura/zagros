"""ai module: ask a language model, or let it call your functions. Needs: ai, scoped by provider name.

The user configures providers in zagros.toml ([[ai.provider]]: Anthropic Messages
API, or any OpenAI-compatible server such as OpenAI, Ollama or LM Studio) and
stores each API key with `zagros ai key set <name>`. The key lives in Windows
Credential Manager; the extension never sees it and it is removed from errors.
Declare the provider in the manifest as "ai:claude", or "ai" for any of them.

Budget: every call spends tokens (input plus output) from a per-extension daily
budget, `ai_budget_tokens` in manifest.toml (default 50000, resets at midnight
UTC, survives restarts). When it is used up every call raises a clear error;
max_tokens is lowered to fit what is left. Do not call ai from on_query on every
keystroke: use on_execute, an alternate action or a trigger word.

Time: the 5 second hook limit does not run while a request is in flight. One
request may take up to 60 seconds, one ai.run up to 120 seconds, and all waits
inside one hook call together up to 180 seconds. Replies over 8 MiB, prompts
over 4 MiB, 200 messages, 64 tools and 20 steps are refused.

Tools run on your thread with your extension's own permissions and nothing
more: a tool that calls fs.read_text still needs fs.read, and the model only
chooses the arguments. Give a tool the smallest job, validate what the model
passes, and treat text fetched from the web or files as untrusted input.

Type description, not runnable Starlark: do not `import` anything. The name `ai` is predeclared in main.star; call `ai.complete(...)` directly.
Keywords: ai, llm, language model, chatgpt, gpt, claude, anthropic, openai, ollama, local model, prompt, completion, chat, agent, tool use, function calling, summarize, translate, embeddings
"""

from typing import Any, Callable, Dict, List, Optional, Union

Message = Dict[str, Any]
"""{"role": "user" | "assistant" | "tool" | "system", "content": str}. An assistant
message may carry "tool_calls" (the list a previous result returned); a "tool"
message needs "tool_call_id" and may set "is_error"."""

class Tool:
    """What ai.tool and ai.registry_tools return. Read-only fields: name, description, params."""
    name: str
    description: str
    params: Dict[str, Any]

class Result:
    """The dict ai.complete and ai.run return."""
    text: str
    tool_calls: List[Dict[str, Any]]
    usage: Dict[str, int]
    stop_reason: str
    model: str
    provider: str

def complete(
    messages: Union[str, List[Message]],
    model: str = "",
    system: str = "",
    max_tokens: int = 1024,
    temperature: Optional[float] = None,
    tools: Optional[List[Tool]] = None,
    tool_choice: str = "",
    provider: str = "",
) -> Dict[str, Any]:
    """Send one request and return {text, tool_calls, usage, stop_reason, model, provider}.

    `messages` is a prompt string or a list of {"role", "content"} dicts. `model`
    and `provider` default to the configured ones. usage has input_tokens,
    output_tokens and total_tokens. stop_reason is "end", "tool_use" or
    "max_tokens". tool_calls is a list of {id, name, args}: this function does
    not run them (use ai.run for that, or answer them yourself with "tool"
    messages). tool_choice is "auto", "none", "required" or a tool name.
    Raises when the budget is used up, the provider fails or the permission is
    missing.

    Needs: ai.
    Example:
        reply = ai.complete("Summarize in one line: " + query, system="Be brief.", max_tokens=100)
        log.info(reply["text"])
    Keywords: ask ai, prompt, chat, llm, summarize, translate, rewrite, classify, extract with ai, generate text, gpt, claude, ollama
    Task: ask a language model a question
    """

def run(
    prompt: Union[str, List[Message]],
    tools: Optional[List[Tool]] = None,
    model: str = "",
    system: str = "",
    max_steps: int = 8,
    max_tokens: int = 1024,
    temperature: Optional[float] = None,
    provider: str = "",
) -> Dict[str, Any]:
    """Run an agent loop: ask, run the tools the model calls, send the results back, repeat until it answers or max_steps (1 to 20) is reached.

    Returns the last reply as {text, tool_calls, usage, stop_reason, model,
    provider, steps}. Here tool_calls is the trace of what ran, each
    {name, args, result, error}; usage adds up all steps. stop_reason is
    "end", "max_tokens" or "max_steps". A tool that raises is reported to the
    model as an error and the loop goes on; a timeout still aborts. Tools
    cannot call ai.run. The loop is only as safe as its tools: see the module
    notes.

    Needs: ai.
    Example:
        def add(a, b):
            return a + b
        tools = [ai.tool(add, description="Add two numbers", params={"a": "number", "b": "number"})]
        out = ai.run("What is 2.5 plus 4?", tools=tools, max_steps=4)
        log.info(out["text"])
    Keywords: agent, agent loop, tool calling, function calling, tool use, autonomous, multi step, assistant, mcp style, run tools
    Task: let a language model call my functions
    """

def tool(
    fn: Callable[..., Any],
    name: str = "",
    description: str = "",
    params: Optional[Dict[str, Any]] = None,
) -> Tool:
    """Describe a Starlark function so the model can call it.

    `name` defaults to the function name (letters, digits, _ and - only).
    `description` defaults to the docstring and is required: the model chooses
    tools from it. `params` is either shorthand, {"city": "string", "days":
    "integer"} (JSON type names, all required), or a full JSON schema
    {"type": "object", "properties": {...}, "required": [...]}. The model's
    arguments arrive as keyword arguments; the return value is sent back as
    text (a string as is, anything else as JSON, capped at 64 KiB).

    Needs: none, but ai.run and ai.complete need ai.
    Example:
        def lookup(word):
            return storage.get("glossary", {}).get(word, "unknown")
        t = ai.tool(lookup, description="Look a term up in the user's glossary", params={"word": "string"})
    Keywords: define tool, function schema, json schema, tool definition, expose function to model
    Task: describe a function as a model tool
    """

def registry_tools(names: List[str]) -> List[Tool]:
    """Offer built-in Zagros tools functions (hash, encode, text, num, dt and the other pure groups) to the model.

    Each name is "group.function" (hash.sha256) or a whole group (hash). The
    model sees them as group_function (hash_sha256) with the schema and
    description taken from the registry. They are pure, so they need no
    permission beyond ai itself. Up to 64 at a time.

    Needs: none, but ai.run and ai.complete need ai.
    Example: out = ai.run("What is the sha256 of 'zagros'?", tools=ai.registry_tools(["hash.sha256"]))
    Keywords: builtin tools, registry, give the model hash, encode, text utilities, calculator, expose tools
    Task: give a language model the built-in tools
    """

def embed(texts: Union[str, List[str]], model: str = "", provider: str = "") -> List[List[float]]:
    """Return one embedding vector per text, from a provider that has an embeddings API.

    Works with openai-kind providers (set embed_model in the config or pass
    model); an anthropic-kind provider raises a clear error. Up to 256 texts and
    4 MiB per call; input tokens count against the budget.

    Needs: ai.
    Example:
        vectors = ai.embed(["red apple", "green pear"])
        log.info(len(vectors[0]))
    Keywords: embeddings, vectors, semantic search, similarity, rag, text to vector
    Task: turn text into embedding vectors
    """

def providers() -> List[Dict[str, Any]]:
    """List the providers this extension may use: dicts with name, kind, model, embeddings (bool) and default (bool). Keys are never included.

    Needs: ai.
    Example:
        names = [p["name"] for p in ai.providers()]
        log.info(names)
    Keywords: list providers, which models, configured ai, available models
    Task: see which AI providers are configured
    """
