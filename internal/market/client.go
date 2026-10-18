package market

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	defaultAPIBase  = "https://api.github.com"
	defaultMaxBytes = 20 << 20
	maxRegistrySize = 8 << 20
	maxRedirects    = 5
)

var (
	// ErrTooLarge is returned when a download exceeds Client.MaxBytes.
	ErrTooLarge = errors.New("market: download exceeds size limit")
	// ErrHostNotAllowed is returned for URLs outside the host allowlist.
	ErrHostNotAllowed = errors.New("market: host not allowed")
	// ErrChecksum is returned when a pinned sha256 does not match.
	ErrChecksum = errors.New("market: sha256 mismatch")
	// ErrNotFound is returned when an id/repo/release cannot be found.
	ErrNotFound = errors.New("market: not found")
)

var defaultHosts = []string{
	"github.com", "api.github.com", "objects.githubusercontent.com", "codeload.github.com",
}

// Client talks to the registry and GitHub. The zero value is not usable; use
// New. All fields may be set before first use (tests point them at httptest
// servers and set Insecure).
type Client struct {
	RegistryURL string        // default DefaultRegistryURL
	APIBase     string        // default https://api.github.com
	CacheDir    string        // registry cache directory; "" disables caching
	Token       string        // optional GitHub token; sent only to APIBase
	HTTP        *http.Client  // optional; its CheckRedirect is replaced on a copy
	MaxBytes    int64         // download cap, default 20 MB
	Timeout     time.Duration // per-request timeout, default 30s
	Insecure    bool          // allow http:// (tests only)
	// ExtraHosts are additional allowed hosts (host or host:port).
	ExtraHosts []string

	mu   sync.Mutex
	last *Registry
}

// New returns a client. dataDir is core DataDir; the cache lives in
// dataDir/cache. registryURL may be "" for the default. GITHUB_TOKEN is read
// from the environment.
func New(dataDir, registryURL string) *Client {
	c := &Client{RegistryURL: registryURL, Token: os.Getenv("GITHUB_TOKEN")}
	if dataDir != "" {
		c.CacheDir = filepath.Join(dataDir, "cache")
	}
	return c
}

func (c *Client) registryURL() string {
	if c.RegistryURL != "" {
		return c.RegistryURL
	}
	return DefaultRegistryURL
}

func (c *Client) apiBase() string {
	if c.APIBase != "" {
		return strings.TrimRight(c.APIBase, "/")
	}
	return defaultAPIBase
}

func (c *Client) maxBytes() int64 {
	if c.MaxBytes > 0 {
		return c.MaxBytes
	}
	return defaultMaxBytes
}

func (c *Client) allowedHost(u *url.URL) bool {
	if u.Scheme != "https" && !(c.Insecure && u.Scheme == "http") {
		return false
	}
	h := strings.ToLower(u.Host)
	for _, a := range defaultHosts {
		if h == a {
			return true
		}
	}
	for _, raw := range append([]string{c.registryURL(), c.apiBase()}, c.ExtraHosts...) {
		if pu, err := url.Parse(raw); err == nil && pu.Host != "" {
			if h == strings.ToLower(pu.Host) {
				return true
			}
		} else if h == strings.ToLower(raw) {
			return true
		}
	}
	return false
}

func (c *Client) httpClient() *http.Client {
	var hc http.Client
	if c.HTTP != nil {
		hc = *c.HTTP
	}
	if hc.Timeout == 0 {
		hc.Timeout = c.Timeout
		if hc.Timeout == 0 {
			hc.Timeout = 30 * time.Second
		}
	}
	hc.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= maxRedirects {
			return fmt.Errorf("market: stopped after %d redirects", maxRedirects)
		}
		if c.registryRedirect(req, via) {
			return nil
		}
		if !c.allowedHost(req.URL) {
			return fmt.Errorf("%w: redirect to %s", ErrHostNotAllowed, req.URL.Host)
		}
		return nil
	}
	return &hc
}

