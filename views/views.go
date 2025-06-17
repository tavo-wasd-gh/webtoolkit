package views

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"net/http"
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

func Render(w http.ResponseWriter, tmpl *template.Template, data interface{}) error {
	if tmpl == nil {
		return fmt.Errorf("template is nil")
	}

	var buf bytes.Buffer

	if err := tmpl.ExecuteTemplate(&buf, tmpl.Name(), data); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	if _, err := buf.WriteTo(w); err != nil {
		return fmt.Errorf("failed to write template: %v", err)
	}

	return nil
}
