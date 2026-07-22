package code

import "html/template"

type HighlightResult struct {
	HTML    template.HTML
	Symbols []Symbol
}

type Symbol struct {
	Kind     string
	Name     string
	Line     int
	Col      int
	Children []Symbol
}

type CodeRenderer interface {
	Render(filename string, content []byte) (*HighlightResult, error)
	DetectLanguage(filename string) string
}
