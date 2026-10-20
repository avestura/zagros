# GitHub Repositories: "gh <query>" searches repositories.

API = "https://api.github.com/search/repositories"

# Runs in the background (task): the launcher shows a loading row meanwhile.
def search(ctx, q):
    url = "%s?q=%s&per_page=%d" % (API, http.quote(q), settings["max_results"])
    if settings["sort"] != "best-match":
        url += "&sort=" + settings["sort"]
    return http.json(url, headers = {"Accept": "application/vnd.github+json"})

def on_query(ctx, query):
    q = query.strip()
    if len(q) < 2:
        return [item(
            id = "hint",
            title = "Search GitHub repositories",
            subtitle = "Type at least two characters after 'gh'",
        )]

    t = task("search", q)
    if not t.done:
        return [item(id = "loading", title = "Searching GitHub for '%s'" % q, subtitle = "Asking api.github.com...", pending = True)]
    if not t.ok:
        return [item(id = "error", title = "GitHub is not reachable", subtitle = t.error)]

    results = []
    for repo in t.value.get("items", []):
        desc = repo.get("description") or "No description"
        results.append(item(
            id = str(repo["id"]),
            title = repo["full_name"],
            subtitle = "%d stars - %s" % (repo["stargazers_count"], desc),
            action = action_open(repo["html_url"]),
            alt = [
                {"name": "Copy clone URL", "action": action_copy(repo["clone_url"])},
                {"name": "Open issues", "action": action_open(repo["html_url"] + "/issues")},
            ],
        ))
    if not results:
        results.append(item(id = "none", title = "No repositories found for '%s'" % q))
    return results
