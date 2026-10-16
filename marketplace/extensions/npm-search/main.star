# npm Search: "npm <query>" searches the npm registry.

API = "https://registry.npmjs.org/-/v1/search"

def on_query(ctx, query):
    q = query.strip()
    if len(q) < 2:
        return [item(
            id = "hint",
            title = "Search npm packages",
            subtitle = "Type at least two characters after 'npm'",
        )]

    resp = http.get("%s?text=%s&size=%d" % (API, http.quote(q), settings["max_results"]))
    if not resp.ok:
        return [item(id = "error", title = "npm registry is not reachable", subtitle = "HTTP %d" % resp.status)]

    results = []
    for obj in resp.json().get("objects", []):
        pkg = obj.get("package", {})
        name = pkg.get("name", "")
        if not name:
            continue
        page = "https://www.npmjs.com/package/" + name
        results.append(item(
            id = name,
            title = "%s  %s" % (name, pkg.get("version", "")),
            subtitle = pkg.get("description") or "No description",
            action = action_open(page),
            alt = [
                {"name": "Copy: npm install " + name, "action": action_copy("npm install " + name)},
                {"name": "Copy name", "action": action_copy(name)},
                {"name": "Open homepage", "action": action_open(pkg.get("links", {}).get("homepage") or page)},
            ],
        ))
    if not results:
        results.append(item(id = "none", title = "No packages found for '%s'" % q))
    return results
