package views

import (
	"bytes"
	"compress/gzip"
	"embed"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
)

func Init(viewFS embed.FS, viewMap map[string][]string, funcMap map[string]interface{}) (map[string]*template.Template, error) {
	templates := make(map[string]*template.Template)
	tmplFuncMap := template.FuncMap(funcMap)

	for name, paths := range viewMap {
		rootName := strings.TrimSuffix(filepath.Base(paths[0]), filepath.Ext(paths[0]))

		tmpl, err := template.New(rootName).Funcs(tmplFuncMap).ParseFS(viewFS, paths...)
		if err != nil {
			return nil, fmt.Errorf("failed to parse templates for %s: %w", name, err)
		}
		templates[name] = tmpl
	}

	return templates, nil
}

func Render(w http.ResponseWriter, r *http.Request, tmpl *template.Template, data interface{}) error {
	if tmpl == nil {
		return fmt.Errorf("template is nil")
	}

	var buf bytes.Buffer

	if err := tmpl.ExecuteTemplate(&buf, tmpl.Name(), data); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	m := minify.New()
	m.AddFunc("text/html", html.Minify)

	minified, err := m.String("text/html", buf.String())
	if err != nil {
		return fmt.Errorf("failed to minify template: %v", err)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Vary", "Accept-Encoding")

	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")

		gzWriter := gzip.NewWriter(w)
		defer gzWriter.Close()

		if _, err := gzWriter.Write([]byte(minified)); err != nil {
			return fmt.Errorf("failed to write gzipped response: %v", err)
		}
	} else {
		if _, err := io.WriteString(w, minified); err != nil {
			return fmt.Errorf("failed to write response: %v", err)
		}
	}

	return nil
}
