// TOK Claw Branding Configuration
// This file is for future rebrand to TOK Claw
// To use: Replace domeclaw_branding.go with this file content

package config

const (
	// TOKClawBanner is the ASCII art banner for TOK Claw
	// T-O-K (3 characters) + CLAW (4 characters)
	TOKClawBanner = `
████████╗ ██████╗ ██╗  ██╗     ██████╗██╗      █████╗ ██╗    ██╗
╚══██╔══╝██╔═══██╗██║ ██╔╝    ██╔════╝██║     ██╔══██╗██║    ██║
   ██║   ██║   ██║█████╔╝     ██║     ██║     ███████║██║ █╗ ██║
   ██║   ██║   ██║██╔═██╗     ██║     ██║     ██╔══██║██║███╗██║
   ██║   ╚██████╔╝██║  ██╗    ╚██████╗███████╗██║  ██║╚███╔███╔╝
   ╚═╝    ╚═════╝ ╚═╝  ╚═╝     ╚═════╝╚══════╝╚═╝  ╚═╝ ╚══╝╚══╝
`

	// TOKClawAppNameDisplay is the display name for the application
	TOKClawAppNameDisplay = "TOK Claw"

	// TOKClawAppShortDescription is the short description shown in CLI
	TOKClawAppShortDescription = "TOK Claw - Personal AI Assistant with Wallet & Webhook"

	// TOKClawAppLongDescription is the detailed description
	TOKClawAppLongDescription = "TOK Claw is a fork of PicoClaw with Ethereum wallet integration and webhook channel support."
)

// GetBanner returns the TOK Claw banner with colors
func TOKClawGetBanner() string {
	return TOKClawBanner
}

// GetAppName returns the application display name
func TOKClawGetAppName() string {
	return TOKClawAppNameDisplay
}
