// JFIN Claw Branding Configuration
// This file is for future rebrand to JFIN Claw
// To use: Replace domeclaw_branding.go with this file content

package config

const (
	// JFINClawBanner is the ASCII art banner for JFIN Claw
	// "JFIN Claw" = 9 characters (slightly wider)
	JFINClawBanner = `
     ██╗███████╗███████╗███╗   ██╗      ██████╗██╗      █████╗ ██╗    ██╗
     ██║██╔════╝██╔════╝████╗  ██║     ██╔════╝██║     ██╔══██╗██║    ██║
     ██║█████╗  █████╗  ██╔██╗ ██║     ██║     ██║     ███████║██║ █╗ ██║
██   ██║██╔══╝  ██╔══╝  ██║╚██╗██║     ██║     ██║     ██╔══██║██║███╗██║
╚█████╔╝██║     ███████╗██║ ╚████║     ╚██████╗███████╗██║  ██║╚███╔███╔╝
 ╚════╝ ╚═╝     ╚══════╝╚═╝  ╚═══╝      ╚═════╝╚══════╝╚═╝  ╚═╝ ╚══╝╚══╝
`

	// AppNameDisplay is the display name for the application
	AppNameDisplay = "JFIN Claw"

	// AppShortDescription is the short description shown in CLI
	AppShortDescription = "JFIN Claw - Personal AI Assistant with Wallet & Webhook"

	// AppLongDescription is the detailed description
	AppLongDescription = "JFIN Claw is a fork of PicoClaw with Ethereum wallet integration and webhook channel support."
)

// GetBanner returns the JFIN Claw banner with colors
func GetBanner() string {
	return JFINClawBanner
}

// GetAppName returns the application display name
func GetAppName() string {
	return AppNameDisplay
}
