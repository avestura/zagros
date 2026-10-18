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

def lookup(word):
    resp = http.get(API + http.quote(word))
    if resp.status == 404:
        return [item(id = "none", title = "No definition found for '%s'" % word, scope = words_scope(word))]
    if not resp.ok:
        return [item(id = "error", title = "Dictionary is not reachable", subtitle = "HTTP %d" % resp.status, scope = words_scope(word))]

    results = []
    for entry in resp.json():
        phonetic = entry.get("phonetic") or ""
        for meaning in entry.get("meanings", []):
            pos = meaning.get("partOfSpeech", "")
            for d in meaning.get("definitions", []):
                if len(results) >= settings["max_results"]:
                    return results
                text = d.get("definition", "")
                example = d.get("example") or ""
                results.append(item(
                    id = "%s-%d" % (word, len(results)),
                    title = text,
                    subtitle = ("%s %s %s" % (word, pos, phonetic)).strip() + ((' - "%s"' % example) if example else ""),
                    action = action_copy("%s (%s): %s" % (word, pos, text)),
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
