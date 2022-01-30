package templates

import (
	"embed"
)

// DashboardContent holds templates
//go:embed dashboard.html.tmpl
var DashboardContent embed.FS
