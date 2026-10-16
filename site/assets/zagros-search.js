/* Zagros docs search core. Pure functions, no DOM: loaded by the site and by scripts/site-search-test.mjs.
   Index entry: { t: page title, h: heading ('' for the page itself), u: url, g: section label, x: text }. */
(function (root, factory) {
  var api = factory();
  if (typeof module === 'object' && module.exports) module.exports = api;
  else root.ZagrosSearch = api;
})(typeof self !== 'undefined' ? self : this, function () {
  'use strict';

  var MAX_RESULTS = 8;

  function tokenize(s) {
    return String(s || '').toLowerCase().split(/[^a-z0-9À-￿]+/).filter(Boolean);
  }

  function prefixHit(tokens, term) {
    for (var i = 0; i < tokens.length; i++) if (tokens[i].indexOf(term) === 0) return tokens[i];
    return null;
  }

  function prepare(d) {
    if (!d._p) d._p = { t: tokenize(d.t), h: tokenize(d.h), x: tokenize(d.x) };
    return d._p;
  }

  /* Returns up to MAX_RESULTS hits, best first: { d, score, terms }.
     Every term must match (AND). A term is a lowercase token prefix. One-letter terms
     only look at titles and headings; terms of two or more letters also search body text. */
  function search(index, query) {
    var terms = tokenize(query);
    if (!terms.length || !index || !index.length) return [];
    var hits = [];
    for (var n = 0; n < index.length; n++) {
      var d = index[n], p = prepare(d), score = 0, ok = true;
      for (var i = 0; i < terms.length && ok; i++) {
        var w = terms[i], s = 0, m;
        if ((m = prefixHit(p.t, w))) s = m === w ? 16 : 10;
        else if ((m = prefixHit(p.h, w))) s = m === w ? 10 : 6;
        else if (w.length >= 2 && prefixHit(p.x, w)) s = 1;
        if (!s) ok = false; else score += s;
      }
      if (!ok) continue;
      if (!d.h) score += 2; // the page itself outranks its sections on a tie
      hits.push({ d: d, score: score, terms: terms, n: n });
    }
    hits.sort(function (a, b) { return b.score - a.score || a.n - b.n; });
    return hits.slice(0, MAX_RESULTS);
  }

  /* Group ranked hits by section label, keeping the order in which each group first appears. */
  function group(hits) {
    var out = [], by = {};
    hits.forEach(function (h) {
      var g = h.d.g || 'Docs';
      if (!by[g]) { by[g] = { label: g, hits: [] }; out.push(by[g]); }
      by[g].hits.push(h);
    });
    return out;
  }

  /* Split text into [{ text, hit }] segments, marking token prefixes that match a term. */
  function segments(text, terms) {
    text = String(text || '');
    var re = /[a-z0-9À-￿]+/gi, out = [], last = 0, m;
    while ((m = re.exec(text))) {
      var low = m[0].toLowerCase(), hit = false;
      for (var i = 0; i < terms.length; i++) if (low.indexOf(terms[i]) === 0) { hit = true; break; }
      if (!hit) continue;
      var len = 0;
      for (var j = 0; j < terms.length; j++) if (low.indexOf(terms[j]) === 0 && terms[j].length > len) len = terms[j].length;
      if (m.index > last) out.push({ text: text.slice(last, m.index), hit: false });
      out.push({ text: text.slice(m.index, m.index + len), hit: true });
      last = m.index + len;
      re.lastIndex = last;
    }
    if (last < text.length) out.push({ text: text.slice(last), hit: false });
    return out;
  }

  /* A ~110 character window of body text around the first matching term. */
  function snippet(d, terms, size) {
    size = size || 110;
    var x = String(d.x || '').replace(/\s+/g, ' ').trim();
    var re = /[a-z0-9À-￿]+/gi, m, at = -1;
    while ((m = re.exec(x))) {
      var low = m[0].toLowerCase();
      if (terms.some(function (t) { return low.indexOf(t) === 0; })) { at = m.index; break; }
    }
    if (at < 0) return x.slice(0, size) + (x.length > size ? '…' : '');
    var start = Math.max(0, at - 30);
    if (start > 0) { var sp = x.indexOf(' ', start); if (sp > 0 && sp < at) start = sp + 1; }
    var end = Math.min(x.length, start + size);
    return (start > 0 ? '…' : '') + x.slice(start, end) + (end < x.length ? '…' : '');
  }

  return { tokenize: tokenize, search: search, group: group, segments: segments, snippet: snippet, MAX_RESULTS: MAX_RESULTS };
});
