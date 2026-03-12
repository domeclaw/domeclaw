// JIB Claw Branding Configuration
// This file is for future rebrand to JIB Claw
// To use: Replace domeclaw_branding.go with this file content

package config

const (
	// JIBClawBanner is the ASCII art banner for JIB Claw
	// J-I-B (3 characters) + CLAW (4 characters)
	JIBClawBanner = `
     ██╗██╗██████╗      ██████╗██╗      █████╗ ██╗    ██╗
     ██║██║██╔══██╗    ██╔════╝██║     ██╔══██╗██║    ██║
     ██║██║██████╔╝    ██║     ██║     ███████║██║ █╗ ██║
██   ██║██║██╔══██╗    ██║     ██║     ██╔══██║██║███╗██║
╚█████╔╝██║██████╔╝    ╚██████╗███████╗██║  ██║╚███╔███╔╝
 ╚════╝ ╚═╝╚═════╝      ╚═════╝╚══════╝╚═╝  ╚═╝ ╚══╝╚══╝
`

	// JIBClawAppNameDisplay is the display name for the application
	JIBClawAppNameDisplay = "JIB Claw"

	// JIBClawAppShortDescription is the short description shown in CLI
	JIBClawAppShortDescription = "JIB Claw - Personal AI Assistant with Wallet & Webhook"

	// JIBClawAppLongDescription is the detailed description
	JIBClawAppLongDescription = "JIB Claw is a fork of PicoClaw with Ethereum wallet integration and webhook channel support."
)

// JIBClawGetBanner returns the JIB Claw banner with colors
func JIBClawGetBanner() string {
	return JIBClawBanner
}

// JIBClawGetAppName returns the application display name
func JIBClawGetAppName() string {
	return JIBClawAppNameDisplay
}
