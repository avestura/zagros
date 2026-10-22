// Package update checks GitHub releases for a newer Zagros. It only reads
// public release metadata; it never downloads or installs anything.
package update

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Repo is the GitHub repository whose releases are checked.
const Repo = "avestura/zagros"

// Asset is one downloadable file of a release.
type Asset struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Size int64  `json:"size"`
}

// Release is the part of a GitHub release Zagros cares about.
type Release struct {
	Tag        string    `json:"tag"`
	Version    string    `json:"version"` // tag without the leading "v"
	URL        string    `json:"url"`     // the release page
	Notes      string    `json:"notes"`
	Published  time.Time `json:"published"`
	Prerelease bool      `json:"prerelease"`
	Assets     []Asset   `json:"assets"`
}

// Result is the outcome of a check.
type Result struct {
	Current   string    `json:"current"`
	Latest    Release   `json:"latest"`
	Available bool      `json:"available"` // Latest is newer than Current
	Dev       bool      `json:"dev"`       // Current is not a release version (dev build), so nothing can be compared
	Ahead     bool      `json:"ahead"`     // Current is newer than Latest (a local or pre-release build)
	Checked   time.Time `json:"checked"`
}

// Options configures a check. The zero value checks the real repository.
type Options struct {
	Repo              string
	BaseURL           string // API base, default https://api.github.com (tests override it)
	HTTP              *http.Client
	IncludePrerelease bool
}

// Check asks GitHub for the latest release and compares it with current (a
// version such as "0.2.0", "v0.2.0" or the output of git describe; "dev" or
// "" mean an unversioned build).
func Check(ctx context.Context, current string, o Options) (Result, error) {
	rel, err := Latest(ctx, o)
	if err != nil {
		return Result{}, err
	}
	res := Result{Current: strings.TrimPrefix(current, "v"), Latest: rel, Checked: time.Now()}
	v, ok := parse(current)
	if !ok {
		res.Dev = true
		return res, nil
	}
	if lv, ok := parse(rel.Version); ok {
		switch c := compare(v, lv); {
		case c < 0:
			res.Available = true
		case c > 0:
			res.Ahead = true
		}
	}
	return res, nil
}

type ghRelease struct {
	TagName     string    `json:"tag_name"`
	HTMLURL     string    `json:"html_url"`
	Body        string    `json:"body"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	PublishedAt time.Time `json:"published_at"`
	Assets      []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
		Size int64  `json:"size"`
	} `json:"assets"`
}

func (g ghRelease) release() Release {
	r := Release{Tag: g.TagName, Version: strings.TrimPrefix(g.TagName, "v"), URL: g.HTMLURL, Notes: g.Body,
		Published: g.PublishedAt, Prerelease: g.Prerelease}
	for _, a := range g.Assets {
		r.Assets = append(r.Assets, Asset{Name: a.Name, URL: a.URL, Size: a.Size})
	}
	return r
}

// Latest fetches the newest published release (stable only unless
// IncludePrerelease is set).
func Latest(ctx context.Context, o Options) (Release, error) {
	repo, base := o.Repo, o.BaseURL
	if repo == "" {
		repo = Repo
	}
	if base == "" {
		base = "https://api.github.com"
	}
	hc := o.HTTP
	if hc == nil {
		hc = &http.Client{Timeout: 15 * time.Second}
	}
	url := fmt.Sprintf("%s/repos/%s/releases/latest", base, repo)
	if o.IncludePrerelease {
		url = fmt.Sprintf("%s/repos/%s/releases?per_page=10", base, repo)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Release{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "zagros-update-check")
	resp, err := hc.Do(req)
	if err != nil {
		return Release{}, fmt.Errorf("update check: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return Release{}, fmt.Errorf("update check: %w", err)
	}
	switch {
	case resp.StatusCode == http.StatusNotFound:
		return Release{}, errors.New("update check: no published release found")
	case resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests:
		return Release{}, errors.New("update check: GitHub rate limit reached, try again later")
	case resp.StatusCode != http.StatusOK:
		return Release{}, fmt.Errorf("update check: GitHub answered %s", resp.Status)
	}
	if !o.IncludePrerelease {
		var g ghRelease
		if err := json.Unmarshal(body, &g); err != nil {
			return Release{}, fmt.Errorf("update check: bad response: %w", err)
		}
		if g.TagName == "" {
			return Release{}, errors.New("update check: response has no tag")
		}
		return g.release(), nil
	}
	var list []ghRelease
	if err := json.Unmarshal(body, &list); err != nil {
		return Release{}, fmt.Errorf("update check: bad response: %w", err)
	}
	for _, g := range list {
		if !g.Draft && g.TagName != "" {
			return g.release(), nil
		}
	}
	return Release{}, errors.New("update check: no published release found")
}

// AssetFor picks the Windows zip for goarch ("amd64", "arm64") from a
// release, matching the names the release workflow produces
// (zagros_<version>_windows_<arch>.zip).
func AssetFor(r Release, goarch string) (Asset, bool) {
	suffix := "_windows_" + goarch + ".zip"
	for _, a := range r.Assets {
		if strings.HasSuffix(a.Name, suffix) {
			return a, true
		}
	}
	return Asset{}, false
}

// Summary is a one-line, human description of a result.
func (r Result) Summary() string {
	switch {
	case r.Available:
		return fmt.Sprintf("Zagros %s is available (you have %s)", r.Latest.Version, orDev(r.Current))
	case r.Dev:
		return fmt.Sprintf("Latest release is %s (this is a development build)", r.Latest.Version)
	case r.Ahead:
		return fmt.Sprintf("You have %s, newer than the latest release %s", r.Current, r.Latest.Version)
	}
	return fmt.Sprintf("Zagros %s is up to date", orDev(r.Current))
}

func orDev(v string) string {
	if v == "" {
		return "dev"
	}
	return v
}

// ---- versions ----

type semver struct {
	major, minor, patch int
	pre                 string
	extra               int // commits after the tag in `git describe` output
}

var verRe = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z.-]+?))?(?:\+[0-9A-Za-z.-]+)?$`)
var describeRe = regexp.MustCompile(`^(v?\d+\.\d+\.\d+)-(\d+)-g[0-9a-f]+(-dirty)?$`)