// registryRedirect lets the registry URL itself move (a Pages site gaining a
// custom domain): a redirect chain that started at the registry URL may reach
// other hosts, and a plain-http hop is upgraded to https rather than followed
// or refused. It reports whether req is acceptable; archives and API calls
// keep the strict allowlist.
func (c *Client) registryRedirect(req *http.Request, via []*http.Request) bool {
	if len(via) == 0 || via[0].URL.String() != c.registryURL() {
		return false
	}
	if req.URL.Scheme == "http" && !c.Insecure {
		req.URL.Scheme = "https"
		if h, port, ok := strings.Cut(req.URL.Host, ":"); ok && port == "80" {
			req.URL.Host = h
		}
	}
	return req.URL.Scheme == "https" || (c.Insecure && req.URL.Scheme == "http")
}

type response struct {
	Status int
	Header http.Header
	Body   []byte
}

// do performs a GET with the allowlist and size cap enforced. Non-2xx
// statuses are returned, not errors.
func (c *Client) do(ctx context.Context, rawURL string, hdr map[string]string, limit int64) (*response, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	if !c.allowedHost(u) {
		return nil, fmt.Errorf("%w: %s", ErrHostNotAllowed, rawURL)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "zagros-market")
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	if api, err := url.Parse(c.apiBase()); err == nil && c.Token != "" && strings.EqualFold(api.Host, u.Host) {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.ContentLength > limit {
		return nil, ErrTooLarge
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > limit {
		return nil, ErrTooLarge
	}
	return &response{Status: resp.StatusCode, Header: resp.Header, Body: body}, nil
}

// List fetches the registry. It sends If-None-Match using the cached ETag and
// falls back to the cached copy (Registry.Stale = true) when the network
// fails. It errors only when there is neither network nor cache.
func (c *Client) List(ctx context.Context) (*Registry, error) {
	cached, etag := c.readCache()
	hdr := map[string]string{"Accept": "application/json"}
	if cached != nil && etag != "" {
		hdr["If-None-Match"] = etag
	}
	resp, err := c.do(ctx, c.registryURL(), hdr, maxRegistrySize)
	var fetchErr error
	switch {
	case err != nil:
		fetchErr = err
	case resp.Status == http.StatusNotModified && cached != nil:
		return c.remember(cached), nil
	case resp.Status != http.StatusOK:
		fetchErr = fmt.Errorf("registry: HTTP %d", resp.Status)
	default:
		reg, perr := parseRegistry(resp.Body)
		if perr != nil {
			fetchErr = perr
			break
		}
		c.writeCache(resp.Body, resp.Header.Get("ETag"))
		return c.remember(reg), nil
	}
	if cached != nil {
		cached.Stale = true
		return c.remember(cached), nil
	}
	return nil, fmt.Errorf("registry unavailable and no cache: %w", fetchErr)
}

func (c *Client) remember(r *Registry) *Registry {
	c.mu.Lock()
	c.last = r
	c.mu.Unlock()
	return r
}

func parseRegistry(b []byte) (*Registry, error) {
	var r Registry
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, fmt.Errorf("registry: %w", err)
	}
	if r.Schema != SchemaVersion {
		return nil, fmt.Errorf("registry: unsupported schema %d", r.Schema)
	}
	return &r, nil
}

func (c *Client) cachePaths() (data, etag string) {
	return filepath.Join(c.CacheDir, "registry.json"), filepath.Join(c.CacheDir, "registry.etag")
}

func (c *Client) readCache() (*Registry, string) {
	if c.CacheDir == "" {
		return nil, ""
	}
	dp, ep := c.cachePaths()
	b, err := os.ReadFile(dp)
	if err != nil {
		return nil, ""
	}
	r, err := parseRegistry(b)
	if err != nil {
		return nil, ""
	}
	e, _ := os.ReadFile(ep)
	return r, strings.TrimSpace(string(e))
}

func (c *Client) writeCache(body []byte, etag string) {
	if c.CacheDir == "" {
		return
	}
	if os.MkdirAll(c.CacheDir, 0o755) != nil {
		return
	}
	dp, ep := c.cachePaths()
	_ = os.WriteFile(dp, body, 0o644)
	if etag != "" {
		_ = os.WriteFile(ep, []byte(etag), 0o644)
	} else {
		_ = os.Remove(ep)
	}
}

// Search ranks entries from the most recent List result (call List first;
// Search returns nil if no registry has been loaded). An empty query returns
// everything sorted by name.
func (c *Client) Search(query string) []Entry {
	c.mu.Lock()
	reg := c.last
	c.mu.Unlock()
	if reg == nil {
		return nil
	}
	return SearchEntries(reg.Extensions, query)
}

// SearchEntries filters and ranks entries for a query (all words must match).
func SearchEntries(entries []Entry, query string) []Entry {
	words := strings.Fields(strings.ToLower(query))
	type scored struct {
		e Entry
		s int
	}
	var out []scored
	for _, e := range entries {
		total := 0
		for _, w := range words {
			s := scoreWord(e, w)
			if s == 0 {
				total = 0
				break
			}
			total += s
		}
		if len(words) == 0 {
			total = 1
		}
		if total > 0 {
			if e.Verified {
				total++
			}
			out = append(out, scored{e, total})
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].s != out[j].s {
			return out[i].s > out[j].s
		}
		return strings.ToLower(out[i].e.Name) < strings.ToLower(out[j].e.Name)
	})
	res := make([]Entry, len(out))
	for i, o := range out {
		res[i] = o.e
	}
	return res
}

