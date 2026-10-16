package marketplace

import (
	"context"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/avestura/zagros/internal/ext"
	"github.com/avestura/zagros/internal/market"
	"github.com/avestura/zagros/internal/theme"
)

func lum(c theme.Color) float64 {
	r, g, b := c.RGB()
	f := func(v uint8) float64 {
		s := float64(v) / 255
		if s <= 0.03928 {
			return s / 12.92
		}
		return math.Pow((s+0.055)/1.055, 2.4)
	}
	return 0.2126*f(r) + 0.7152*f(g) + 0.0722*f(b)
}

func contrast(a, b theme.Color) float64 {
	la, lb := lum(a), lum(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

func TestThemes(t *testing.T) {
	files, _ := filepath.Glob(filepath.Join("themes", "*.toml"))
	if len(files) < 7 {
		t.Fatalf("want at least 7 starter themes, found %d", len(files))
	}
	for _, f := range files {
		id := strings.TrimSuffix(filepath.Base(f), ".toml")
		t.Run(id, func(t *testing.T) {
			data, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			th, err := market.ValidateTheme(data, id)
			if err != nil {
				t.Fatal(err)
			}
			if th.Name != id {
				t.Errorf("theme name %q must equal file name %q", th.Name, id)
			}
			c := th.Colors
			pairs := []struct {
				name string
				fg   theme.Color
				bg   theme.Color
				min  float64
			}{
				{"text/background", c.Text, c.Background, 4.5},
				{"input_text/input_bg", c.InputText, c.InputBg, 4.5},
				{"sel_text/sel_bg", c.SelText, c.SelBg, 3},
				{"text_dim/background", c.TextDim, c.Background, 3},
			}
			for _, p := range pairs {
				if got := contrast(p.fg, p.bg); got < p.min {
					t.Errorf("%s contrast %.2f < %.1f", p.name, got, p.min)
				}
			}
			if c.Background == 0 || c.Accent == 0 {
				t.Errorf("background and accent must be set explicitly")
			}
		})
	}
}

// rt serves canned bodies for URLs containing a key.
type rt map[string]string

func (r rt) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, body := range r {
		if strings.Contains(req.URL.String(), k) {
			return &http.Response{StatusCode: 200, Header: http.Header{}, Request: req,
				Body: io.NopCloser(strings.NewReader(body))}, nil
		}
	}
	return &http.Response{StatusCode: 404, Header: http.Header{}, Request: req,
		Body: io.NopCloser(strings.NewReader("{}"))}, nil
}

var fixtures = map[string]struct {
	query string
	web   rt
	want  string // substring expected in the first item's title
}{
	"wiki": {"golang", rt{"rest.php/v1/search/title": `{"pages":[{"id":1,"key":"Go_(programming_language)","title":"Go (programming language)","description":"Programming language"}]}`}, "Go (programming"},
	"hackernews": {"", rt{
		"topstories.json": `[11,12]`,
		"item/11.json":    `{"title":"Story One","url":"https://example.com","score":10,"by":"a","descendants":3}`,
		"item/12.json":    `{"title":"Story Two","score":5,"by":"b","descendants":0}`,
	}, "Story One"},
	"npm-search": {"left-pad", rt{"/-/v1/search": `{"objects":[{"package":{"name":"left-pad","version":"1.3.0","description":"pad","links":{"npm":"x"}}}]}`}, "left-pad"},
	"weather":    {"paris", rt{"wttr.in": `{"current_condition":[{"temp_C":"12","temp_F":"54","FeelsLikeC":"10","FeelsLikeF":"50","humidity":"70","windspeedKmph":"9","windspeedMiles":"6","winddir16Point":"NW","weatherDesc":[{"value":"Cloudy"}]}],"nearest_area":[{"areaName":[{"value":"Paris"}],"country":[{"value":"France"}]}],"weather":[{"date":"2025-01-01","maxtempC":"13","mintempC":"5","maxtempF":"55","mintempF":"41","hourly":[{"weatherDesc":[{"value":"Rain"}]}]}]}`}, "12"},
	"dictionary": {"serendipity", rt{"entries/en/serendipity": `[{"word":"serendipity","phonetic":"/x/","meanings":[{"partOfSpeech":"noun","definitions":[{"definition":"A happy accident.","example":"What serendipity"}]}]}]`}, "happy accident"},
	"snippets":   {"", nil, "No snippets yet"},
}

func TestExtensions(t *testing.T) {
	dirs, _ := filepath.Glob(filepath.Join("extensions", "*"))
	if len(dirs) < 4 {
		t.Fatalf("want at least 4 extensions, found %d", len(dirs))
	}
	for _, dir := range dirs {
		id := filepath.Base(dir)
		t.Run(id, func(t *testing.T) {
			if is := ext.Validate(dir); len(is) > 0 {
				t.Fatalf("validate: %v", is)
			}
			for _, f := range []string{"manifest.toml", "main.star", "README.md"} {
				if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
					t.Error(err)
				}
			}
			fx, ok := fixtures[id]
			if !ok {
				t.Fatalf("no query fixture for %s", id)
			}
			e, err := ext.OpenExtension(dir, ext.Options{
				DataDir:    t.TempDir(),
				HTTPClient: &http.Client{Transport: fx.web},
			})
			if err != nil {
				t.Fatal(err)
			}
			if e.Manifest.ID != id {
				t.Errorf("manifest id %q != dir %q", e.Manifest.ID, id)
			}
			if len(e.Manifest.Keywords) == 0 {
				t.Errorf("missing keywords")
			}
			e.SetGranted(e.Manifest.Permissions)
			ctx := context.Background()
			if err := e.Start(ctx); err != nil {
				t.Fatal(err)
			}
			items, err := e.Query(ctx, fx.query)
			if err != nil {
				t.Fatal(err)
			}
			if len(items) == 0 || !strings.Contains(items[0].Title, fx.want) {
				t.Fatalf("query %q: got %+v, want first title containing %q", fx.query, items, fx.want)
			}
			if id == "snippets" {
				testSnippets(t, e)
			}
		})
	}
}

