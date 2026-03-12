// INK Claw Branding Configuration
// This file is for future rebrand to INK Claw
// To use: Replace domeclaw_branding.go with this file content

package config

const (
	// INKClawBanner is the ASCII art banner for INK Claw
	// I-N-K (3 characters) + CLAW (4 characters)
	INKClawBanner = `
██╗███╗   ██╗██╗  ██╗     ██████╗██╗      █████╗ ██╗    ██╗
██║████╗  ██║██║ ██╔╝    ██╔════╝██║     ██╔══██╗██║    ██║
██║██╔██╗ ██║████╔╝     ██║     ██║     ███████║██║ █╗ ██║
██║██║╚██╗██║██╔═██╗     ██║     ██║     ██╔══██║██║███╗██║
██║██║ ╚████║██║  ██╗    ╚██████╗███████╗██║  ██║╚███╔███╔╝
╚═╝╚═╝  ╚═══╝╚═╝  ╚═╝     ╚═════╝╚══════╝╚═╝  ╚═╝ ╚══╝╚══╝
`

	// INKClawAppNameDisplay is the display name for the application
	INKClawAppNameDisplay = "INK Claw"

	// INKClawAppShortDescription is the short description shown in CLI
	INKClawAppShortDescription = "INK Claw - Personal AI Assistant with Wallet & Webhook"

	// INKClawAppLongDescription is the detailed description
	INKClawAppLongDescription = "INK Claw is a fork of PicoClaw with Ethereum wallet integration and webhook channel support."
)

// INKClawGetBanner returns the INK Claw banner with colors
func INKClawGetBanner() string {
	return INKClawBanner
}

// INKClawGetAppName returns the application display name
func INKClawGetAppName() string {
	return INKClawAppNameDisplay
}
