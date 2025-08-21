package main

import (
	"bytes"
	"io"
	"os"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/microcosm-cc/bluemonday"
)

// LoadComponent reads a .md file from ContentDir and returns rendered HTML (no header/footer)
func LoadComponent(filename string) string {
	b, err := os.ReadFile(ContentDir + "/" + filename + ".md")
	if err != nil {
		return ""
	}
	return MarkdownToHTMLWithOptions(string(b), func(o MDOptions) MDOptions {
		o.FullRender = false
		return o
	})
}

// MarkdownToHTML keeps your original signature, using sensible defaults.
func MarkdownToHTML(contentMarkdown string, fullRender bool) string {
	return MarkdownToHTMLWithOptions(contentMarkdown, func(o MDOptions) MDOptions {
		o.FullRender = fullRender
		return o
	})
}

// MarkdownToHTMLWithOptions lets you tweak renderer behavior via a mutator (no breaking change upstream).
func MarkdownToHTMLWithOptions(contentMarkdown string, mutate func(MDOptions) MDOptions) string {
	opts := defaultMDOptions(false)
	if mutate != nil {
		opts = mutate(opts)
	}

	// Markdown parser
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.Footnotes | parser.Tables
	p := parser.NewWithExtensions(extensions)

	// ---- Syntax highlight hook
	hook := func(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
		if !entering {
			return ast.GoToNext, false
		}
		switch n := node.(type) {
		case *ast.FencedCodeBlock:
			lang := extractFenceLang(n.Info)
			code := string(n.Literal)
			var htmlOut string
			var err error
			if opts.SyntaxHighlight {
				htmlOut, err = highlightCodeToHTML(code, lang, opts)
				if err != nil {
					htmlOut = fallbackPlain(code, lang, opts)
				}
			} else {
				htmlOut = fallbackPlain(code, lang, opts)
			}
			io.WriteString(w, htmlOut)
			return ast.SkipChildren, true
		case *ast.CodeBlock:
			// Indented code block (no language)
			code := string(n.Literal)
			htmlOut := ""
			if opts.SyntaxHighlight {
				var err error
				htmlOut, err = highlightCodeToHTML(code, "", opts)
				if err != nil {
					htmlOut = fallbackPlain(code, "", opts)
				}
			} else {
				htmlOut = fallbackPlain(code, "", opts)
			}
			io.WriteString(w, htmlOut)
			return ast.SkipChildren, true
		default:
			return ast.GoToNext, false
		}
	}

	// HTML renderer with our hook; target=_blank for external links
	htmlFlags := html.CommonFlags | html.HrefTargetBlank | html.NofollowLinks
	rendererOpts := html.RendererOptions{
		Flags:          htmlFlags,
		RenderNodeHook: hook,
		// KeepUnsafe:   false (default) — we sanitize later anyway
	}
	renderer := html.NewRenderer(rendererOpts)

	// Parse + render
	doc := p.Parse([]byte(contentMarkdown))
	var buf bytes.Buffer
	buf.Write(markdown.Render(doc, renderer))

	// Sanitize: allow Chroma’s markup + our toolbar
	policy := bluemonday.UGCPolicy()
	policy.AllowElements("span", "table", "thead", "tbody", "tr", "td", "a", "button", "div", "pre", "code")
	policy.AllowAttrs("class").OnElements("span", "table", "thead", "tbody", "tr", "td", "a", "div", "pre", "code", "button")
	policy.AllowAttrs("id").OnElements("a", "code")
	policy.AllowAttrs("type", "aria-label").OnElements("button")
	// (Optional) keep target=_blank on links we just set:
	policy.AllowAttrs("target", "rel").OnElements("a")

	body := string(policy.SanitizeBytes(buf.Bytes()))

	if !opts.FullRender {
		return body
	}

	// Header/footer components (markdown -> HTML)
	header := "<header>" + LoadComponent("header") + "</header>"
	footer := "<footer>" + LoadComponent("footer") + "</footer>"

	// Provide Chroma CSS if using classes
	chromaCSSLink := ``
	if opts.UseClasses {
		// You have two options:
		// 1) Write the CSS to /_static/chroma.css at startup and serve it
		// 2) Inline a <style> tag here (simpler to start)
		if css, err := chromaCSS(opts.HighlightStyle); err == nil {
			chromaCSSLink = "<style>" + css + "</style>"
		}
	}

	return header + body + footer + `
<link rel="stylesheet" href="/_static/style.css">` + chromaCSSLink + `
<script src="/_static/index.js" defer></script>`
}
