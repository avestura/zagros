# Dictionary: "define <word>" lists English definitions; Tab on any row enters
# the Dictionary context, where typing a word shows its definitions.

API = "https://api.dictionaryapi.dev/api/v2/entries/en/"

def words_scope(word):
    return scope(id = "words", label = "Dictionary", placeholder = "Type a word", query = word)

def hint(word):
    return [item(
        id = "hint",
        title = "Dictionary",
        subtitle = "Press Tab, then type a word to look it up",
        scope = words_scope(word),
    )]

# Runs in the background (task): the launcher shows a loading row meanwhile.
def fetch(ctx, word):
    resp = http.get(API + http.quote(word))
    if resp.status == 404:
        return {"status": 404}
    if not resp.ok:
        return {"status": resp.status}
    rows = []
    for entry in resp.json():
        phonetic = entry.get("phonetic") or ""
        for meaning in entry.get("meanings", []):
            pos = meaning.get("partOfSpeech", "")
            for d in meaning.get("definitions", []):
                rows.append({"pos": pos, "phonetic": phonetic, "text": d.get("definition", ""), "example": d.get("example") or ""})
    return {"status": 200, "rows": rows}

def lookup(word):
    t = task("fetch", word)
    if not t.done:
        return [item(id = "loading", title = "Looking up '%s'" % word, subtitle = "Asking dictionaryapi.dev...", pending = True)]
    if not t.ok:
        return [item(id = "error", title = "Dictionary is not reachable", subtitle = t.error, scope = words_scope(word))]
    r = t.value
    if r["status"] == 404:
        return [item(id = "none", title = "No definition found for '%s'" % word, scope = words_scope(word))]
    if r["status"] != 200:
        return [item(id = "error", title = "Dictionary is not reachable", subtitle = "HTTP %d" % r["status"], scope = words_scope(word))]

    results = []
    for row in r["rows"]:
        if len(results) >= settings["max_results"]:
            break
        example = row["example"]
        results.append(item(
            id = "%s-%d" % (word, len(results)),
            title = row["text"],
            subtitle = ("%s %s %s" % (word, row["pos"], row["phonetic"])).strip() + ((' - "%s"' % example) if example else ""),
            action = action_copy("%s (%s): %s" % (word, row["pos"], row["text"])),
            scope = words_scope(word),
        ))
    if not results:
        results.append(item(id = "none", title = "No definition found for '%s'" % word, scope = words_scope(word)))
    return results

def search(query):
    word = query.strip().lower()
    if len(word) < 2:
        return hint(word)
    return lookup(word)

def on_query(ctx, query):
    return search(query)

def on_scope(ctx, scope_id, query, data):
    return search(query)
