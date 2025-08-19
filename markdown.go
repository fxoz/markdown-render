package main

import (
	"os"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

func LoadComponent(filename string) string {
	contentMarkdown := ""
	if content, err := os.ReadFile(ContentDir + "/" + filename + ".md"); err != nil {
		return ""
	} else {
		contentMarkdown = string(content)
	}

	return MarkdownToHTML(contentMarkdown, false)
}

func MarkdownToHTML(contentMarkdown string, fullRender bool) string {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse([]byte(contentMarkdown))

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	body := string(markdown.Render(doc, renderer))
	if !fullRender {
		return body
	}

	// TODO - Find a cleaner way for HTML injection
	header := "<header>" + LoadComponent("header") + "</header>"
	footer := "<footer>" + LoadComponent("footer") + "</footer>"

	result := header + body + footer + `
<link rel="stylesheet" href="/_static/style.css">
<script src="/_static/index.js"></script>`

	return result
}