func scoreWord(e Entry, w string) int {
	id, name := strings.ToLower(e.ID), strings.ToLower(e.Name)
	switch {
	case id == w || name == w:
		return 100
	case strings.HasPrefix(id, w) || strings.HasPrefix(name, w):
		return 60
	case strings.Contains(id, w) || strings.Contains(name, w):
		return 40
	}
	for _, t := range e.Tags {
		if strings.EqualFold(t, w) {
			return 30
		}
	}
	if strings.Contains(strings.ToLower(e.Author), w) {
		return 15
	}
	if strings.Contains(strings.ToLower(e.Description), w) {
		return 10
	}
	return 0
}

// CheckUpdates compares installed (id -> version) with the registry and
// returns the extensions that have a newer version, sorted by id.
func (c *Client) CheckUpdates(ctx context.Context, installed map[string]string) ([]Update, error) {
	reg, err := c.List(ctx)
	if err != nil {
		return nil, err
	}
	return UpdatesFor(reg, installed), nil
}

// UpdatesFor is the pure part of CheckUpdates.
func UpdatesFor(reg *Registry, installed map[string]string) []Update {
	var ups []Update
	for _, e := range reg.Extensions {
		cur, ok := installed[e.ID]
		if !ok || e.Latest == "" {
			continue
		}
		if CompareVersions(e.Latest, cur) > 0 {
			ups = append(ups, Update{ID: e.ID, Installed: cur, Latest: e.Latest, Entry: e})
		}
	}
	sort.Slice(ups, func(i, j int) bool { return ups[i].ID < ups[j].ID })
	return ups
}

// Install resolves ref, downloads it and hands it to inst. The caller is
// responsible for showing pkg.Manifest.Permissions and asking for consent
// before calling this (the ext side re-validates the archive).
func (c *Client) Install(ctx context.Context, inst Installer, ref string) (*Package, error) {
	pkg, err := c.Resolve(ctx, ref)
	if err != nil {
		return nil, err
	}
	if pkg.Kind == KindTheme {
		return nil, fmt.Errorf("%s is a theme; use InstallTheme", ref)
	}
	id, err := inst.Install(ctx, pkg.Zip, pkg.Source)
	if err != nil {
		return nil, err
	}
	pkg.ID = id
	return pkg, nil
}

func sum256(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
