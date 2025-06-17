package views

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"net/http"
)

func Render(w http.ResponseWriter, tmpl *template.Template, data interface{}) error {
	if tmpl == nil {
		return fmt.Errorf("template is nil")
	}

	var buf bytes.Buffer

	var root string
	for _, t := range tmpl.Templates() {
		root = t.Name()
		break
	}

	if root == "" {
		return fmt.Errorf("template has no named blocks")
	}

	if err := tmpl.ExecuteTemplate(&buf, root, data); err != nil {
		return fmt.Errorf("failed to execute template %q: %v", root, err)
	}

	if _, err := buf.WriteTo(w); err != nil {
		return fmt.Errorf("failed to write template: %v", err)
	}

	return nil
}

func Render(w http.ResponseWriter, tmpl *template.Template, data interface{}) error {
	if tmpl == nil {
		return fmt.Errorf("template is nil")
	}

	var buf bytes.Buffer

	if err := tmpl.ExecuteTemplate(&buf, "base", data); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	if _, err := buf.WriteTo(w); err != nil {
		return fmt.Errorf("failed to write template: %v", err)
	}

	return nil
}
