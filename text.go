package main

import (
	"os"
	"strings"

	"golang.org/x/image/font"
)

func loadParagraphs(path string) ([]string, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []string{"（ファイルが見つかりません）"}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	raw := string(data)
	for len(raw) > 0 && raw[len(raw)-1] == '\n' {
		raw = raw[:len(raw)-1]
	}
	return strings.Split(raw, "\n"), nil
}

func wrapParagraphs(paragraphs []string, face font.Face, maxW int) []string {
	var wrapped []string
	for _, para := range paragraphs {
		if para == "" {
			wrapped = append(wrapped, "")
			continue
		}
		line := ""
		for _, ch := range para {
			test := line + string(ch)
			testW := font.MeasureString(face, test).Ceil()
			if testW > maxW {
				if line != "" {
					wrapped = append(wrapped, line)
					line = string(ch)
				} else {
					wrapped = append(wrapped, string(ch))
					line = ""
				}
			} else {
				line = test
			}
		}
		if line != "" {
			wrapped = append(wrapped, line)
		}
	}
	return wrapped
}

func paginate(lines []string, rowsPerPage int) [][]string {
	if rowsPerPage <= 0 {
		rowsPerPage = 1
	}
	var pages [][]string
	for i := 0; i < len(lines); i += rowsPerPage {
		end := i + rowsPerPage
		if end > len(lines) {
			end = len(lines)
		}
		pages = append(pages, lines[i:end])
	}
	return pages
}
