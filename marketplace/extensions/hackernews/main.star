# Hacker News: "hn" lists top stories, "hn <words>" filters them.

BASE = "https://hacker-news.firebaseio.com/v0"
TTL = 300  # seconds the story list is reused from storage

# Runs in the background (task): the launcher shows a loading row meanwhile.
def load_stories(ctx, n):
    now = int(time.now().unix)
    cache = storage.get("cache")
    if cache and now - cache["at"] < TTL and cache["count"] == n:
        return cache["stories"]
    ids = http.json(BASE + "/topstories.json")[:n]
    replies = http.fetch_all(["%s/item/%d.json" % (BASE, sid) for sid in ids], 8)
    stories = []
    for i in range(len(ids)):
        if not replies[i].ok:
            continue
        s = replies[i].json()
        if not s or s.get("dead") or s.get("deleted"):
            continue
        stories.append({
            "id": ids[i],
            "title": s.get("title", ""),
            "url": s.get("url", ""),
            "score": s.get("score", 0),
            "by": s.get("by", ""),
            "comments": s.get("descendants", 0),
        })
    if not stories:
        fail("no stories came back")
    storage.set("cache", {"at": now, "count": n, "stories": stories})
    return stories

def on_query(ctx, query):
    q = query.strip().lower()
    t = task("load_stories", min(max(settings["count"], 1), 15))
    if not t.done:
        return [item(id = "loading", title = "Loading top stories", subtitle = "Asking Hacker News...", pending = True)]
    if not t.ok:
        return [item(id = "error", title = "Hacker News is not reachable", subtitle = t.error)]
    results = []
    for s in t.value:
        if q and q not in s["title"].lower():
            continue
        thread = "https://news.ycombinator.com/item?id=%d" % s["id"]
        results.append(item(
            id = str(s["id"]),
            title = s["title"],
            subtitle = "%d points by %s, %d comments" % (s["score"], s["by"], s["comments"]),
            action = action_open(s["url"] or thread),
            alt = [
                {"name": "Open comments", "action": action_open(thread)},
                {"name": "Copy link", "action": action_copy(s["url"] or thread)},
            ],
        ))
    if not results:
        results.append(item(id = "none", title = "No matching stories"))
    return results
