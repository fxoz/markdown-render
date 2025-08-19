package main

import (
	"os"
	"path/filepath"
	"strings"
)

func ProcessPage(path string, dir string, pathsHtml map[string]string) {
	contentMarkdown := ""
	if strings.HasSuffix(path, "page.md") {
		if content, err := os.ReadFile(path); err != nil {
			panic("Error reading file: " + err.Error())
		} else {
			contentMarkdown = string(content)
		}

		path = strings.ReplaceAll(path, "\\", "/")
		finalPath := strings.TrimPrefix(path, dir)
		finalPath = strings.TrimSuffix(finalPath, "/page.md")
		finalPath = "/" + finalPath
		finalPath = strings.ReplaceAll(finalPath, "//", "/")

		pathsHtml[finalPath] = MarkdownToHTML(contentMarkdown, true)
	}
}

func SetupFiles() map[string]string {
	pathsHtml := map[string]string{}

	err := filepath.WalkDir(ContentDir, func(path string, d os.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		ProcessPage(path, ContentDir, pathsHtml)

		return nil
	})

	if err != nil {
		panic(err)
	}

	return pathsHtml
}
