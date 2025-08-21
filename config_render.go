package main

type MDOptions struct {
	FullRender          bool
	SyntaxHighlight     bool
	HighlightStyle      string // e.g. "github", "monokai", "dracula", "solarized-light"
	UseClasses          bool   // true => CSS classes (customizable), false => inline styles
	LineNumbers         bool
	LinkableLineNumbers bool // adds <a id="L123"> anchors
	TabWidth            int
	ShowLanguageLabel   bool
	CopyButton          bool
}

func defaultMDOptions(full bool) MDOptions {
	return MDOptions{
		FullRender:          full,
		SyntaxHighlight:     true,
		HighlightStyle:      "github",
		UseClasses:          true, // enables CSS-based theming later
		LineNumbers:         true,
		LinkableLineNumbers: true,
		TabWidth:            4,
		ShowLanguageLabel:   true,
		CopyButton:          true,
	}
}
