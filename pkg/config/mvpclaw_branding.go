// M Claw Branding Configuration
// This file is for future rebrand to M Claw
// To use: Replace domeclaw_branding.go with this file content

package config

const (
	// MVPClawBanner is the ASCII art banner for M Claw
	MVPClawBanner = `
███╗   ███╗    ██████╗██╗      █████╗ ██╗    ██╗
████╗ ████║   ██╔════╝██║     ██╔══██╗██║    ██║
██╔████╔██║   ██║     ██║     ███████║██║ █╗ ██║
██║╚██╔╝██║   ██║     ██║     ██╔══██║██║███╗██║
██║ ╚═╝ ██║   ╚██████╗███████╗██║  ██║╚███╔███╔╝
╚═╝     ╚═╝    ╚═════╝╚══════╝╚═╝  ╚═╝ ╚══╝╚══╝
`

	// MVPClawAppNameDisplay is the display name for the application
	MVPClawAppNameDisplay = "M Claw"

	// MVPClawAppShortDescription is the short description shown in CLI
	MVPClawAppShortDescription = "M Claw - Personal AI Assistant with Wallet & Webhook"

	// MVPClawAppLongDescription is the detailed description
	MVPClawAppLongDescription = "M Claw is a fork of PicoClaw with Ethereum wallet integration and webhook channel support."
)

// GetBanner returns the M Claw banner with colors
func MVPClawGetBanner() string {
	return MVPClawBanner
}

// GetAppName returns the application display name
func MVPClawGetAppName() string {
	return MVPClawAppNameDisplay
}
