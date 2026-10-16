#!/usr/bin/env node
// Builds the site, then checks the search index and the matching logic, and serves
// site-dist/ under a /zagros/ base path to make sure the index loads from there.
import { execFileSync } from 'node:child_process';
import fs from 'node:fs';
import path from 'node:path';
import assert from 'node:assert/strict';
import { createRequire } from 'node:module';
import { fileURLToPath } from 'node:url';
import { serve } from './site-serve.mjs';

const ROOT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const require = createRequire(import.meta.url);
execFileSync(process.execPath, [path.join(ROOT, 'scripts/site-build.mjs')], { stdio: 'inherit' });

const core = require(path.join(ROOT, 'site/assets/zagros-search.js'));
const index = JSON.parse(fs.readFileSync(path.join(ROOT, 'site-dist/search-index.json'), 'utf8'));
let failed = 0;
const test = (name, fn) => {
  try { fn(); console.log('ok   ' + name); } catch (e) { failed++; console.log('FAIL ' + name + '\n     ' + e.message.split('\n')[0]); }
};

test('index is non-empty and every entry has t, u, g, x', () => {
  assert.ok(index.length > 50, 'only ' + index.length + ' entries');
  for (const d of index) assert.ok(d.t && d.u && d.g && typeof d.x === 'string' && d.x.length, JSON.stringify(d).slice(0, 80));
});
test('body text keeps its letters (no stripped characters)', () => {
  const all = index.map((d) => d.x).join(' ').toLowerCase();
  for (const w of ['permission', 'extension', 'sandbox']) assert.ok(all.includes(w), w + ' missing from index text');
});

for (const q of ['e', 'hotkey', 'permission', 'theme', 'everything', 'clipboard', 'alt+space']) {
  test(`query "${q}" returns results`, () => {
    const hits = core.search(index, q);
    assert.ok(hits.length > 0, 'no results');
    assert.ok(hits.length <= core.MAX_RESULTS);
  });
}
test('one letter matches titles and headings only', () => {
  for (const h of core.search(index, 'e')) {
    const starts = (s) => core.tokenize(s).some((t) => t.startsWith('e'));
    assert.ok(starts(h.d.t) || starts(h.d.h), `"${h.d.t} / ${h.d.h}" matched on body text`);
  }
});
test('two letters search body text', () => {
  const body = index.find((d) => !core.tokenize(d.t + ' ' + d.h).some((t) => t.startsWith('sa')) && core.tokenize(d.x).some((t) => t.startsWith('sa')));
  assert.ok(body && core.search(index, 'sa').length > 0);
});
test('matching is case-insensitive prefix matching', () => {
  assert.deepEqual(core.search(index, 'HOTKE').map((h) => h.d.u), core.search(index, 'hotke').map((h) => h.d.u));
  assert.ok(core.search(index, 'hotke').length > 0);
});
test('multi-word queries are AND', () => {
  const both = core.search(index, 'theme permission').length;
  assert.ok(both <= core.search(index, 'theme').length && both <= core.search(index, 'permission').length);
  assert.equal(core.search(index, 'hotkey qqqzzz').length, 0);
});
test('gibberish returns nothing', () => assert.equal(core.search(index, 'qqqzzzxx').length, 0));
test('titles outrank body-only matches', () => {
  const hits = core.search(index, 'permissions');
  assert.ok(/permission/i.test(hits[0].d.t + ' ' + hits[0].d.h), hits[0].d.t);
});
test('snippet and highlight segments keep the text intact', () => {
  const [h] = core.search(index, 'permission');
  const snip = core.snippet(h.d, h.terms);
  assert.equal(core.segments(snip, h.terms).map((s) => s.text).join(''), snip);
  assert.ok(core.segments(snip, h.terms).some((s) => s.hit));
});
test('results group by section', () => {
  const g = core.group(core.search(index, 'e'));
  assert.ok(g.length >= 1 && g.every((x) => x.label && x.hits.length));
});

// The page asks for <data-base> + 'search-index.json'; check that resolves under a /zagros/ base path.
const server = await serve(4199, '/zagros/');
try {
  const page = await (await fetch('http://localhost:4199/zagros/docs/permissions.html')).text();
  const base = /<html[^>]*data-base="([^"]*)"/.exec(page)?.[1];
  assert.equal(base, '../');
  const url = new URL(base + 'search-index.json', 'http://localhost:4199/zagros/docs/permissions.html');
  assert.equal(url.pathname, '/zagros/search-index.json');
  const r = await fetch(url);
  assert.equal(r.status, 200);
  assert.equal((await r.json()).length, index.length);
  assert.equal((await fetch('http://localhost:4199/zagros/search-index.js')).status, 200);
  assert.equal((await fetch('http://localhost:4199/zagros/assets/zagros-search.js')).status, 200);
  console.log('ok   index loads under the /zagros/ base path');
} catch (e) { failed++; console.log('FAIL base path: ' + e.message); }
server.close();

if (failed) { console.log(`\n${failed} failed`); process.exit(1); }
console.log('\nall search tests passed');
