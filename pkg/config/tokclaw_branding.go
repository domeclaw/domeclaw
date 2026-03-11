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

	// AppNameDisplay is the display name for the application
	AppNameDisplay = "TOK Claw"

	// AppShortDescription is the short description shown in CLI
	AppShortDescription = "TOK Claw - Personal AI Assistant with Wallet & Webhook"

	// AppLongDescription is the detailed description
	AppLongDescription = "TOK Claw is a fork of PicoClaw with Ethereum wallet integration and webhook channel support."
)

// GetBanner returns the TOK Claw banner with colors
func GetBanner() string {
	return TOKClawBanner
}

// GetAppName returns the application display name
func GetAppName() string {
	return AppNameDisplay
}
