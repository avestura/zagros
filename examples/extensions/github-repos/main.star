# GitHub Repositories: "gh <query>" searches repositories.

API = "https://api.github.com/search/repositories"

def on_query(ctx, query):
    q = query.strip()
    if len(q) < 2:
        return [item(
            id = "hint",
            title = "Search GitHub repositories",
            subtitle = "Type at least two characters after 'gh'",
        )]

    url = "%s?q=%s&per_page=%d" % (API, http.quote(q), settings["max_results"])
    if settings["sort"] != "best-match":
        url += "&sort=" + settings["sort"]

    data = http.json(url, headers = {"Accept": "application/vnd.github+json"})
    results = []
    for repo in data.get("items", []):
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
