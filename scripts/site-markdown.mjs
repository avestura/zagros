// Small dependency-free Markdown to HTML converter for the Zagros site.
// Covers what the docs use: headings, paragraphs, fenced code, lists (nested),
// tables, blockquotes (rendered as Boxy callouts), rules, inline code, bold,
// italic, links, images and raw HTML blocks.

export function escapeHtml(s) {
  return String(s).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}

export function slugify(s) {
  return String(s).toLowerCase().replace(/<[^>]+>/g, '').replace(/[^a-z0-9\s-]/g, '').trim().replace(/\s+/g, '-') || 'section';
}

export function parseFrontMatter(src) {
  const m = /^---\r?\n([\s\S]*?)\r?\n---\r?\n?/.exec(src);
  if (!m) return { data: {}, body: src };
  const data = {};
  for (const line of m[1].split(/\r?\n/)) {
    const kv = /^([A-Za-z0-9_-]+)\s*:\s*(.*)$/.exec(line);
    if (kv) data[kv[1]] = kv[2].trim().replace(/^["']|["']$/g, '');
  }
  return { data, body: src.slice(m[0].length) };
}

const ICONS = { info: 'i-info', success: 'i-success', warning: 'i-warning', danger: 'i-error' };

/**
 * @param {string} src markdown (front matter already removed)
 * @param {{ resolveLink?: (href: string) => string }} opts
 * @returns {{ html: string, headings: {level:number,text:string,id:string}[], title: string|null, text: string }}
 */
export function render(src, opts = {}) {
  const resolveLink = opts.resolveLink || ((h) => h);
  const headings = [];
  const used = new Map();
  let title = null;
  const plain = [];

  function uniqueId(text) {
    const base = slugify(text);
    const n = used.get(base) || 0;
    used.set(base, n + 1);
    return n ? `${base}-${n}` : base;
  }

  function inline(text) {
    const stash = [];
    const keep = (html) => `\u0000${stash.push(html) - 1}\u0000`;
    // code spans
    text = text.replace(/(`+)([\s\S]*?[^`])\1(?!`)/g, (_, __, code) => keep(`<code>${escapeHtml(code.trim())}</code>`));
    // images
    text = text.replace(/!\[([^\]]*)\]\(([^)\s]+)(?:\s+"[^"]*")?\)/g, (_, alt, url) =>
      keep(`<img src="${escapeHtml(url)}" alt="${escapeHtml(alt)}" loading="lazy">`));
    // links
    text = text.replace(/\[([^\]]+)\]\(([^)\s]+)(?:\s+"[^"]*")?\)/g, (_, label, url) => {
      const href = resolveLink(url);
      const ext = /^https?:/.test(href) ? ' rel="noopener"' : '';
      return keep(`<a href="${escapeHtml(href)}"${ext}>${inlineNoLink(label)}</a>`);
    });
    // autolinks
    text = text.replace(/<(https?:\/\/[^>\s]+)>/g, (_, url) => keep(`<a href="${escapeHtml(url)}" rel="noopener">${escapeHtml(url)}</a>`));
    // inline html tags we allow through
    text = text.replace(/<(\/?(?:kbd|br|sup|sub|b|i|em|strong|code|span))(\s[^>]*)?>/g, (m) => keep(m));
    text = escapeHtml(text);
    text = text.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>').replace(/__([^_]+)__/g, '<strong>$1</strong>');
    text = text.replace(/(^|[^*\w])\*([^*\s][^*]*)\*(?!\w)/g, '$1<em>$2</em>').replace(/(^|[^_\w])_([^_\s][^_]*)_(?![\w])/g, '$1<em>$2</em>');
    text = text.replace(/~~([^~]+)~~/g, '<del>$1</del>');
    return text.replace(/\u0000(\d+)\u0000/g, (_, i) => stash[+i]);
  }
  function inlineNoLink(t) {
    const stash = [];
    t = t.replace(/`([^`]+)`/g, (_, c) => `\u0000${stash.push(`<code>${escapeHtml(c)}</code>`) - 1}\u0000`);
    t = escapeHtml(t).replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');
    return t.replace(/\u0000(\d+)\u0000/g, (_, i) => stash[+i]);
  }
  const stripMd = (t) => t.replace(/!?\[([^\]]*)\]\([^)]*\)/g, '$1').replace(/[`*_~]/g, '');

  const lines = src.replace(/\r\n?/g, '\n').split('\n');
  const out = [];

  const isTableSep = (l) => /^\s*\|?\s*:?-{2,}:?\s*(\|\s*:?-{2,}:?\s*)*\|?\s*$/.test(l) && l.includes('-') && (l.includes('|') || /^\s*:?-+:?\s*$/.test(l));
  const splitRow = (l) => {
    let s = l.trim();
    if (s.startsWith('|')) s = s.slice(1);
    if (s.endsWith('|') && !s.endsWith('\\|')) s = s.slice(0, -1);
    return s.split(/(?<!\\)\|/).map((c) => c.trim().replace(/\\\|/g, '|'));
  };
  const listRe = /^(\s*)([-*+]|\d+[.)])\s+(.*)$/;

  function parseList(start, baseIndent) {
    // returns [html, nextIndex]
    const first = listRe.exec(lines[start]);
    const ordered = /\d/.test(first[2]);
    const tag = ordered ? 'ol' : 'ul';
    let i = start;
    const items = [];
    while (i < lines.length) {
      const m = listRe.exec(lines[i]);
      if (!m || m[1].length < baseIndent) break;
      if (m[1].length > baseIndent) break; // handled as nested below
      let body = [m[3]];
      let nested = '';
      i++;
      while (i < lines.length) {
        const l = lines[i];
        if (l.trim() === '') {
          // blank line: continue only if next non-blank is indented deeper
          let j = i + 1;
          while (j < lines.length && lines[j].trim() === '') j++;
          if (j < lines.length && /^\s+/.test(lines[j]) && lines[j].search(/\S/) > baseIndent) { i = j; continue; }
          break;
        }
        const nm = listRe.exec(l);
        if (nm && nm[1].length > baseIndent) {
          const [h, ni] = parseList(i, nm[1].length);
          nested += h;
          i = ni;
          continue;
        }
        if (nm && nm[1].length <= baseIndent) break;
        if (l.search(/\S/) > baseIndent) { body.push(l.trim()); i++; continue; }
        break;
      }
      items.push(`<li>${inline(body.join(' '))}${nested}</li>`);
    }
    return [`<${tag}>${items.join('')}</${tag}>`, i];
  }

  let i = 0;
  while (i < lines.length) {
    const line = lines[i];
    if (line.trim() === '') { i++; continue; }

    // fenced code
    let m = /^(\s*)(```+|~~~+)\s*([\w+#.-]*)[^\n]*$/.exec(line);
    if (m) {
      const fence = m[2];
      const lang = m[3] || '';
      const buf = [];
      i++;
      while (i < lines.length && !lines[i].trim().startsWith(fence)) { buf.push(lines[i].slice(Math.min(m[1].length, lines[i].search(/\S|$/)))); i++; }
      i++;
      const code = buf.join('\n');
      plain.push(code);
      const rows = buf.map((l) => `<span class="ln">${escapeHtml(l) || ' '}</span>`).join('');
      out.push(`<!-- boxy-ignore-start: documentation code sample -->
<div class="bx-codeblock"><div class="bx-codeblock__head"><span class="bx-label">${escapeHtml(lang || 'text')}</span><span class="bx-codeblock__spacer"></span><button class="bx-codeblock__btn" type="button" data-copy="self">copy</button></div><pre><code>${rows}</code></pre></div>
<!-- boxy-ignore-end -->`);
      continue;
    }

    // heading
    m = /^(#{1,6})\s+(.*?)\s*#*\s*$/.exec(line);
    if (m) {
      const level = m[1].length;
      const raw = stripMd(m[2]);
      if (level === 1 && title === null) { title = raw; i++; continue; }
      const id = uniqueId(raw);
      headings.push({ level, text: raw, id });
      plain.push(raw);
      const anchor = level <= 3 ? ` <a class="anchor" href="#${id}" aria-label="Link to this section">#</a>` : '';
      out.push(`<h${level} id="${id}">${inline(m[2])}${anchor}</h${level}>`);
      i++;
      continue;
    }

    // hr
    if (/^\s*([-*_])(\s*\1){2,}\s*$/.test(line)) { out.push('<hr>'); i++; continue; }

    // table
    if (line.includes('|') && i + 1 < lines.length && isTableSep(lines[i + 1])) {
      const head = splitRow(line);
      const aligns = splitRow(lines[i + 1]).map((c) => (c.startsWith(':') && c.endsWith(':') ? 'center' : c.endsWith(':') ? 'right' : ''));
      i += 2;
      const rows = [];
      while (i < lines.length && lines[i].trim() !== '' && lines[i].includes('|')) { rows.push(splitRow(lines[i])); i++; }
      const al = (k) => (aligns[k] ? ` style="text-align: ${aligns[k]}"` : '');
      const th = head.map((c, k) => `<th${al(k)}>${inline(c)}</th>`).join('');
      const trs = rows.map((r) => `<tr>${head.map((_, k) => `<td${al(k)}>${inline(r[k] ?? '')}</td>`).join('')}</tr>`).join('');
      rows.forEach((r) => plain.push(r.join(' ')));
      out.push(`<div class="bx-table-wrap"><table class="bx-table"><thead><tr>${th}</tr></thead><tbody>${trs}</tbody></table></div>`);
      continue;
    }

    // blockquote -> callout
    if (/^\s*>/.test(line)) {
      const buf = [];
      while (i < lines.length && /^\s*>/.test(lines[i])) { buf.push(lines[i].replace(/^\s*>\s?/, '')); i++; }
      let kind = 'info';
      let txt = buf.join('\n');
      const k = /^\s*(?:\*\*)?(Note|Tip|Info|Warning|Caution|Danger|Important)(?:\*\*)?\s*[:.]?\s*(?:\*\*)?/i.exec(txt);
      if (k) {
        const w = k[1].toLowerCase();
        kind = w === 'warning' || w === 'caution' || w === 'important' ? 'warning' : w === 'danger' ? 'danger' : w === 'tip' ? 'success' : 'info';
      }
      const inner = render(txt, opts);
      plain.push(inner.text);
      out.push(`<div class="bx-callout bx-callout--${kind}" role="note"><svg class="bx-icon" aria-hidden="true"><use href="#${ICONS[kind]}"/></svg><div>${inner.html}</div></div>`);
      continue;
    }

    // list
    m = listRe.exec(line);
    if (m) {
      const [h, ni] = parseList(i, m[1].length);
      out.push(h);
      i = ni;
      continue;
    }

    // raw html block
    if (/^\s*<(div|p|details|summary|table|section|figure|img|a |br|!--)/i.test(line)) {
      const buf = [];
      while (i < lines.length && lines[i].trim() !== '') { buf.push(lines[i]); i++; }
      out.push(buf.join('\n'));
      continue;
    }

    // paragraph
    const buf = [];
    while (i < lines.length && lines[i].trim() !== '' &&
      !/^(#{1,6}\s|```|~~~|\s*>|\s*([-*_])(\s*\2){2,}\s*$)/.test(lines[i]) &&
      !(listRe.test(lines[i]) && buf.length) &&
      !(lines[i].includes('|') && i + 1 < lines.length && isTableSep(lines[i + 1]))) {
      buf.push(lines[i].trim());
      i++;
    }
    if (!buf.length) { i++; continue; }
    plain.push(stripMd(buf.join(' ')));
    out.push(`<p>${inline(buf.join(' '))}</p>`);
  }

  return { html: out.join('\n'), headings, title, text: plain.join(' ') };
}
