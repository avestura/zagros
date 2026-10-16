// Package theme defines Zagros themes. A theme is a TOML file in
// <ConfigDir>/themes/*.toml or one of the built-ins.
package theme

// Color is 0xRRGGBB.
type Color uint32

type Theme struct {
	Name   string `toml:"name"`
	Author string `toml:"author"`
	Dark   bool   `toml:"dark"`

	Font     string `toml:"font"`      // e.g. "Segoe UI"
	MonoFont string `toml:"mono_font"` // e.g. "Cascadia Mono"
	FontSize int    `toml:"font_size"` // points, default 11

	// Metrics in pixels at 96 DPI; the UI scales by monitor DPI.
	Width       int  `toml:"width"`        // window width, default 640
	InputHeight int  `toml:"input_height"` // default 48
	RowHeight   int  `toml:"row_height"`   // default 44
	Padding     int  `toml:"padding"`      // default 12
	BorderWidth int  `toml:"border_width"` // default 1
	Radius      int  `toml:"radius"`       // 0 for sharp (Boxy)
	ShowIcons   bool `toml:"show_icons"`
	ShowTags    bool `toml:"show_tags"`

	Colors Colors `toml:"colors"`
}

// Colors are hex strings in TOML ("#0f62fe"); see Color.UnmarshalText in color.go.
type Colors struct {
	Background  Color `toml:"background"`
	Border      Color `toml:"border"`
	InputBg     Color `toml:"input_bg"`
	InputText   Color `toml:"input_text"`
	Placeholder Color `toml:"placeholder"`
	Text        Color `toml:"text"`
	TextDim     Color `toml:"text_dim"`
	SelBg       Color `toml:"sel_bg"`
	SelText     Color `toml:"sel_text"`
	SelTextDim  Color `toml:"sel_text_dim"`
	Accent      Color `toml:"accent"`
	Tag         Color `toml:"tag"`
	TagBg       Color `toml:"tag_bg"`
}
