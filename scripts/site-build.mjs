#!/usr/bin/env node
// Zagros site builder. Dependency-free, Node 18+.
//   node scripts/site-build.mjs            -> writes site-dist/
// Sources: docs/**/*.md (and site/content/*.md as fallbacks), site/assets/*.
// Missing source files are skipped with a warning.

import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { render, parseFrontMatter, escapeHtml as esc } from './site-markdown.mjs';

const ROOT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const OUT = path.join(ROOT, 'site-dist');
const SITE_URL = (process.env.SITE_URL || 'https://avestura.github.io/zagros').replace(/\/$/, '');
const MARK = '<svg class="site-mark__glyph" viewBox="0 0 16 16" width="22" height="22" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="square" stroke-linejoin="miter" aria-hidden="true"><path d="M1 13L4.5 7L6.5 10L9 3L11.5 8L13 6L15 13"/></svg>';
const REPO = 'https://github.com/avestura/zagros';
const warn = (m) => console.warn(`warning: ${m}`);

const readIf = (p) => (fs.existsSync(p) ? fs.readFileSync(p, 'utf8') : null);
const toPosix = (p) => p.split(path.sep).join('/');

function walk(dir, ext) {
  if (!fs.existsSync(dir)) return [];
  const res = [];
  for (const e of fs.readdirSync(dir, { withFileTypes: true })) {
    const p = path.join(dir, e.name);
    if (e.isDirectory()) res.push(...walk(p, ext));
    else if (e.name.endsWith(ext)) res.push(p);
  }
  return res.sort();
}

// ---------------------------------------------------------------- page model
const pages = []; // {src, out, title, group, order, html, headings, text, desc, kind}
const routes = new Map(); // repo-relative source path -> out path

// Docs layout: [source candidates, output slug, nav title override, group, order]
const DOC_MAP = [
  [['site/content/getting-started.md', 'docs/getting-started.md'], 'getting-started', 'Getting started', 'Guide', 1],
  [['site/content/configuration.md', 'docs/configuration.md'], 'configuration', 'Configuration', 'Guide', 2],
  [['docs/ui.md'], 'ui', 'The launcher UI', 'Guide', 3],
  [['docs/builtin-providers.md'], 'builtin-providers', 'Built-in providers', 'Guide', 4],
  [['docs/extensions/README.md'], 'extensions', 'Extensions guide and API', 'Extensions', 1],
  [['site/content/permissions.md', 'docs/permissions.md'], 'permissions', 'Permissions', 'Extensions', 2],
  [['docs/marketplace.md'], 'marketplace', 'Marketplace and publishing', 'Extensions', 3],
  [['docs/themes.md'], 'themes', 'Themes', 'Customize', 1],
  [['docs/ARCHITECTURE.md'], 'architecture', 'Architecture', 'Project', 1],
];

