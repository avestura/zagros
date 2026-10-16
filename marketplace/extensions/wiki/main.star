# Wikipedia: "wiki <topic>" searches article titles on en.wikipedia.org.

API = "https://en.wikipedia.org/w/rest.php/v1/search/title"

def article_url(key):
    return "https://en.wikipedia.org/wiki/" + http.quote(key.replace(" ", "_")).replace("%2F", "/")

def on_query(ctx, query):
    q = query.strip()
    if len(q) < 2:
        return [item(
            id = "hint",
            title = "Search Wikipedia",
            subtitle = "Type at least two characters after 'wiki'",
        )]

    resp = http.get("%s?q=%s&limit=%d" % (API, http.quote(q), settings["max_results"]))
    if not resp.ok:
        return [item(id = "error", title = "Wikipedia is not reachable", subtitle = "HTTP %d" % resp.status)]

    results = []
    for page in resp.json().get("pages", []):
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
