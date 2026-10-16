# Hacker News

`hn` shows the current top stories. Add words to filter them by title. Enter
opens the article, the actions menu opens the comments or copies the link.

The story list is cached in extension storage for five minutes so typing a
filter does not hit the API on every keystroke.

Permissions: `network:hacker-news.firebaseio.com`, `open`, `clipboard.write`.
