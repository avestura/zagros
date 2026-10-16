# Hello: the smallest useful Zagros extension.

def on_load(ctx):
    log.info("hello extension loaded, version", ctx.version)

def on_query(ctx, query):
    name = query.strip() or "world"
    greeting = "Hello, %s!" % name
    return [
        item(
            id = "greeting",
            title = greeting,
            subtitle = "Press Enter to copy",
            action = action_copy(greeting),
        ),
    ]
