# Snippets: "sn" lists snippets, "sn <filter>" searches them,
# "sn add <name> = <text>" saves one. Data lives in extension storage.

PREFIX = "s:"

def all_snippets():
    out = {}
    for k in storage.keys():
        if k.startswith(PREFIX):
            out[k[len(PREFIX):]] = storage.get(k)
    return out

def on_query(ctx, query):
    q = query.strip()

    if q == "add" or q.startswith("add "):
        body = q[3:].strip()
        if "=" not in body:
            return [item(id = "add-hint", title = "Save a snippet", subtitle = "sn add <name> = <text>")]
        name, text = body.split("=", 1)
        name, text = name.strip(), text.strip()
        if not name or not text:
            return [item(id = "add-hint", title = "Save a snippet", subtitle = "sn add <name> = <text>")]
        replacing = storage.get(PREFIX + name) != None
        return [item(
            id = "add:" + name,
            title = ("Replace snippet '%s'" if replacing else "Save snippet '%s'") % name,
            subtitle = text,
            action = action_call("add", {"name": name, "text": text}),
        )]

    snippets = all_snippets()
    if not snippets:
        return [item(id = "empty", title = "No snippets yet", subtitle = "sn add <name> = <text>")]

    words = q.lower().split()
    results = []
    for name in sorted(snippets.keys()):
        text = snippets[name]
        hay = (name + " " + text).lower()
        if not all([w in hay for w in words]):
            continue
        results.append(item(
            id = "snippet:" + name,
            title = name,
            subtitle = text.replace("\n", " "),
            action = action_copy(text),
            alt = [{"name": "Delete snippet", "action": action_call("delete", name)}],
        ))
    if not results:
        results.append(item(id = "none", title = "No snippet matches '%s'" % q, subtitle = "sn add <name> = <text>"))
    return results

def on_execute(ctx, it):
    if it.name == "add":
        storage.set(PREFIX + it.data["name"], it.data["text"])
    elif it.name == "delete":
        storage.delete(PREFIX + it.data)
    return None
