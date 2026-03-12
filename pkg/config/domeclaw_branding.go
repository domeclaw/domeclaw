// DomeClaw Branding Configuration
// This is the main branding file for DomeClaw

package config

const (
	// DomeClawBanner is the ASCII art banner for DomeClaw
	// DOME (4 characters) + CLAW (4 characters)
	DomeClawBanner = `
██████╗  ██████╗ ███╗   ███╗███████╗     ██████╗██╗      █████╗ ██╗    ██╗
██╔══██╗██╔═══██╗████╗ ████║██╔════╝    ██╔════╝██║     ██╔══██╗██║    ██║
██║  ██║██║   ██║██╔████╔██║█████╗      ██║     ██║     ███████║██║ █╗ ██║
██║  ██║██║   ██║██║╚██╔╝██║██╔══╝      ██║     ██║     ██╔══██║██║███╗██║
██████╔╝╚██████╔╝██║ ╚═╝ ██║███████╗    ╚██████╗███████╗██║  ██║╚███╔███╔╝
╚═════╝  ╚═════╝ ╚═╝     ╚═╝╚══════╝     ╚═════╝╚══════╝╚═╝  ╚═╝ ╚══╝╚══╝
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