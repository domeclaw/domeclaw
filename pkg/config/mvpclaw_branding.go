// MVP Claw Branding Configuration
// This file is for future rebrand to MVP Claw
// To use: Replace domeclaw_branding.go with this file content

package config

const (
	// MVPClawBanner is the ASCII art banner for MVP Claw
	// "MVP Claw" = 8 characters (same width as DomeClaw)
	MVPClawBanner = `
███╗   ███╗██╗   ██╗██████╗      ██████╗██╗      █████╗ ██╗    ██╗
████╗ ████║██║   ██║██╔══██╗    ██╔════╝██║     ██╔══██╗██║    ██║
██╔████╔██║██║   ██║██████╔╝    ██║     ██║     ███████║██║ █╗ ██║
██║╚██╔╝██║╚██╗ ██╔╝██╔═══╝     ██║     ██║     ██╔══██║██║███╗██║
██║ ╚═╝ ██║ ╚████╔╝ ██║         ╚██████╗███████╗██║  ██║╚███╔███╔╝
╚═╝     ╚═╝  ╚═══╝  ╚═╝          ╚═════╝╚══════╝╚═╝  ╚═╝ ╚══╝╚══╝
`

	// AppNameDisplay is the display name for the application
	AppNameDisplay = "MVP Claw"

	// AppShortDescription is the short description shown in CLI
	AppShortDescription = "MVP Claw - Personal AI Assistant with Wallet & Webhook"

	// AppLongDescription is the detailed description
	AppLongDescription = "MVP Claw is a fork of PicoClaw with Ethereum wallet integration and webhook channel support."
)

// GetBanner returns the MVP Claw banner with colors
func GetBanner() string {
	return MVPClawBanner
}

// GetAppName returns the application display name
func GetAppName() string {
	return AppNameDisplay
}
