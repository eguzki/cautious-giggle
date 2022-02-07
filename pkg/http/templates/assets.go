package templates

import (
	"embed"
)

// TemplatesFS holds templates
//go:embed dashboard.html.tmpl
//go:embed servicediscovery.html.tmpl
//go:embed newapi.html.tmpl
//go:embed newapi-oasform.html.tmpl
//go:embed api.html.tmpl
//go:embed gateways.html.tmpl
//go:embed newplan.html.tmpl
//go:embed plan.html.tmpl
//go:embed users.html.tmpl
//go:embed newuser.html.tmpl
//go:embed user.html.tmpl
var TemplatesFS embed.FS
