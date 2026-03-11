// DomeClaw - Personal AI Assistant with Wallet & Webhook
// Forked from PicoClaw: https://github.com/sipeed/picoclaw
// License: MIT

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/sipeed/picoclaw/cmd/picoclaw/internal"
	"github.com/sipeed/picoclaw/cmd/picoclaw/internal/agent"
	"github.com/sipeed/picoclaw/cmd/picoclaw/internal/auth"
	"github.com/sipeed/picoclaw/cmd/picoclaw/internal/cron"
	"github.com/sipeed/picoclaw/cmd/picoclaw/internal/gateway"
	"github.com/sipeed/picoclaw/cmd/picoclaw/internal/migrate"
	"github.com/sipeed/picoclaw/cmd/picoclaw/internal/onboard"
	"github.com/sipeed/picoclaw/cmd/picoclaw/internal/skills"
	"github.com/sipeed/picoclaw/cmd/picoclaw/internal/status"
	"github.com/sipeed/picoclaw/cmd/picoclaw/internal/version"
	"github.com/sipeed/picoclaw/cmd/picoclaw/internal/wallet"

	"github.com/sipeed/picoclaw/pkg/config"
)

func NewPicoclawCommand() *cobra.Command {
	short := fmt.Sprintf("%s %s v%s\n\n", internal.Logo, config.AppNameDisplay, config.GetVersion())

	cmd := &cobra.Command{
		Use:     "domeclaw",
		Short:   short,
		Example: "domeclaw version",
	}

	cmd.AddCommand(
		onboard.NewOnboardCommand(),
		agent.NewAgentCommand(),
		auth.NewAuthCommand(),
		gateway.NewGatewayCommand(),
		status.NewStatusCommand(),
		cron.NewCronCommand(),
		migrate.NewMigrateCommand(),
		skills.NewSkillsCommand(),
		wallet.NewWalletCommand(),

		version.NewVersionCommand(),
	)

	return cmd
}

const (
	colorBlue = "\033[1;38;2;62;93;185m"
	colorRed  = "\033[1;38;2;213;70;70m"
)

// getBanner returns the DomeClaw colored banner
func getBanner() string {
	return "\r\n" +
		colorBlue + "██████╗  ██████╗ ███╗   ███╗███████╗ " + colorRed + "██████╗██╗      █████╗ ██╗    ██╗\n" +
		colorBlue + "██╔══██╗██╔═══██╗████╗ ████║██╔════╝" + colorRed + "██╔════╝██║     ██╔══██╗██║    ██║\n" +
		colorBlue + "██║  ██║██║   ██║██╔████╔██║█████╗  " + colorRed + "██║     ██║     ███████║██║ █╗ ██║\n" +
		colorBlue + "██║  ██║██║   ██║██║╚██╔╝██║██╔══╝  " + colorRed + "██║     ██║     ██╔══██║██║███╗██║\n" +
		colorBlue + "██████╔╝╚██████╔╝██║ ╚═╝ ██║███████╗" + colorRed + "╚██████╗███████╗██║  ██║╚███╔███╔╝\n" +
		colorBlue + "╚═════╝  ╚═════╝ ╚═╝     ╚═╝╚══════╝" + colorRed + " ╚═════╝╚══════╝╚═╝  ╚═╝ ╚══╝╚══╝\n " +
		"\033[0m\r\n"
}

func main() {
	fmt.Printf("%s", getBanner())
	cmd := NewPicoclawCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
