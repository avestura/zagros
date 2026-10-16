# Weather: "weather [place]" shows current conditions and three days from wttr.in.

def first(lst, key = "value"):
    if lst and len(lst) > 0:
        return lst[0].get(key, "")
    return ""

def on_query(ctx, query):
    place = query.strip()
    if not place:
        place = settings["default_location"].strip()
    metric = settings["units"] == "metric"
    t_unit = "C" if metric else "F"
    w_unit = "km/h" if metric else "mph"

    resp = http.get("https://wttr.in/%s?format=j1" % http.quote(place), headers = {"Accept-Language": "en"})
    if not resp.ok:
        return [item(id = "error", title = "Weather is not available", subtitle = "wttr.in answered HTTP %d" % resp.status)]
    data = resp.json()

    cur = data["current_condition"][0]
    area = data.get("nearest_area", [{}])[0]
    where = first(area.get("areaName"))
    country = first(area.get("country"))
    if country:
        where = "%s, %s" % (where, country)
    link = "https://wttr.in/" + http.quote(place)

    temp = cur["temp_" + t_unit]
    feels = cur["FeelsLike" + t_unit]
    wind = cur["windspeedKmph" if metric else "windspeedMiles"]
    summary = "%s: %s, %s°%s (feels %s°), humidity %s%%, wind %s %s %s" % (
        where, first(cur.get("weatherDesc")), temp, t_unit, feels,
        cur["humidity"], wind, w_unit, cur.get("winddir16Point", ""),
    )

    results = [item(
        id = "now",
        title = "%s°%s  %s" % (temp, t_unit, first(cur.get("weatherDesc"))),
        subtitle = "%s - feels %s°, humidity %s%%, wind %s %s" % (where, feels, cur["humidity"], wind, w_unit),
        action = action_open(link),
        alt = [{"name": "Copy summary", "action": action_copy(summary)}],
    )]
    for day in data.get("weather", []):
        hourly = day.get("hourly", [])
        desc = first(hourly[len(hourly) // 2].get("weatherDesc")) if hourly else ""
        results.append(item(
            id = day["date"],
            title = "%s  %s° / %s°%s" % (day["date"], day["maxtemp" + t_unit], day["mintemp" + t_unit], t_unit),
            subtitle = desc,
            action = action_open(link),
        ))
    return results
