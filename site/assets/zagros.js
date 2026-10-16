/* Zagros site behaviour: theme toggle, copy buttons, doc search, TOC spy, marketplace. */
(function () {
  'use strict';
  var root = document.documentElement;
  var base = root.getAttribute('data-base') || '';

  function store(k, v) {
    try { if (v === undefined) return localStorage.getItem(k); localStorage.setItem(k, v); } catch (e) {}
    return null;
  }
  function esc(s) {
    return String(s == null ? '' : s).replace(/[&<>"]/g, function (c) {
      return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;' }[c];
    });
  }

  /* theme */
  document.querySelectorAll('[data-act="theme"]').forEach(function (b) {
    var val = b.querySelector('.val');
    function cur() {
      return root.getAttribute('data-theme') || (matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light');
    }
    function paint() { if (val) val.textContent = cur(); }
    paint();
    b.addEventListener('click', function () {
      var n = cur() === 'dark' ? 'light' : 'dark';
      root.setAttribute('data-theme', n);
      store('boxy:theme', n);
      paint();
    });
  });

  /* copy */
  document.addEventListener('click', function (e) {
    var b = e.target.closest('[data-copy]');
    if (!b) return;
    var text = b.getAttribute('data-copy');
    if (text === 'self') {
      var blk = b.closest('.bx-codeblock');
      text = blk ? blk.querySelector('pre').innerText : '';
    }
    var old = b.textContent;
    function done() { b.textContent = 'copied'; setTimeout(function () { b.textContent = old; }, 1200); }
    if (navigator.clipboard) navigator.clipboard.writeText(text).then(done, done); else done();
  });

  /* TOC scroll spy */
  var toc = document.querySelector('.bx-toc');
  if (toc && 'IntersectionObserver' in window) {
    var links = {};
    toc.querySelectorAll('a').forEach(function (a) { links[a.getAttribute('href').slice(1)] = a; });
    var io = new IntersectionObserver(function (es) {
      es.forEach(function (en) {
        if (en.isIntersecting && links[en.target.id]) {
          Object.keys(links).forEach(function (k) { links[k].removeAttribute('aria-current'); });
          links[en.target.id].setAttribute('aria-current', 'true');
        }
      });
    }, { rootMargin: '-80px 0px -70% 0px' });
    Object.keys(links).forEach(function (id) { var h = document.getElementById(id); if (h) io.observe(h); });
  }

  /* docs search: the matching lives in zagros-search.js, this is the UI around it */
  var input = document.getElementById('site-search');
  var box = document.getElementById('search-results');
  var wrap = document.querySelector('.site-search');
  var core = window.ZagrosSearch;
  if (input && box && wrap && core) {
    var index = null, state = 'idle', sel = -1, cur = [];
    var toggle = wrap.querySelector('.site-search__toggle');
    var closeBtn = wrap.querySelector('.site-search__close');

    function loadScript(cb) {
      var el = document.createElement('script');
      el.src = base + 'search-index.js';
      el.onload = function () { cb(window.__ZAGROS_INDEX || null); };
      el.onerror = function () { cb(null); };
      document.head.appendChild(el);
    }
    function load() {
      if (state === 'ready' || state === 'loading') return;
      state = 'loading';
      var done = function (j) {
        if (j && j.length) { index = j; state = 'ready'; } else { state = 'failed'; }
        render();
      };
      /* Relative to the page's data-base, so it works under /zagros/ and on any nesting depth.
         fetch() is blocked on file://, so fall back to the same data as a script. */
      fetch(base + 'search-index.json')
        .then(function (r) { if (!r.ok) throw new Error(r.status); return r.json(); })
        .then(done, function () { loadScript(done); });
    }

    function setOpen(open) {
      box.hidden = !open;
      input.setAttribute('aria-expanded', open ? 'true' : 'false');
      if (!open) input.removeAttribute('aria-activedescendant');
    }
    function marked(text, terms) {
      return core.segments(text, terms).map(function (p) {
        return p.hit ? '<mark>' + esc(p.text) + '</mark>' : esc(p.text);
      }).join('');
    }
    function select(i, scroll) {
      var items = box.querySelectorAll('.sr-item');
      if (!items.length) return;
      sel = (i + items.length) % items.length;
      items.forEach(function (a, n) { a.setAttribute('aria-selected', n === sel ? 'true' : 'false'); });
      input.setAttribute('aria-activedescendant', items[sel].id);
      if (scroll !== false) items[sel].scrollIntoView({ block: 'nearest' });
    }
    function message(text) { return '<p class="bx-palette__empty" role="status">' + text + '</p>'; }

    function render() {
      var raw = input.value.trim();
      sel = -1; cur = [];
      if (!core.tokenize(raw).length) { setOpen(false); return; }
      if (state === 'loading' || state === 'idle') { box.innerHTML = message('Loading search index…'); setOpen(true); return; }
      if (state === 'failed') { box.innerHTML = message('Search is unavailable: the index could not be loaded.'); setOpen(true); return; }
      var hits = core.search(index, raw);
      if (!hits.length) { box.innerHTML = message('No results for “' + esc(raw) + '”'); setOpen(true); return; }
      var terms = hits[0].terms, n = 0, html = '';
      core.group(hits).forEach(function (g) {
        html += '<div class="bx-palette__group" role="presentation">' + esc(g.label) + '</div>';
        g.hits.forEach(function (h) {
          var d = h.d;
          html += '<a class="sr-item" role="option" id="sr-' + n + '" aria-selected="false" href="' + esc(base + d.u) + '">' +
            '<span class="sr-item__title">' + marked(d.t, terms) + (d.h ? ' <span class="sep">/</span> ' + marked(d.h, terms) : '') + '</span>' +
            '<span class="sr-item__snip">' + marked(core.snippet(d, terms), terms) + '</span></a>';
          cur.push(base + d.u); n++;
        });
      });
      html += '<div class="bx-palette__foot" aria-hidden="true"><kbd class="bx-kbd">↑</kbd><kbd class="bx-kbd">↓</kbd><span class="bx-label">move</span><kbd class="bx-kbd">↵</kbd><span class="bx-label">open</span><kbd class="bx-kbd">esc</kbd><span class="bx-label">close</span></div>';
      box.innerHTML = html;
      setOpen(true);
      select(0, false);
    }

    function openField() {
      wrap.classList.add('is-open');
      toggle.setAttribute('aria-expanded', 'true');
      input.focus();
    }
    function closeField() {
      wrap.classList.remove('is-open');
      toggle.setAttribute('aria-expanded', 'false');
      setOpen(false);
    }

    input.addEventListener('input', function () { load(); render(); });
    input.addEventListener('focus', function () { load(); if (input.value.trim()) render(); });
    input.addEventListener('keydown', function (e) {
      if (e.key === 'ArrowDown' || e.key === 'ArrowUp') {
        if (box.hidden && input.value.trim()) render();
        if (!cur.length) return;
        e.preventDefault();
        select(sel + (e.key === 'ArrowDown' ? 1 : -1));
      } else if (e.key === 'Enter') {
        if (cur.length) { e.preventDefault(); location.href = cur[Math.max(sel, 0)]; }
      } else if (e.key === 'Escape') {
        if (!box.hidden) setOpen(false);
        else { input.value = ''; input.blur(); closeField(); }
        e.preventDefault();
      }
    });
    box.addEventListener('mousemove', function (e) {
      var a = e.target.closest('.sr-item');
      if (!a) return;
      var i = Array.prototype.indexOf.call(box.querySelectorAll('.sr-item'), a);
      if (i !== sel) select(i, false);
    });
    box.addEventListener('mousedown', function (e) { e.preventDefault(); }); /* keep focus in the field */
    toggle.addEventListener('click', openField);
    closeBtn.addEventListener('click', function () { input.value = ''; closeField(); toggle.focus(); });
    document.addEventListener('click', function (e) { if (!e.target.closest('.site-search')) closeField(); });
    document.addEventListener('keydown', function (e) {
      var typing = /INPUT|TEXTAREA|SELECT/.test(document.activeElement.tagName) || document.activeElement.isContentEditable;
      if ((e.key === '/' && !typing) || (e.key === 'k' && (e.ctrlKey || e.metaKey))) {
        e.preventDefault();
        if (toggle.offsetParent !== null) openField(); else input.focus();
      }
    });
  }

  /* marketplace */
  var grid = document.getElementById('mk-grid');
  if (grid) {
    var all = [], activeTags = {}, query = '';
    var tagBox = document.getElementById('mk-tags');
    var qbox = document.getElementById('mk-q');
    var count = document.getElementById('mk-count');
    var cmd = function (e) { return 'zagros ext install ' + e.id; };
    var gh = function (e) { return e.repo ? 'zagros ext install github:' + e.repo : ''; };
    var render = function () {
      var tags = Object.keys(activeTags);
      var list = all.filter(function (e) {
        var hay = [e.id, e.name, e.description, e.author, (e.tags || []).join(' ')].join(' ').toLowerCase();
        if (query && hay.indexOf(query) < 0) return false;
        return tags.every(function (t) { return (e.tags || []).indexOf(t) >= 0; });
      });
      count.textContent = list.length + ' of ' + all.length + ' extensions';
      if (!all.length) {
        grid.innerHTML = '<div class="bx-empty" style="grid-column: 1 / -1"><p class="bx-label">Registry empty</p>' +
          '<h3>No extensions published yet</h3><p class="muted">Be the first: follow <a href="docs/marketplace.html">publishing to the marketplace</a>. ' +
          'You can already install any extension straight from GitHub with <code>zagros ext install github:owner/repo</code>.</p></div>';
        return;
      }
      if (!list.length) {
        grid.innerHTML = '<div class="bx-empty" style="grid-column: 1 / -1"><h3>Nothing matches</h3><p class="muted">Try a different search or clear the tag filters.</p></div>';
        return;
      }
      grid.innerHTML = list.map(function (e) {
        var perms = (e.permissions || []).map(function (p) {
          var name = typeof p === 'string' ? p : (p.name || p.id || '');
          var danger = /^(process.run|fs.write|network|clipboard.read)/.test(name);
          return '<li><span class="bx-tag bx-tag--sm' + (danger ? ' bx-tag--warning' : '') + '">' + esc(name) + '</span></li>';
        }).join('') || '<li><span class="bx-tag bx-tag--sm bx-tag--success">no permissions</span></li>';
        var c = cmd(e), g = gh(e), link = e.repo ? 'https://github.com/' + e.repo : '';
        return '<article class="bx-card mk-card"><div class="bx-card__head"><div><p class="bx-label">' + esc(e.author || 'unknown') +
          '</p><h3 style="font-size: var(--bx-text-md)">' + esc(e.name || e.id) + '</h3></div><span class="bx-tag bx-tag--sm">' + esc(e.latest ? 'v' + e.latest : '') + '</span></div>' +
          '<div class="bx-card__body"><p class="muted" style="font-size: var(--bx-text-base)">' + esc(e.description) + '</p>' +
          '<div><p class="bx-label" style="margin-block-end: 4px">Permissions</p><ul class="mk-perms">' + perms + '</ul></div>' +
          '<div class="mk-tags">' + (e.tags || []).map(function (t) { return '<span class="bx-tag bx-tag--sm">' + esc(t) + '</span>'; }).join('') + '</div>' +
          '<div class="mk-install"><code>' + esc(c) + '</code><button class="bx-btn bx-btn--sm bx-btn--ghost" data-copy="' + esc(c) + '">copy</button></div>' + (g ? '<div class="mk-install"><code>' + esc(g) + '</code><button class="bx-btn bx-btn--sm bx-btn--ghost" data-copy="' + esc(g) + '">copy</button></div>' : '') +
          (link ? '<a href="' + esc(link) + '" style="font-size: var(--bx-text-sm)">Source</a>' : '') + '</div></article>';
      }).join('');
    };
    var renderTags = function () {
      var set = {};
      all.forEach(function (e) { (e.tags || []).forEach(function (t) { set[t] = 1; }); });
      tagBox.innerHTML = Object.keys(set).sort().map(function (t) {
        return '<button class="bx-tag" aria-pressed="' + (activeTags[t] ? 'true' : 'false') + '" data-tag="' + esc(t) + '">' + esc(t) + '</button>';
      }).join('');
    };
    tagBox.addEventListener('click', function (e) {
      var b = e.target.closest('[data-tag]');
      if (!b) return;
      var t = b.getAttribute('data-tag');
      if (activeTags[t]) delete activeTags[t]; else activeTags[t] = true;
      renderTags();
      render();
    });
    qbox.addEventListener('input', function () { query = qbox.value.trim().toLowerCase(); render(); });
    fetch('registry.json').then(function (r) { if (!r.ok) throw new Error(r.status); return r.json(); })
      .then(function (j) {
        all = Array.isArray(j) ? j : (j.extensions || j.items || []);
        renderTags();
        render();
      })
      .catch(function () { all = []; renderTags(); render(); });
  }
})();