// parse reads a semantic version, tolerating a leading "v" and `git describe`
// suffixes (-N-gHASH[-dirty], counted as "N commits after the tag").
func parse(s string) (semver, bool) {
	s = strings.TrimSpace(s)
	extra := 0
	if m := describeRe.FindStringSubmatch(s); m != nil {
		s = m[1]
		extra, _ = strconv.Atoi(m[2])
	}
	m := verRe.FindStringSubmatch(s)
	if m == nil {
		return semver{}, false
	}
	a, _ := strconv.Atoi(m[1])
	b, _ := strconv.Atoi(m[2])
	c, _ := strconv.Atoi(m[3])
	return semver{major: a, minor: b, patch: c, pre: m[4], extra: extra}, true
}

func compare(a, b semver) int {
	for _, p := range [][2]int{{a.major, b.major}, {a.minor, b.minor}, {a.patch, b.patch}} {
		if p[0] != p[1] {
			if p[0] < p[1] {
				return -1
			}
			return 1
		}
	}
	// A pre-release is older than the same version without one.
	switch {
	case a.pre == "" && b.pre != "":
		return 1
	case a.pre != "" && b.pre == "":
		return -1
	case a.pre != b.pre:
		return comparePre(a.pre, b.pre)
	}
	switch {
	case a.extra > b.extra:
		return 1
	case a.extra < b.extra:
		return -1
	}
	return 0
}

func comparePre(a, b string) int {
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(as) && i < len(bs); i++ {
		x, ex := strconv.Atoi(as[i])
		y, ey := strconv.Atoi(bs[i])
		switch {
		case ex == nil && ey == nil && x != y:
			if x < y {
				return -1
			}
			return 1
		case ex == nil && ey != nil:
			return -1
		case ex != nil && ey == nil:
			return 1
		case ex != nil && ey != nil && as[i] != bs[i]:
			if as[i] < bs[i] {
				return -1
			}
			return 1
		}
	}
	switch {
	case len(as) < len(bs):
		return -1
	case len(as) > len(bs):
		return 1
	}
	return 0
}

// CompareVersions returns -1, 0 or 1 for a < b, a == b, a > b. ok is false
// when either side is not a semantic version.
func CompareVersions(a, b string) (c int, ok bool) {
	x, ok1 := parse(a)
	y, ok2 := parse(b)
	if !ok1 || !ok2 {
		return 0, false
	}
	return compare(x, y), true
}

// ---- cache ----

// CachePath is where the last result is kept.
func CachePath(dataDir string) string { return filepath.Join(dataDir, "update.json") }

// LoadCache reads the cached result if it is younger than maxAge.
func LoadCache(dataDir string, maxAge time.Duration) (Result, bool) {
	b, err := os.ReadFile(CachePath(dataDir))
	if err != nil {
		return Result{}, false
	}
	var r Result
	if json.Unmarshal(b, &r) != nil || time.Since(r.Checked) >= maxAge {
		return Result{}, false
	}
	return r, true
}

// SaveCache stores r.
func SaveCache(dataDir string, r Result) {
	if b, err := json.Marshal(r); err == nil {
		_ = os.MkdirAll(dataDir, 0o755)
		_ = os.WriteFile(CachePath(dataDir), b, 0o644)
	}
}

// CheckCached returns a cached result younger than maxAge, otherwise checks
// and caches. A cached result is re-evaluated against current so an upgrade
// clears the "available" flag immediately.
func CheckCached(ctx context.Context, dataDir, current string, maxAge time.Duration, o Options) (Result, error) {
	if r, ok := LoadCache(dataDir, maxAge); ok {
		r.Current = strings.TrimPrefix(current, "v")
		r.Available, r.Ahead, r.Dev = false, false, false
		if v, ok := parse(current); !ok {
			r.Dev = true
		} else if lv, ok := parse(r.Latest.Version); ok {
			switch c := compare(v, lv); {
			case c < 0:
				r.Available = true
			case c > 0:
				r.Ahead = true
			}
		}
		return r, nil
	}
	r, err := Check(ctx, current, o)
	if err == nil {
		SaveCache(dataDir, r)
	}
	return r, err
}
