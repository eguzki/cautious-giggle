package templates

import (
	"embed"
)

// DashboardContent holds templates
//go:embed dashboard.html.tmpl
var DashboardContent embed.FS

// ServiceDiscoveryContent holds templates
//go:embed servicediscovery.html.tmpl
var ServiceDiscoveryContent embed.FS

// NewApiContent holds templates
//go:embed newapi.html.tmpl
var NewApiContent embed.FS
