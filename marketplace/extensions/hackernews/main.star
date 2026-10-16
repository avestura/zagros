# Hacker News: "hn" lists top stories, "hn <words>" filters them.

BASE = "https://hacker-news.firebaseio.com/v0"
TTL = 300  # seconds the story list is reused from storage

def load_stories():
    now = int(time.now().unix)
    n = min(max(settings["count"], 1), 15)
    cache = storage.get("cache")
    if cache and now - cache["at"] < TTL and cache["count"] == n:
        return cache["stories"]
    ids = http.json(BASE + "/topstories.json")[:n]
    stories = []
    for sid in ids:
        s = http.json("%s/item/%d.json" % (BASE, sid))
        if not s or s.get("dead") or s.get("deleted"):
            continue
        stories.append({
            "id": sid,
            "title": s.get("title", ""),
            "url": s.get("url", ""),
            "score": s.get("score", 0),
            "by": s.get("by", ""),
            "comments": s.get("descendants", 0),
        })
    storage.set("cache", {"at": now, "count": n, "stories": stories})
    return stories

def on_query(ctx, query):
    q = query.strip().lower()
    results = []
    for s in load_stories():
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
