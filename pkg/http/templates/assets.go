package templates

import (
	"embed"
)

// Content holds templates
//go:embed login.html.tmpl
var content embed.FS

func LoginTemplateSource() ([]byte, error) {
	return content.ReadFile("login.html.tmpl")
}
