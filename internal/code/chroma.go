package code

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

type ChromaRenderer struct {
	lightStyle  *chroma.Style
	darkStyle   *chroma.Style
	formatter   *html.Formatter
}

func NewChromaRenderer(light, dark string) *ChromaRenderer {
	ls := styles.Get(light)
	if ls == nil {
		ls = styles.GitHub
	}
	ds := styles.Get(dark)
	if ds == nil {
		ds = styles.Monokai
	}
	return &ChromaRenderer{
		lightStyle: ls,
		darkStyle:  ds,
		formatter:  html.New(html.WithClasses(true), html.TabWidth(2)),
	}
}

func (r *ChromaRenderer) DetectLanguage(filename string) string {
	lexer := lexers.Match(filename)
	if lexer == nil {
		return "text"
	}
	return strings.ToLower(lexer.Config().Name)
}

func (r *ChromaRenderer) Render(filename string, content []byte) (*HighlightResult, error) {
	lexer := lexers.Match(filename)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	iterator, err := lexer.Tokenise(nil, string(content))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := r.formatter.Format(&buf, r.darkStyle, iterator); err != nil {
		return nil, err
	}

	result := &HighlightResult{
		HTML:    template.HTML(buf.String()),
		Symbols: []Symbol{},
	}

	filepath.Ext(filename)

	return result, nil
}

func GenerateChromaCSS(path, lightTheme, darkTheme string) error {
	lf := html.New(html.WithClasses(true))
	df := html.New(html.WithClasses(true))

	light := styles.Get(lightTheme)
	if light == nil {
		light = styles.GitHub
	}
	dark := styles.Get(darkTheme)
	if dark == nil {
		dark = styles.Monokai
	}

	var lbuf bytes.Buffer
	lf.WriteCSS(&lbuf, light)

	var dbuf bytes.Buffer
	df.WriteCSS(&dbuf, dark)

	var out bytes.Buffer
	out.WriteString("/* Chroma: light theme (" + lightTheme + ") */\n")
	for _, line := range strings.Split(lbuf.String(), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		out.WriteString(":root " + line + "\n")
	}
	out.WriteString("\n/* Chroma: dark theme (" + darkTheme + ") */\n")
	for _, line := range strings.Split(dbuf.String(), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		out.WriteString(".dark " + line + "\n")
	}

	return os.WriteFile(path, out.Bytes(), 0644)
}
