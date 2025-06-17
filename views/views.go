package views

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
)

func Init(viewFS embed.FS, viewMap map[string][]string, funcMap map[string]interface{}) (map[string]*template.Template, error) {
	templates := make(map[string]*template.Template)
	tmplFuncMap := template.FuncMap(funcMap)

	for name, paths := range viewMap {
		var contents []string

		for _, path := range paths {
			content, err := viewFS.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("failed to read template %s: %w", path, err)
			}
			contents = append(contents, string(content))
		}

		tmpl, err := template.New(filepath.Base(paths[0])).Funcs(tmplFuncMap).Parse(contents[0])
		if err != nil {
			return nil, fmt.Errorf("failed to parse base template %s: %w", paths[0], err)
		}

		for _, content := range contents[1:] {
			if _, err := tmpl.Parse(content); err != nil {
				return nil, fmt.Errorf("failed to parse included template: %w", err)
			}
		}

		templates[name] = tmpl
	}

	return templates, nil
}

func Render(w http.ResponseWriter, tmpl *template.Template, data interface{}) error {
	if tmpl == nil {
		return fmt.Errorf("template is nil")
	}

	var buf bytes.Buffer

	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	if _, err := buf.WriteTo(w); err != nil {
		return fmt.Errorf("failed to write template: %v", err)
	}

	return nil
}