func testSnippets(t *testing.T, e *ext.Extension) {
	ctx := context.Background()
	items, err := e.Query(ctx, "add greet = hello there")
	if err != nil || len(items) != 1 || items[0].Action.Kind != ext.KindExtCall {
		t.Fatalf("add: %v %+v", err, items)
	}
	if err := e.Execute(ctx, items[0].Action); err != nil {
		t.Fatal(err)
	}
	items, err = e.Query(ctx, "gre")
	if err != nil || len(items) != 1 || items[0].Title != "greet" || items[0].Action.Kind != "copy" || items[0].Action.Target != "hello there" {
		t.Fatalf("list: %v %+v", err, items)
	}
	if err := e.Execute(ctx, items[0].Alt[0].Action); err != nil {
		t.Fatal(err)
	}
	items, _ = e.Query(ctx, "")
	if len(items) != 1 || items[0].Title != "No snippets yet" {
		t.Fatalf("after delete: %+v", items)
	}
}

// Every theme and extension must have a registry entry that builds offline.
func TestRegistryCoversMarketplace(t *testing.T) {
	root := ".."
	specs, err := market.LoadSpecs(filepath.Join(root, "registry", "extensions"))
	if err != nil {
		t.Fatal(err)
	}
	byDir := map[string]market.Spec{}
	for _, s := range specs {
		if s.Repo == "avestura/zagros" {
			byDir[s.Subdir] = s
		}
	}
	check := func(sub string, kind string) {
		s, ok := byDir[sub]
		if !ok {
			t.Errorf("no registry entry with subdir %q", sub)
			return
		}
		if s.Kind != kind {
			t.Errorf("%s: kind %q, want %q", sub, s.Kind, kind)
		}
		e, err := market.EntryFromLocal(s, root)
		if err != nil {
			t.Errorf("%s: %v", sub, err)
			return
		}
		if e.Description == "" || len(e.Tags) == 0 {
			t.Errorf("%s: entry needs a description and tags", sub)
		}
	}
	themes, _ := filepath.Glob(filepath.Join("themes", "*.toml"))
	for _, f := range themes {
		check("marketplace/themes/"+filepath.Base(f), market.KindTheme)
	}
	exts, _ := filepath.Glob(filepath.Join("extensions", "*"))
	for _, d := range exts {
		check("marketplace/extensions/"+filepath.Base(d), market.KindExtension)
	}
}
