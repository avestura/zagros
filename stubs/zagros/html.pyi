"""HTML parsing: Parse HTML, select elements with CSS selectors, extract text and attributes.

Generated from the tools registry; do not edit by hand.
Type description, not runnable Starlark: do not `import` anything. The name `html` is predeclared in main.star; call `html.<name>(...)` directly.
Needs: none. These are pure functions: no file, network, clipboard or environment access.
Keywords: html, parse html, css selector, query selector, scrape, web page, dom, extract text, links, attributes, strip tags
"""

from typing import Any, Dict, List

def attr(text: str, selector: str, name: str) -> List[str]:
    """Return the value of an attribute for every element matching a CSS selector, in document order; elements without the attribute are skipped. Typical uses are href of links and src of images.

        text: The HTML text.
        selector: A CSS selector.
        name: The attribute name, for example href.

    Example: html.attr("<a href='/a'>x</a><a>y</a>", "a", "href")
    Keywords: get attribute, extract links, href, src, image urls, get href, collect links, html, attribute, links
    Needs: none (pure function).
    """

def outer_html(text: str, selector: str) -> List[str]:
    """Return the HTML source of every element matching a CSS selector (the element itself, re-serialized by the parser).

        text: The HTML text.
        selector: A CSS selector.

    Example: html.outer_html("<div><p>a</p></div>", "p")
    Keywords: outerHTML, element source, get markup, html snippet, html, markup, source
    Needs: none (pure function).
    """

def parse(text: str, fragment: bool = False) -> Any:
    """Parse HTML the way a browser does (missing tags are filled in, bad markup is repaired) into a tree. Every element is a dict with tag (lower case), attrs (dict of strings), text (the element's own text with whitespace collapsed) and children (list of elements). With fragment=True the markup is treated as the inside of a body and the result is a dict with tag "#fragment"; otherwise it is the html element. Comments and doctype are dropped. Never raises on bad markup; raises on input over 16 MB.

        text: The HTML text.
        fragment: Parse as a fragment rather than a whole page.

    Example: html.parse("<ul><li>a</li><li>b</li></ul>", fragment=True)
    Keywords: html to dict, html to json, dom, scrape, web page, read html, parse web page, html tree, html, parse, tree
    Needs: none (pure function).
    """

def select(text: str, selector: str, limit: int = 0) -> Any:
    """Find the elements of an HTML page that match a CSS selector and return them as element dicts (see html.parse). Supported: type, *, #id, .class, [attr] with = ~= |= ^= $= *= (and the i flag), the combinators space > + ~, comma lists, and :first-child, :last-child, :only-child, :nth-child(an+b|odd|even), :nth-last-child, :first-of-type, :last-of-type, :nth-of-type, :empty, :root, :not(), :is(), :checked, :disabled. Results are in document order; limit caps their number (0 means all). Raises on an invalid selector.

        text: The HTML text.
        selector: A CSS selector such as a.external[href^="https"].
        limit: Maximum number of results; 0 returns all.

    Example: html.select("<p class='x'>a</p><p>b</p>", "p.x")
    Keywords: css selector, querySelectorAll, jquery, find elements, scrape, cheerio, beautifulsoup, extract from html, html, select, css, selector, query, find
    Needs: none (pure function).
    """

def text(text: str, selector: str = "") -> str:
    """Extract the readable text of an HTML page, or of the elements matching a CSS selector. Scripts, styles, templates and comments are skipped, block elements and <br> break lines, and runs of spaces collapse. With a selector, the text of each match is returned on its own line.

        text: The HTML text.
        selector: Optional CSS selector; empty reads the whole page.

    Example: html.text("<h1>Title</h1><p>Hello <b>world</b></p>")
    Keywords: strip html, html to text, visible text, innerText, textContent, plain text from html, readable text, html, text, strip, readable
    Needs: none (pure function).
    """
