// DomeClaw Branding Configuration
// This file is DomeClaw-specific and should not conflict with upstream PicoClaw
// It separates branding from config to minimize merge conflicts

package config

const (
	// DomeClawBanner is the ASCII art banner for DomeClaw
	// Using 8-char width for "DomeClaw" vs 7-char for "PicoClaw"
	DomeClawBanner = `
██████╗  ██████╗ ███╗   ███╗███████╗ ██████╗██╗      █████╗ ██╗    ██╗
██╔══██╗██╔═══██╗████╗ ████║██╔════╝██╔════╝██║     ██╔══██╗██║    ██║
██║  ██║██║   ██║██╔████╔██║█████╗  ██║     ██║     ███████║██║ █╗ ██║
██║  ██║██║   ██║██║╚██╔╝██║██╔══╝  ██║     ██║     ██╔══██║██║███╗██║
██████╔╝╚██████╔╝██║ ╚═╝ ██║███████╗╚██████╗███████╗██║  ██║╚███╔███╔╝
╚═════╝  ╚═════╝ ╚═╝     ╚═╝╚══════╝ ╚═════╝╚══════╝╚═╝  ╚═╝ ╚══╝╚══╝
`

	// AppNameDisplay is the display name for the application
	AppNameDisplay = "DomeClaw"

	// AppShortDescription is the short description shown in CLI
	AppShortDescription = "DomeClaw - Personal AI Assistant with Wallet & Webhook"

	// AppLongDescription is the detailed description
	AppLongDescription = "DomeClaw is a fork of PicoClaw with Ethereum wallet integration and webhook channel support."
)

// GetBanner returns the DomeClaw banner with colors
func GetBanner() string {
	return DomeClawBanner
}

// GetAppName returns the application display name
func GetAppName() string {
	return AppNameDisplay
}