function makeResolver(srcRel, outPath) {
  return (href) => {
    if (/^([a-z][a-z0-9+.-]*:|#|\/\/)/i.test(href)) return href;
    const [file, hash] = href.split('#');
    const abs = path.posix.normalize(path.posix.join(path.posix.dirname(srcRel), file));
    if (file.endsWith('.md')) {
      const hit = routes.get(abs) || routes.get(abs.replace(/^docs\//, 'site/content/'));
      if (hit) return relTo(outPath, hit) + (hash ? '#' + hash : '');
      warn(`${srcRel}: unresolved link ${href}`);
    }
    if (abs.startsWith('..')) return href;
    const isDir = !path.extname(abs) || href.endsWith('/');
    return `${REPO}/${isDir ? 'tree' : 'blob'}/main/${abs}${hash ? '#' + hash : ''}`;
  };
}

function relTo(fromOut, toOut) {
  const rel = path.posix.relative(path.posix.dirname(fromOut), toOut);
  return rel || path.posix.basename(toOut);
}
const rootOf = (out) => '../'.repeat(out.split('/').length - 1);

function firstParagraph(md) {
  const m = md.split(/\n\s*\n/).map((s) => s.trim()).find((s) => s && !/^(#|```|\||[-*]\s|>|<)/.test(s));
  return m ? m.replace(/\s+/g, ' ').replace(/[`*_]|\[([^\]]+)\]\([^)]*\)/g, '$1').slice(0, 200) : '';
}

function loadDoc(srcRel, out, extra = {}) {
  const raw = readIf(path.join(ROOT, srcRel));
  if (raw === null) return null;
  const { data, body } = parseFrontMatter(raw);
  const page = { src: srcRel, out, data, body, ...extra };
  page.desc = firstParagraph(body);
  return page;
}

function renderPage(p) {
  const r = render(p.body, { resolveLink: makeResolver(p.src, p.out) });
  p.html = r.html;
  p.headings = r.headings;
  p.text = r.text;
  p.title = p.title || r.title || path.basename(p.out, '.html');
  p.pageTitle = r.title || p.title;
}

// docs pages
const docPages = [];
for (const [cands, slug, title, group, order] of DOC_MAP) {
  const src = cands.find((c) => fs.existsSync(path.join(ROOT, c)));
  if (!src) { warn(`docs source missing for "${title}" (${cands.join(' or ')}), skipped`); continue; }
  const p = loadDoc(src, `docs/${slug}.html`, { title, group, order, kind: 'doc' });
  docPages.push(p);
  routes.set(src, p.out);
}
// extra extension docs: docs/extensions/*.md besides README
for (const f of walk(path.join(ROOT, 'docs/extensions'), '.md')) {
  const rel = toPosix(path.relative(ROOT, f));
  if (rel === 'docs/extensions/README.md') continue;
  const slug = 'extensions-' + path.basename(f, '.md').toLowerCase();
  const p = loadDoc(rel, `docs/${slug}.html`, { group: 'Extensions', order: 10, kind: 'doc' });
  docPages.push(p);
  routes.set(rel, p.out);
}
// any other docs/*.md at top level not mapped (kept out of ADR / workstreams)
for (const f of walk(path.join(ROOT, 'docs'), '.md')) {
  const rel = toPosix(path.relative(ROOT, f));
  if (routes.has(rel) || /^docs\/(adr|workstreams|extensions)\//.test(rel)) continue;
  const slug = path.basename(f, '.md').toLowerCase();
  const p = loadDoc(rel, `docs/${slug}.html`, { group: 'More', order: 50, kind: 'doc' });
  docPages.push(p);
  routes.set(rel, p.out);
}

// ADRs
const adrs = [];
for (const f of walk(path.join(ROOT, 'docs/adr'), '.md')) {
  const rel = toPosix(path.relative(ROOT, f));
  const base = path.basename(f, '.md');
  if (base.toLowerCase() === 'readme') { routes.set(rel, 'decisions.html'); continue; }
  const p = loadDoc(rel, `decisions/${base}.html`, { kind: 'adr' });
  p.num = (/^(\d+)/.exec(base) || [])[1] || '';
  adrs.push(p);
  routes.set(rel, p.out);
}
// workstreams
const streams = [];
for (const f of walk(path.join(ROOT, 'docs/workstreams'), '.md')) {
  const rel = toPosix(path.relative(ROOT, f));
  const base = path.basename(f, '.md');
  if (base.toLowerCase() === 'readme') { routes.set(rel, 'roadmap.html'); continue; }
  const p = loadDoc(rel, 'roadmap.html', { kind: 'stream', id: base });
  streams.push(p);
  routes.set(rel, `roadmap.html#${base}`);
}
routes.set('docs/adr/README.md', 'decisions.html');
routes.set('docs/workstreams/README.md', 'roadmap.html');

[...docPages, ...adrs, ...streams].forEach(renderPage);
const adrReadme = readIf(path.join(ROOT, 'docs/adr/README.md'));
const wsReadme = readIf(path.join(ROOT, 'docs/workstreams/README.md'));
if (adrReadme === null) warn('docs/adr/README.md missing');
if (wsReadme === null) warn('docs/workstreams/README.md missing');

docPages.sort((a, b) => ['Guide', 'Extensions', 'Customize', 'Project', 'More'].indexOf(a.group) - ['Guide', 'Extensions', 'Customize', 'Project', 'More'].indexOf(b.group) || a.order - b.order);

// ------------------------------------------------------------------ chrome
const sprite = (() => {
  const s = readIf(path.join(ROOT, 'site/assets/boxy-icons.svg'));
  if (!s) { warn('site/assets/boxy-icons.svg missing'); return ''; }
  return s.replace(/<\?xml[^>]*>\s*/, '').replace(/<svg\b([^>]*)>/, (m, a) => `<svg${a.replace(/\s(width|height|style)="[^"]*"/g, '')} width="0" height="0" style="position: absolute" aria-hidden="true">`);
})();

const NAV = [
  ['Docs', 'docs/getting-started.html', 'docs'],
  ['Marketplace', 'marketplace.html', 'marketplace'],
  ['Decisions', 'decisions.html', 'decisions'],
  ['Roadmap', 'roadmap.html', 'roadmap'],
];

const SEARCH_ICON = '<svg class="bx-icon" aria-hidden="true"><use href="#i-search"/></svg>';
const CLOSE_ICON = '<svg class="bx-icon" aria-hidden="true"><use href="#i-x"/></svg>';
const GH_ICON = '<svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor" aria-hidden="true"><path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27s1.36.09 2 .27c1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.01 8.01 0 0 0 16 8c0-4.42-3.58-8-8-8z"/></svg>';

function shell({ out, title, desc, mode = 'blueprint', section = '', body, noindex = false }) {
  const R = rootOf(out);
  const full = title === 'Zagros' ? 'Zagros: a keyboard launcher for Windows' : `${title} | Zagros`;
  const d = esc(desc || 'Zagros is a keyboard launcher for Windows with Starlark extensions, permissions, themes and a Git-based marketplace. Inspired by Keypirinha.');
  const canonical = `${SITE_URL}/${out === 'index.html' ? '' : out}`;
  const nav = NAV.map(([label, href, key]) => `<a href="${R}${href}"${key === section ? ' aria-current="page"' : ''}>${label}</a>`).join('');
  return `<!doctype html>
<html lang="en" data-mode="${mode}" data-base="${R}">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>${esc(full)}</title>
<meta name="description" content="${d}">
<meta property="og:title" content="${esc(full)}">
<meta property="og:description" content="${d}">
<meta property="og:type" content="website">
<meta property="og:url" content="${canonical}">
${noindex ? '<meta name="robots" content="noindex">\n' : ''}<link rel="canonical" href="${canonical}">
<link rel="icon" type="image/svg+xml" href="${R}favicon.svg">
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=Inter+Tight:wght@500;600;700&family=JetBrains+Mono:wght@400;500&display=swap">
<link rel="stylesheet" href="${R}assets/boxy.css">
<link rel="stylesheet" href="${R}assets/boxy-components.css">
<link rel="stylesheet" href="${R}assets/site.css">
<link rel="stylesheet" href="${R}assets/zagros.css">
<script>try{var r=document.documentElement,t=localStorage.getItem('boxy:theme');if(t)r.setAttribute('data-theme',t);}catch(e){}</script>
</head>
<body>
<a class="skip" href="#main">Skip to content</a>
${sprite}
<header class="site-head">
  <div class="site-head__inner">
    <a class="site-mark" href="${R}index.html">${MARK}ZAGROS</a>
    <nav class="site-nav" aria-label="Primary">${nav}</nav>
    <span class="site-head__spacer"></span>
    <div class="site-controls">
      <div class="site-search">
        <button class="ctl ctl--icon site-search__toggle" type="button" aria-label="Search the documentation" aria-expanded="false" aria-controls="site-search">${SEARCH_ICON}</button>
        <div class="bx-search bx-search--sm site-search__field">
          ${SEARCH_ICON}
          <input id="site-search" class="bx-input" type="search" placeholder="Search docs" aria-label="Search the documentation" autocomplete="off" spellcheck="false" role="combobox" aria-expanded="false" aria-controls="search-results" aria-autocomplete="list">
          <kbd class="bx-kbd site-search__hint" aria-hidden="true">/</kbd>
          <button class="ctl ctl--icon site-search__close" type="button" aria-label="Close search">${CLOSE_ICON}</button>
        </div>
        <div id="search-results" class="search-results" role="listbox" aria-label="Search results" hidden></div>
      </div>
      <button class="ctl" data-act="theme" title="Toggle theme"><span class="val">light</span></button>
      <a class="ctl ctl--icon" href="${REPO}" aria-label="GitHub repository">${GH_ICON}</a>
    </div>
  </div>
</header>
<main id="main" class="site-main">
${body}
</main>
<footer class="site-foot">
  <div class="site-foot__cols">
    <div>
      <a class="site-mark" href="${R}index.html">${MARK}ZAGROS</a>
      <p class="muted mt-4" style="font-size: var(--bx-text-base); max-inline-size: 40ch">A keyboard launcher for Windows. Inspired by <a href="https://keypirinha.com/">Keypirinha</a> by Jean-Charles Lefebvre.</p>
    </div>
    <div><p class="bx-label">Docs</p><ul>
      <li><a href="${R}docs/getting-started.html">Getting started</a></li>
      <li><a href="${R}docs/extensions.html">Extensions</a></li>
      <li><a href="${R}docs/themes.html">Themes</a></li>
      <li><a href="${R}marketplace.html">Marketplace</a></li></ul></div>
    <div><p class="bx-label">Project</p><ul>
      <li><a href="${REPO}">GitHub</a></li>
      <li><a href="${REPO}/releases/latest">Releases</a></li>
      <li><a href="${R}decisions.html">Decisions</a></li>
      <li><a href="${R}roadmap.html">Roadmap</a></li></ul></div>
    <div><p class="bx-label">Credits</p><ul>
      <li><a href="https://keypirinha.com/">Keypirinha</a> (inspiration)</li>
      <li><a href="https://github.com/avestura/boxy-design-skill">Boxy design system</a> (MIT)</li>
      <li><a href="https://starlark-lang.org/">Starlark</a></li></ul></div>
  </div>
  <div class="site-foot__bar">
    <span class="bx-label">MIT &middot; &copy; Avestura</span>
    <span class="bx-label">UI styled with Boxy &middot; radius 0 &middot; 4px grid</span>
  </div>
</footer>
<script src="${R}assets/zagros-search.js"></script>
<script src="${R}assets/zagros.js"></script>
</body>
</html>
`;
}

function tocHtml(headings) {
  const hs = headings.filter((h) => h.level === 2 || h.level === 3);
  if (hs.length < 2) return '';
  return `<p class="bx-label" style="margin-block-end: 12px">On this page</p><nav class="bx-toc" aria-label="On this page">${hs
    .map((h) => `<a${h.level === 3 ? ' class="sub"' : ''} href="#${h.id}">${esc(h.text)}</a>`).join('')}</nav>`;
}

function sidebar(currentOut, R) {
  const groups = [];
  for (const p of docPages) {
    let g = groups.find((x) => x.name === p.group);
    if (!g) groups.push((g = { name: p.group, items: [] }));
    g.items.push(p);
  }
  return groups.map((g) => `<h4 class="bx-label">${esc(g.name)}</h4><ul>${g.items.map((p) =>
    `<li><a href="${R}${p.out}"${p.out === currentOut ? ' aria-current="page"' : ''}>${esc(p.title)}</a></li>`).join('')}</ul>`).join('');
}

function docLayout(out, title, desc, section, side, inner, toc, edit) {
  const R = rootOf(out);
  return shell({
    out, title, desc, section,
    body: `<div class="doc-shell">
<aside class="doc-side" aria-label="Documentation">${side}</aside>
<article class="doc-main"><div class="bx-prose">${inner}</div>${edit || ''}</article>
<aside class="doc-toc" aria-label="Table of contents">${toc}</aside>
</div>`,
  });
}

const docOuts = new Set(docPages.map((p) => p.out));
const write = (rel, content) => {
  if (rel.endsWith('.html')) {
    content = content.replace(/href="((?:\.\.\/)*)docs\/([a-z0-9-]+)\.html((?:#[^"]*)?)"/g, (m, up, slug, h) => (docOuts.has(`docs/${slug}.html`) ? m : `href="${REPO}/tree/main/docs"`));
  }
  const p = path.join(OUT, rel);
  fs.mkdirSync(path.dirname(p), { recursive: true });
  fs.writeFileSync(p, content);
};

const statusTag = (s) => {
  const v = (s || 'unknown').toLowerCase();
  const cls = { accepted: 'success', shipped: 'success', done: 'success', 'in-progress': 'accent', partial: 'warning', proposed: 'warning', planned: '', superseded: 'danger', blocked: 'danger' }[v] ?? '';
  return `<span class="bx-tag${cls ? ' bx-tag--' + cls : ''}">${esc(v)}</span>`;
};

// ------------------------------------------------------------------ build
fs.rmSync(OUT, { recursive: true, force: true });
fs.mkdirSync(OUT, { recursive: true });
fs.cpSync(path.join(ROOT, 'site/assets'), path.join(OUT, 'assets'), { recursive: true });

write('favicon.svg', `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16"><style>path{stroke:#6f6f6f}@media (prefers-color-scheme:dark){path{stroke:#f4f4f4}}@media (prefers-color-scheme:light){path{stroke:#161616}}</style><path d="M1 13L4.5 7L6.5 10L9 3L11.5 8L13 6L15 13" fill="none" stroke-width="1.75" stroke-linecap="square" stroke-linejoin="miter"/></svg>
`);

// registry.json: use a generated one when present, else an empty valid registry.
for (const c of ['site/registry.json', 'registry.json', 'site-dist-extra/registry.json']) {
  const p = path.join(ROOT, c);
  if (fs.existsSync(p)) { fs.copyFileSync(p, path.join(OUT, 'registry.json')); break; }
}
if (!fs.existsSync(path.join(OUT, 'registry.json'))) {
  warn('no registry.json found; writing an empty registry');
  write('registry.json', JSON.stringify({ schema: 1, generated: new Date().toISOString(), extensions: [] }, null, 2) + '\n');
}

// docs pages
docPages.forEach((p, idx) => {
  const prev = docPages[idx - 1], next = docPages[idx + 1];
  const R = rootOf(p.out);
  const pager = (prev || next) ? `<nav class="doc-pager" aria-label="Pager">${prev ? `<a href="${R}${prev.out}"><span class="bx-label">Previous</span>${esc(prev.title)}</a>` : '<span></span>'}${next ? `<a href="${R}${next.out}"><span class="bx-label">Next</span>${esc(next.title)}</a>` : ''}</nav>` : '';
  const edit = `<p class="bx-label mt-6"><a href="${REPO}/edit/main/${p.src}">Edit this page on GitHub</a></p>`;
  const inner = `<p class="bx-label">${esc(p.group)}</p><h1>${esc(p.pageTitle)}</h1>${p.html}`;
  write(p.out, docLayout(p.out, p.pageTitle, p.desc, 'docs', sidebar(p.out, R), inner, tocHtml(p.headings), pager + edit));
});
if (!docPages.length) warn('no docs pages were built');

// decisions
{
  const R = '';
  const items = adrs.map((p) => `<a class="list-row" href="${p.out}"><span class="bx-mono subtle">${esc(p.num)}</span><span><b>${esc((p.pageTitle || '').replace(/^ADR\s*\d+:\s*/i, ''))}</b><br><span class="muted" style="font-size: var(--bx-text-base)">${esc(p.desc)}</span></span>${statusTag(p.data.status)}</a>`).join('');
  const intro = adrReadme ? render(parseFrontMatter(adrReadme).body, { resolveLink: makeResolver('docs/adr/README.md', 'decisions.html') }).html : '';
  write('decisions.html', shell({
    out: 'decisions.html', title: 'Decisions', section: 'decisions', mode: 'editorial',
    desc: 'Architecture decision records for Zagros: why it is built the way it is.',
    body: `<section class="sec"><div class="sec__head"><p class="bx-label">Architecture decision records</p><h1>Decisions</h1><div class="bx-prose" style="max-inline-size: 64ch">${intro}</div></div>
${adrs.length ? `<div class="list-rows">${items}</div>` : '<div class="bx-empty"><h3>No decisions recorded yet</h3></div>'}</section>`,
  }));
  for (const p of adrs) {
    const R2 = rootOf(p.out);
    const side = `<h4 class="bx-label">Decisions</h4><ul>${adrs.map((a) => `<li><a href="${R2}${a.out}"${a.out === p.out ? ' aria-current="page"' : ''}>${esc(a.num)} ${esc((a.pageTitle || '').replace(/^ADR\s*\d+:\s*/i, ''))}</a></li>`).join('')}</ul><p style="margin-block-start: 16px"><a href="${R2}decisions.html">All decisions</a></p>`;
    const inner = `<div class="doc-meta">${statusTag(p.data.status)}${p.data.date ? `<span class="bx-label">${esc(p.data.date)}</span>` : ''}</div><h1>${esc(p.pageTitle)}</h1>${p.html}`;
    write(p.out, docLayout(p.out, p.pageTitle, p.desc, 'decisions', side, inner, tocHtml(p.headings), `<p class="bx-label mt-6"><a href="${REPO}/edit/main/${p.src}">Edit this page on GitHub</a></p>`));
  }
}

// roadmap
{
  const counts = {};
  streams.forEach((s) => { const k = (s.data.status || 'unknown').toLowerCase(); counts[k] = (counts[k] || 0) + 1; });
  const summary = Object.entries(counts).map(([k, n]) => `<span>${statusTag(k)} <span class="bx-mono">${n}</span></span>`).join(' ');
  const rows = streams.map((s) => {
    const tasks = (s.body.match(/^\s*- \[[ xX]\]/gm) || []);
    const done = tasks.filter((t) => /x/i.test(t)).length;
    const pct = tasks.length ? Math.round((done / tasks.length) * 100) : 0;
    return `<tr><td><a href="#${s.id}">${esc(s.pageTitle || s.id)}</a></td><td>${statusTag(s.data.status)}</td><td class="num bx-mono">${tasks.length ? `${done}/${tasks.length}` : '-'}</td><td style="min-inline-size: 120px"><div class="bx-progress" role="progressbar" aria-valuenow="${pct}" aria-valuemin="0" aria-valuemax="100"><i style="inline-size: ${pct}%"></i></div></td></tr>`;
  }).join('');
  const intro = wsReadme ? render(parseFrontMatter(wsReadme).body, { resolveLink: makeResolver('docs/workstreams/README.md', 'roadmap.html') }).html : '';
  const sections = streams.map((s) => {
    const adrLinks = (s.data.adrs || '').split(/[,\s]+/).filter(Boolean).map((n) => {
      const a = adrs.find((x) => x.num === n.padStart(4, '0'));
      return a ? `<a class="bx-tag" href="${a.out}">ADR ${esc(n)}</a>` : `<span class="bx-tag">ADR ${esc(n)}</span>`;
    }).join(' ');
    return `<section id="${s.id}" class="bx-card" style="margin-block-end: 24px"><div class="bx-card__head"><div><p class="bx-label">${esc(s.id)}</p><h2 style="font-size: var(--bx-text-xl)">${esc(s.pageTitle)}</h2></div>${statusTag(s.data.status)}</div><div class="bx-card__body bx-prose">${s.html}${adrLinks ? `<p>${adrLinks}</p>` : ''}</div></section>`;
  }).join('');
  write('roadmap.html', shell({
    out: 'roadmap.html', title: 'Roadmap', section: 'roadmap', mode: 'editorial',
    desc: 'Delivery status of Zagros workstreams.',
    body: `<section class="sec"><div class="sec__head"><p class="bx-label">Delivery</p><h1>Roadmap</h1><div class="bx-prose" style="max-inline-size: 64ch">${intro}</div></div>
${streams.length ? `<p class="row" style="gap: 16px; margin-block-end: 24px">${summary}</p>
<div class="bx-table-wrap" style="margin-block-end: 48px"><table class="bx-table"><thead><tr><th>Workstream</th><th>Status</th><th class="num">Tasks</th><th>Progress</th></tr></thead><tbody>${rows}</tbody></table></div>${sections}` : '<div class="bx-empty"><h3>No workstreams tracked yet</h3></div>'}</section>`,
  }));
}

// marketplace
write('marketplace.html', shell({
  out: 'marketplace.html', title: 'Marketplace', section: 'marketplace', mode: 'blueprint',
  desc: 'Browse Zagros extensions. Each one lists the permissions it asks for before you install it.',
  body: `<section class="sec">
<div class="sec__head"><p class="bx-label">Extensions</p><h1>Marketplace</h1>
<p>A Git-based registry: no accounts, no server. Every extension lists the permissions it requests. Install from the command line with the id, or straight from GitHub with <code>github:owner/repo</code>. <a href="docs/marketplace.html">How to publish</a>.</p></div>
<div class="mk-tools"><input id="mk-q" class="bx-input" type="search" placeholder="Search extensions" aria-label="Search extensions"><span id="mk-count" class="bx-label" aria-live="polite"></span></div>
<div id="mk-tags" class="mk-tags" style="margin-block-end: 24px" aria-label="Filter by tag"></div>
<noscript><p class="bx-callout bx-callout--warning">The marketplace list needs JavaScript. You can also read <a href="registry.json">registry.json</a> directly.</p></noscript>
<div id="mk-grid" class="card-grid"><div class="bx-empty" style="grid-column: 1 / -1"><p class="bx-label">Loading</p><h3>Fetching registry.json</h3></div></div>
</section>`,
}));

// landing
write('index.html', shell({ out: 'index.html', title: 'Zagros', mode: 'editorial', section: '', body: landing(), desc: undefined }));

// 404
write('404.html', shell({
  out: '404.html', title: 'Not found', noindex: true,
  body: `<section class="sec"><p class="bx-label">404</p><h1>That page does not exist</h1><p class="lede mt-4">It may have moved. Try the <a href="/zagros/docs/getting-started.html">documentation</a> or go back to the <a href="/zagros/">home page</a>.</p></section>`,
}));

// search index, sitemap, robots
{
  const idx = [];
  const add = (p, label) => {
    const t = p.pageTitle || p.title;
    // split by headings: simple approach, one entry per page plus one per h2/h3
    const g = p.kind === 'adr' ? 'Decisions' : p.group || 'Docs';
    idx.push({ t, h: '', u: p.out, g, x: ((p.desc || '') + ' ' + (p.text || '')).replace(/\s+/g, ' ').slice(0, 1500) });
    for (const h of p.headings.filter((h) => h.level <= 3)) idx.push({ t, h: h.text, u: `${p.out}#${h.id}`, g, x: sectionText(p, h) });
  };
  const sectionText = (p, h) => {
    const m = new RegExp(`<h${h.level} id="${h.id}">[\\s\\S]*?(?=<h[1-3] id=|$)`).exec(p.html);
    return m ? m[0].replace(/<[^>]+>/g, ' ').replace(/&[a-z]+;/g, ' ').replace(/\s+/g, ' ').slice(0, 1000) : '';
  };
  docPages.forEach(add);
  adrs.forEach(add);
  write('search-index.json', JSON.stringify(idx));
  // Same data as a script, so search also works when pages are opened from disk (file://), where fetch() is blocked.
  write('search-index.js', `window.__ZAGROS_INDEX=${JSON.stringify(idx)};`);
  const urls = ['', 'marketplace.html', 'decisions.html', 'roadmap.html', ...docPages.map((p) => p.out), ...adrs.map((p) => p.out)];
  write('sitemap.xml', `<?xml version="1.0" encoding="UTF-8"?>\n<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">\n${urls.map((u) => `  <url><loc>${SITE_URL}/${u}</loc></url>`).join('\n')}\n</urlset>\n`);
  write('robots.txt', `User-agent: *\nAllow: /\n\nSitemap: ${SITE_URL}/sitemap.xml\n`);
  write('.nojekyll', '');
}

console.log(`site built: ${docPages.length} docs, ${adrs.length} decisions, ${streams.length} workstreams -> ${path.relative(ROOT, OUT)}`);

// ----------------------------------------------------------------- landing
function landing() {
  const icon = (n) => `<svg class="bx-icon bx-icon--lg" aria-hidden="true"><use href="#${n}"/></svg>`;
  const feats = [
    ['i-bolt', 'Fuzzy and fast', 'Type a few letters of an app, file or command. Results are ranked by match quality and by what you actually pick (frecency).'],
    ['i-search', 'Everything search', 'Instant file search through voidtools Everything when it is running. No indexer of our own.'],
    ['i-chart', 'Calculator and units', 'Type <code>2*(3+4)</code> or <code>10 km in mi</code> and copy the answer with Enter.'],
    ['i-lock', 'Password generator', 'Generate strong passwords and passphrases locally. Nothing leaves your machine.'],
    ['i-edit', 'Text tools', 'Case conversion, encode and decode, hashes, UUIDs, timestamps and more on any text.'],
    ['i-clip', 'Clipboard history', 'Search what you copied earlier and paste it back.'],
    ['i-columns', 'Window switcher', 'Jump to any open window by title.'],
    ['i-globe', 'Web search', 'Keyword searches such as <code>g query</code>. Add your own in the config file.'],
    ['i-star', 'Emoji', 'Find an emoji by name and copy it.'],
    ['i-term', 'Starlark extensions', 'Write extensions in a small, safe Python-like language. No compile step, hot reload.'],
    ['i-gear', 'Themes in TOML', 'A theme is one readable file. Switch themes without restarting.'],
    ['i-branch', 'Git marketplace', 'Install with <code>zagros ext install github:owner/repo</code>. The registry is a Git repository.'],
  ];
  const providers = [
    ['Applications', 'Start menu, desktop and extra scan folders'], ['Everything', 'File search through voidtools Everything'],
    ['Calculator', 'Expressions, functions, hex and binary'], ['Unit converter', 'Length, mass, temperature, data, time and more'],
    ['Passwords', 'Random passwords and passphrases'], ['Text tools', 'Transform, encode, hash'],
    ['Clipboard', 'Searchable clipboard history'], ['Windows', 'Switch between open windows'],
    ['Web search', 'Configurable keyword searches'], ['Emoji', 'Search by name'],
    ['Aliases', 'Your own keywords for paths, URLs and commands'], ['Extensions', 'Everything you install'],
  ];
  const perms = [
    ['network:api.github.com', 'Connect to api.github.com only', ''],
    ['clipboard.write', 'Put text on your clipboard', ''],
    ['fs.read:~/Documents', 'Read files under Documents', ''],
    ['process.run:git', 'Start the program git', 'warn'],
  ];
  return `
<section class="hero">
  <div class="hero__grid" aria-hidden="true"></div>
  <div class="hero-split">
    <div>
      <p class="bx-label">01 / Keyboard launcher for Windows</p>
      <h1>Type less. Launch more.</h1>
      <p class="hero__sub lede mt-5">Zagros opens with one hotkey and finds apps, files, windows, clipboard entries and answers as you type. It is a single native executable, and it is extended with small sandboxed scripts that ask for permission first.</p>
      <div class="cta-row">
        <a class="bx-btn bx-btn--primary bx-btn--lg" href="${REPO}/releases/latest"><svg class="bx-icon" aria-hidden="true"><use href="#i-download"/></svg>Download for Windows</a>
        <a class="bx-btn bx-btn--lg" href="docs/getting-started.html">Read the docs</a>
      </div>
      <p class="bx-label mt-5">MIT &middot; Go &middot; native Win32 &middot; amd64 and arm64</p>
    </div>
    <div class="launcher" role="img" aria-label="Mock of the Zagros launcher window searching for fire">
      <div class="launcher__input"><svg class="bx-icon" aria-hidden="true"><use href="#i-search"/></svg><span>fire</span><span class="launcher__caret" aria-hidden="true"></span></div>
      <div class="launcher__row" aria-selected="true"><span class="launcher__icon">Fx</span><span>Firefox<small>C:\\Program Files\\Mozilla Firefox</small></span><span class="bx-tag bx-tag--sm">App</span></div>
      <div class="launcher__row"><span class="launcher__icon">Fe</span><span>Fire Emoji<small>Copy to clipboard</small></span><span class="bx-tag bx-tag--sm">Emoji</span></div>
      <div class="launcher__row"><span class="launcher__icon">Fw</span><span>firewall-rules.txt<small>D:\\Notes</small></span><span class="bx-tag bx-tag--sm">File</span></div>
      <div class="launcher__row"><span class="launcher__icon">Wn</span><span>Windows Defender Firewall<small>Open window</small></span><span class="bx-tag bx-tag--sm">Window</span></div>
      <div class="launcher__foot"><span class="bx-label">Enter open</span><span class="bx-label">Tab actions</span><span class="bx-label">Esc close</span></div>
    </div>
  </div>
</section>

<section class="sec">
  <div class="sec__head"><p class="bx-label">02 / What it does</p><h2>Everything you reach for, from the keyboard</h2>
  <p>Built-in providers cover the daily basics. Everything else is an extension away.</p></div>
  <div class="bx-collapse bx-collapse--3">
    ${feats.map(([i, t, d]) => `<div class="feat">${icon(i)}<h3>${t}</h3><p>${d}</p></div>`).join('\n    ')}
  </div>
</section>

<section class="sec">
  <div class="sec__head"><p class="bx-label">03 / Built in</p><h2>Providers that ship in the box</h2>
  <p>Each provider can be turned off in the config. <a href="docs/builtin-providers.html">Read the reference.</a></p></div>
  <div class="bx-collapse bx-collapse--4">
    ${providers.map(([n, d]) => `<div class="feat"><h3>${n}</h3><p>${d}</p></div>`).join('\n    ')}
  </div>
</section>

<section class="sec">
  <div class="two-col">
    <div>
      <p class="bx-label">04 / Extensions and permissions</p>
      <h2 class="mt-2">Extensions can only do what you allow</h2>
      <p class="muted mt-4">Extensions are Starlark scripts with a manifest. By default they can return results and nothing else. Network, files, clipboard and processes each need a permission, declared in the manifest and approved by you at install, much like on Android. The host checks the grant on every call, and an update that asks for more asks you again.</p>
      <ul class="stack--tight mt-5" style="list-style: none; padding: 0; display: grid; font-size: var(--bx-text-base)">
        <li>Scoped permissions such as <code>network:api.github.com</code></li>
        <li>Review and revoke grants at any time</li>
        <li>Install from the registry or any <code>github:owner/repo</code></li>
      </ul>
      <p class="mt-5"><a class="bx-btn" href="docs/permissions.html">Permission model</a> <a class="bx-btn bx-btn--ghost" href="docs/extensions.html">Write an extension</a></p>
    </div>
    <div>
      <div class="bx-dialog consent" role="img" aria-label="Mock permission consent dialog">
        <div style="padding: 16px; border-block-end: 1px solid var(--bx-line)"><p class="bx-label">Install extension</p><h3 style="font-size: var(--bx-text-lg)">GitHub Notifications</h3><p class="muted" style="font-size: var(--bx-text-sm)">octo/zagros-notifications &middot; v1.2.0</p></div>
        <div style="padding: 12px 16px" class="bx-label">This extension asks to</div>
        ${perms.map(([k, d, w]) => `<div class="perm"><svg class="bx-icon" aria-hidden="true"><use href="#${w ? 'i-warning' : 'i-check'}"/></svg><span>${d}<small class="bx-mono">${k}</small></span><span class="bx-tag bx-tag--sm${w ? ' bx-tag--warning' : ''}">${w ? 'sensitive' : 'allow'}</span></div>`).join('')}
        <div style="display: flex; justify-content: flex-end; gap: 8px; padding: 12px 16px"><span class="bx-btn bx-btn--sm">Cancel</span><span class="bx-btn bx-btn--sm bx-btn--primary">Install</span></div>
      </div>
      <p class="bx-label mt-4">Illustration. The real dialog is drawn by the launcher.</p>
    </div>
  </div>
</section>

<section class="sec">
  <div class="two-col">
    <div>
      <p class="bx-label">05 / Extension example</p>
      <h2 class="mt-2">Declared in a manifest</h2>
      <p class="muted mt-4">No SDK, no build step. Drop the folder into your extensions directory or publish it to GitHub. See the <a href="docs/extensions.html">extensions guide</a> for the full host API.</p>
    </div>
    <div class="bx-codeblock"><div class="bx-codeblock__head"><span class="bx-label">manifest.toml</span><span class="bx-codeblock__spacer"></span><button class="bx-codeblock__btn" type="button" data-copy="self">copy</button></div><pre><code><span class="ln">id = "me/hello"</span><span class="ln">name = "Hello"</span><span class="ln">version = "0.1.0"</span><span class="ln">keywords = ["hi"]</span><span class="ln">permissions = ["clipboard.write"]</span></code></pre></div>
  </div>
</section>

<section class="sec">
  <div class="sec__head"><p class="bx-label">06 / Themes</p><h2>Square by default, yours if you want</h2>
  <p>Themes are TOML files: colours, fonts, row height, border width, corner radius. The built-in themes follow the Boxy design system, and Zagros follows the Windows light and dark setting unless you pick one. <a href="docs/themes.html">Theme reference</a>.</p></div>
  <div class="theme-strip">
    <div class="feat"><p class="bx-label">Boxy Light</p><div class="launcher mt-4" style="box-shadow: none"><div class="launcher__input"><span>calc 2^10</span></div><div class="launcher__row" aria-selected="true"><span class="launcher__icon">=</span><span>1024<small>Calculator</small></span><span></span></div></div></div>
    <div class="feat" data-theme="dark" style="background: var(--bx-canvas); color: var(--bx-ink)"><p class="bx-label">Boxy Dark</p><div class="launcher mt-4" style="box-shadow: none"><div class="launcher__input"><span>10 km in mi</span></div><div class="launcher__row" aria-selected="true"><span class="launcher__icon">=</span><span>6.2137 mi<small>Unit converter</small></span><span></span></div></div></div>
    <div class="feat"><p class="bx-label">Your theme</p><div class="bx-codeblock mt-4"><pre><code><span class="ln">name = "Mine"</span><span class="ln">radius = 0</span><span class="ln">[colors]</span><span class="ln">show_icons = true</span></code></pre></div></div>
  </div>
</section>

<section class="sec">
  <div class="credit">
    <p class="bx-label">Inspired by Keypirinha</p>
    <h2 style="font-size: var(--bx-text-2xl)">Zagros would not exist without Keypirinha</h2>
    <p class="muted" style="max-inline-size: 68ch"><a href="https://keypirinha.com/">Keypirinha</a>, by Jean-Charles Lefebvre, is a fast, keyboard-driven launcher for Windows with a package system that many of us have relied on for years. Zagros borrows its spirit: a small window, fuzzy search, extensions you can read. It is an independent project with its own code, and it is not affiliated with Keypirinha.</p>
    <p><a class="bx-btn" href="https://keypirinha.com/">Visit keypirinha.com</a></p>
  </div>
</section>

<section class="sec" style="border-block-end: 0">
  <div class="sec__head"><p class="bx-label">07 / Get it</p><h2>Download Zagros</h2>
  <p>Grab the zip for your CPU from the latest release, unzip, run <code>zagros.exe</code>, press <span class="bx-kbd">Alt</span> + <span class="bx-kbd">Space</span>.</p></div>
  <div class="cta-row" style="margin-block-start: 0"><a class="bx-btn bx-btn--primary bx-btn--lg" href="${REPO}/releases/latest">Download from GitHub releases</a><a class="bx-btn bx-btn--lg" href="${REPO}">View source</a></div>
</section>`;
}
