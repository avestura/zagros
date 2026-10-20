# Wikipedia: "wiki <topic>" searches article titles on en.wikipedia.org.

API = "https://en.wikipedia.org/w/rest.php/v1/search/title"

def article_url(key):
    return "https://en.wikipedia.org/wiki/" + http.quote(key.replace(" ", "_")).replace("%2F", "/")

# Runs in the background (task): the launcher shows a loading row meanwhile.
def search(ctx, q):
    resp = http.get("%s?q=%s&limit=%d" % (API, http.quote(q), settings["max_results"]))
    if not resp.ok:
        return {"status": resp.status}
    return {"status": 200, "pages": resp.json().get("pages", [])}

def on_query(ctx, query):
    q = query.strip()
    if len(q) < 2:
        return [item(
            id = "hint",
            title = "Search Wikipedia",
            subtitle = "Type at least two characters after 'wiki'",
        )]

    t = task("search", q)
    if not t.done:
        return [item(id = "loading", title = "Searching Wikipedia for '%s'" % q, subtitle = "Asking en.wikipedia.org...", pending = True)]
    if not t.ok:
        return [item(id = "error", title = "Wikipedia is not reachable", subtitle = t.error)]
    if t.value["status"] != 200:
        return [item(id = "error", title = "Wikipedia is not reachable", subtitle = "HTTP %d" % t.value["status"])]

    results = []
    for page in t.value["pages"]:
        key = page.get("key", "")
        if not key:
            continue
        url = article_url(key)
        results.append(item(
            id = str(page.get("id", key)),
            title = page.get("title", key),
            subtitle = page.get("description") or "Wikipedia article",
            action = action_open(url),
            alt = [{"name": "Copy link", "action": action_copy(url)}],
        ))
    if not results:
        results.append(item(
            id = "none",
            title = "No article found for '%s'" % q,
            action = action_open("https://en.wikipedia.org/w/index.php?search=" + http.quote(q)),
            subtitle = "Open a full-text search instead",
        ))
    return results
