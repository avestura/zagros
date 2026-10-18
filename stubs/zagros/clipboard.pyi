"""clipboard: read and replace the system clipboard text.

Type description, not runnable Starlark: do not `import` anything. The name `clipboard` is predeclared in main.star; call `clipboard.get()` directly.
Needs: clipboard.read / clipboard.write.
Keywords: clipboard, paste, copy, pasteboard, copied text, system clipboard
"""

def get() -> str:
    """Current clipboard text.

    Example: text = clipboard.get()
    Task: read the clipboard
    Keywords: paste, read clipboard, get copied text, what is on the clipboard, pasteboard
    Needs: clipboard.read
    """

def set(text: str) -> None:
    """Replace the clipboard text.

    Example: clipboard.set("copied")
    Task: put text on the clipboard
    Keywords: copy text, write clipboard, set clipboard, copy to clipboard, put text on clipboard
    Needs: clipboard.write
    """
