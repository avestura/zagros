# npm Search: "npm <query>" searches the npm registry.

API = "https://registry.npmjs.org/-/v1/search"

# Runs in the background (task): the launcher shows a loading row meanwhile.
def search(ctx, q):
    resp = http.get("%s?text=%s&size=%d" % (API, http.quote(q), settings["max_results"]))
    if not resp.ok:
        return {"status": resp.status}
    pkgs = []
    for obj in resp.json().get("objects", []):
        pkg = obj.get("package", {})
        if pkg.get("name"):
            pkgs.append({"name": pkg["name"], "version": pkg.get("version", ""), "description": pkg.get("description") or "", "homepage": pkg.get("links", {}).get("homepage") or ""})
    return {"status": 200, "packages": pkgs}

def on_query(ctx, query):
    q = query.strip()
    if len(q) < 2:
        return [item(
            id = "hint",
            title = "Search npm packages",
            subtitle = "Type at least two characters after 'npm'",
        )]

    t = task("search", q)
    if not t.done:
        return [item(id = "loading", title = "Searching npm for '%s'" % q, subtitle = "Asking registry.npmjs.org...", pending = True)]
    if not t.ok:
        return [item(id = "error", title = "npm registry is not reachable", subtitle = t.error)]
    if t.value["status"] != 200:
        return [item(id = "error", title = "npm registry is not reachable", subtitle = "HTTP %d" % t.value["status"])]

    results = []
    for pkg in t.value["packages"]:
        name = pkg["name"]
        page = "https://www.npmjs.com/package/" + name
        results.append(item(
            id = name,
            title = "%s  %s" % (name, pkg["version"]),
            subtitle = pkg["description"] or "No description",
            action = action_open(page),
            alt = [
                {"name": "Copy: npm install " + name, "action": action_copy("npm install " + name)},
                {"name": "Copy name", "action": action_copy(name)},
                {"name": "Open homepage", "action": action_open(pkg["homepage"] or page)},
            ],
        ))
    if not results:
        results.append(item(id = "none", title = "No packages found for '%s'" % q))
    return results
