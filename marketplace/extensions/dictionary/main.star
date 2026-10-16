# Dictionary: "define <word>" lists English definitions.

API = "https://api.dictionaryapi.dev/api/v2/entries/en/"

def on_query(ctx, query):
    word = query.strip().lower()
    if len(word) < 2:
        return [item(id = "hint", title = "Look up an English word", subtitle = "Type a word after 'define'")]

    resp = http.get(API + http.quote(word))
    if resp.status == 404:
        return [item(id = "none", title = "No definition found for '%s'" % word)]
    if not resp.ok:
        return [item(id = "error", title = "Dictionary is not reachable", subtitle = "HTTP %d" % resp.status)]

    results = []
    for entry in resp.json():
        phonetic = entry.get("phonetic", "")
        for meaning in entry.get("meanings", []):
            pos = meaning.get("partOfSpeech", "")
            for d in meaning.get("definitions", []):
                if len(results) >= settings["max_results"]:
                    return results
                text = d.get("definition", "")
                example = d.get("example", "")
                results.append(item(
                    id = "%s-%d" % (word, len(results)),
                    title = text,
                    subtitle = ("%s %s" % (pos, phonetic)).strip() + ((' - "%s"' % example) if example else ""),
                    action = action_copy("%s (%s): %s" % (word, pos, text)),
                    alt = [{"name": "Copy word", "action": action_copy(word)}],
                ))
    if not results:
        results.append(item(id = "none", title = "No definition found for '%s'" % word))
    return results
