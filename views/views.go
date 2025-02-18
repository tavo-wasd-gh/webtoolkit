package views

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
)

func Init(viewFS embed.FS, viewMap map[string]string, funcMap map[string]interface{}) (map[string]*template.Template, error) {
	templates := make(map[string]*template.Template)

	tmplFuncMap := template.FuncMap(funcMap)

	for name, path := range viewMap {
		content, err := viewFS.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read template %s: %w", path, err)
		}

		tmpl, err := template.New(filepath.Base(path)).Funcs(tmplFuncMap).Parse(string(content))
		if err != nil {
			return nil, fmt.Errorf("failed to parse template %s: %w", path, err)
		}

		templates[name] = tmpl
	}

	return templates, nil
}

func Render(w http.ResponseWriter, tmpl *template.Template, data interface{}) error {
	var buf bytes.Buffer

	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	if _, err := buf.WriteTo(w); err != nil {
		return fmt.Errorf("failed to write template: %v", err)
	}

	return nil
}
