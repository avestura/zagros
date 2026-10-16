package theme

// Boxy cool-gray ramp.
const (
	n0   Color = 0xffffff
	n50  Color = 0xf6f7f8
	n100 Color = 0xeceef0
	n200 Color = 0xdde1e4
	n300 Color = 0xc7ccd1
	n500 Color = 0x7f8993
	n600 Color = 0x616b75
	n800 Color = 0x343b41
	n900 Color = 0x23282d
	n950 Color = 0x16191c
)

func baseMetrics(name, author string, dark bool) Theme {
	return Theme{
		Name: name, Author: author, Dark: dark,
		Font: "Segoe UI", MonoFont: "Cascadia Mono", FontSize: 11,
		Width: 640, InputHeight: 52, RowHeight: 44, Padding: 12,
		BorderWidth: 1, Radius: 0, ShowIcons: true, ShowTags: true,
	}
}

func boxyLight() Theme {
	t := baseMetrics("boxy-light", "Zagros", false)
	t.Colors = Colors{
		Background: n0, Border: n300,
		InputBg: n0, InputText: n900, Placeholder: n500,
		Text: n900, TextDim: n600,
		SelBg: n100, SelText: n950, SelTextDim: n600,
		Accent: 0x0f62fe,
		Tag:    n600, TagBg: n100,
	}
	return t
}

func boxyDark() Theme {
	t := baseMetrics("boxy-dark", "Zagros", true)
	t.Colors = Colors{
		Background: n950, Border: n800,
		InputBg: n900, InputText: n100, Placeholder: n500,
		Text: n100, TextDim: n500,
		SelBg: n800, SelText: n0, SelTextDim: n300,
		Accent: 0x78a9ff,
		Tag:    n300, TagBg: n800,
	}
	return t
}

// classic is a dark, Keypirinha-like theme.
func classic() Theme {
	t := baseMetrics("classic", "Zagros", true)
	t.Font = "Segoe UI"
	t.InputHeight = 46
	t.RowHeight = 40
	t.Radius = 0
	t.Colors = Colors{
		Background: 0x2b2b2b, Border: 0x555555,
		InputBg: 0x1e1e1e, InputText: 0xf0f0f0, Placeholder: 0x808080,
		Text: 0xe0e0e0, TextDim: 0x9a9a9a,
		SelBg: 0x3b6ea5, SelText: 0xffffff, SelTextDim: 0xc8d8ea,
		Accent: 0x5b9bd5,
		Tag:    0xb0b0b0, TagBg: 0x3c3c3c,
	}
	return t
}

func nord() Theme {
	t := baseMetrics("nord", "Zagros", true)
	t.Radius = 6
	t.Colors = Colors{
		Background: 0x2e3440, Border: 0x4c566a,
		InputBg: 0x3b4252, InputText: 0xeceff4, Placeholder: 0x7b88a1,
		Text: 0xeceff4, TextDim: 0x9aa5b8,
		SelBg: 0x434c5e, SelText: 0xffffff, SelTextDim: 0xd8dee9,
		Accent: 0x88c0d0,
		Tag:    0x81a1c1, TagBg: 0x3b4252,
	}
	return t
}

var builtins = []func() Theme{boxyLight, boxyDark, classic, nord}
