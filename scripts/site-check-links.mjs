#!/usr/bin/env node
// Checks internal links and asset references in site-dist/. Exit 1 on broken links.
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const OUT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..', 'site-dist');
const files = [];
(function walk(d) {
  for (const e of fs.readdirSync(d, { withFileTypes: true })) {
    const p = path.join(d, e.name);
    if (e.isDirectory()) walk(p); else files.push(p);
  }
})(OUT);

const ids = new Map();
const idsOf = (f) => {
  if (!ids.has(f)) ids.set(f, new Set([...fs.readFileSync(f, 'utf8').matchAll(/\bid="([^"]+)"/g)].map((m) => m[1])));
  return ids.get(f);
};
let bad = 0, checked = 0;
for (const f of files.filter((f) => f.endsWith('.html'))) {
  const html = fs.readFileSync(f, 'utf8');
  for (const m of html.matchAll(/\s(?:href|src)="([^"]*)"/g)) {
    const ref = m[1];
    if (!ref || /^(https?:|mailto:|data:|\/\/|javascript:)/.test(ref)) continue;
    // links the 404 page hardcodes against the Pages base path
    if (ref.startsWith('/zagros/')) continue;
    const [p, hash] = ref.split('#');
    const target = p ? path.resolve(path.dirname(f), p) : f;
    checked++;
    if (!fs.existsSync(target) || fs.statSync(target).isDirectory()) {
      console.error(`BROKEN ${path.relative(OUT, f)} -> ${ref}`); bad++; continue;
    }
    if (hash && target.endsWith('.html') && !idsOf(target).has(hash)) {
      console.error(`MISSING ANCHOR ${path.relative(OUT, f)} -> ${ref}`); bad++;
    }
  }
}
console.log(`checked ${checked} links in ${files.filter((f) => f.endsWith('.html')).length} pages, ${bad} broken`);
process.exit(bad ? 1 : 0);
