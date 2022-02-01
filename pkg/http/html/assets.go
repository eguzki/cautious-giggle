package html

import (
	"embed"
)

// LoginContent holds login html page
//go:embed login.html
var LoginContent embed.FS

// NewGatewayContent holds newgateway html page
//go:embed newgateway.html
var NewGatewayContent embed.FS
