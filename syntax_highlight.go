package main

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"strings"

	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// highlightCodeToHTML renders code with chroma using <pre><code> (WithClasses when requested).
// It wraps the result in a toolbar+container to host language label and copy button.
// The line numbers are rendered in a <table> layout so copy can ignore the numbers.
func highlightCodeToHTML(code, lang string, opts MDOptions) (string, error) {
	// Pick lexer: explicit language from fence info, otherwise detect
	var lx chroma.Lexer
	if lang != "" {
		lx = lexers.Get(lang)
	}
	if lx == nil {
		lx = lexers.Analyse(code)
	}
	if lx == nil {
		lx = lexers.Fallback
	}
	lx = chroma.Coalesce(lx)

	// Style
	style := styles.Get(opts.HighlightStyle)
	if style == nil {
		style = styles.Fallback
	}

	// Formatter
	fmtr := chromahtml.New(
		chromahtml.WithClasses(opts.UseClasses),
		chromahtml.WithLineNumbers(opts.LineNumbers),
		// chromahtml.WithLineNumbersInTable(true), // numbers in a separate column (easier copy)
		chromahtml.WithLinkableLineNumbers(opts.LinkableLineNumbers, "L"),
		// chromahtml.WithTabWidth(opts.TabWidth),
	)

	it, err := lx.Tokenise(nil, code)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer

	// Optional toolbar above the code block:
	buf.WriteString(`<div class="codeblock"`)
	if lang != "" {
		buf.WriteString(` data-lang="` + html.EscapeString(strings.ToLower(lang)) + `"`)
	}
	buf.WriteString(`>`)

	if opts.ShowLanguageLabel || opts.CopyButton {
		buf.WriteString(`<div class="codeblock-toolbar">`)
		if opts.ShowLanguageLabel {
			label := lang
			if label == "" {
				label = "text"
			}
			buf.WriteString(`<span class="codeblock-lang">` + html.EscapeString(strings.ToUpper(label)) + `</span>`)
		}
		if opts.CopyButton {
			buf.WriteString(`<button type="button" class="codeblock-copy" aria-label="Copy code">Copy</button>`)
		}
		buf.WriteString(`</div>`)
	}

	buf.WriteString(`<div class="codeblock-body">`)
	if err := fmtr.Format(&buf, style, it); err != nil {
		return "", err
	}
	buf.WriteString(`</div></div>`)

	return buf.String(), nil
}

// chromaCSS returns the stylesheet for the selected style when UseClasses=true.
// Serve this as /_static/chroma.css or inline it.
func chromaCSS(styleName string) (string, error) {
	style := styles.Get(styleName)
	if style == nil {
		style = styles.Fallback
	}
	var buf bytes.Buffer
	fmtr := chromahtml.New(chromahtml.WithClasses(true))
	if err := fmtr.WriteCSS(&buf, style); err != nil {
		return "", err
	}
	// Add a few CSS hooks for toolbar and layout
	io.WriteString(&buf, `
/* generic codeblock frame */
.codeblock { margin: 1rem 0; border: 1px solid #e5e7eb; border-radius: 0.5rem; overflow: hidden; }
.codeblock-toolbar { display:flex; align-items:center; justify-content:space-between; padding:0.375rem 0.5rem; font: 12px/1.2 system-ui, sans-serif; background:#f8fafc; border-bottom:1px solid #e5e7eb; }
.codeblock-lang { opacity:0.75; letter-spacing:0.04em; }
.codeblock-copy { border:0; background:#eef2ff; padding:0.25rem 0.5rem; border-radius:0.375rem; cursor:pointer; }
.codeblock-copy.is-copied { background:#dcfce7; }
.codeblock-body { overflow:auto; }

/* chroma tables for line numbers */
.chroma table { border-spacing: 0; width: 100%; }
.chroma td { vertical-align: top; }
.chroma .lntd { width: 1%; user-select: none; opacity: 0.6; }
.chroma .lnt { padding: 0 0.75rem 0 0.75rem; text-align: right; }
.chroma .code { width: 99%; }
`)
	return buf.String(), nil
}

// extractFenceLang parses the first token in the info string, e.g. "go linenos"
func extractFenceLang(info []byte) string {
	s := strings.TrimSpace(string(info))
	if s == "" {
		return ""
	}
	fields := strings.Fields(s)
	return strings.ToLower(fields[0])
}

// fallbackPlain pre escapes code if highlighting is disabled
func fallbackPlain(code, lang string, opts MDOptions) string {
	escaped := html.EscapeString(code)
	langClass := ""
	if lang != "" {
		langClass = fmt.Sprintf(` class="language-%s"`, html.EscapeString(strings.ToLower(lang)))
	}
	return fmt.Sprintf(`<div class="codeblock"%s><div class="codeblock-toolbar">%s<button type="button" class="codeblock-copy">Copy</button></div><div class="codeblock-body"><pre><code%[2]s>%s</code></pre></div></div>`,
		ifElse(lang != "", ` data-lang="`+html.EscapeString(lang)+`"`, ""),
		langClass,
		escaped,
	)
}

func ifElse[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}
