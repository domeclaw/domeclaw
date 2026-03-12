package wallet

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sipeed/picoclaw/cmd/picoclaw/internal"
	"github.com/sipeed/picoclaw/pkg/wallet"
)

type deps struct {
	workspace    string
	walletService *wallet.Service
}

func NewWalletCommand() *cobra.Command {
	var d deps

	cmd := &cobra.Command{
		Use:   "wallet",
		Short: "Manage Ethereum wallets and blockchain operations",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := internal.LoadConfig()
			if err != nil {
				return fmt.Errorf("error loading config: %w", err)
			}

			// Check if wallet functionality is enabled
			if !cfg.Wallet.Enabled {
				return fmt.Errorf("wallet functionality is disabled in configuration")
			}

			d.workspace = cfg.WorkspacePath()
			
			// Convert config types to wallet types
			walletChains := make([]wallet.ChainConfig, len(cfg.Wallet.Chains))
			for i, chain := range cfg.Wallet.Chains {
				walletChains[i] = wallet.ChainConfig{
					Name:         chain.Name,
					ChainID:      chain.ChainID,
					RPC:          chain.RPC,
					Explorer:     chain.Explorer,
					Currency:     chain.Currency,
					IsNative:     chain.IsNative,
					GasToken:     chain.GasToken,
					GasTokenName: chain.GasTokenName,
					Decimal:      chain.Decimal,
				}
			}
			
			// Initialize wallet service
			d.walletService = wallet.NewService(wallet.Config{
				Enabled: cfg.Wallet.Enabled,
				Chains:  walletChains,
			}, d.workspace)

			// Initialize wallet service (connect to blockchains)
			if err := d.walletService.Initialize(cmd.Context()); err != nil {
				return fmt.Errorf("failed to initialize wallet service: %w", err)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	walletServiceFn := func() (*wallet.Service, error) {
		if d.walletService == nil {
			return nil, fmt.Errorf("wallet service is not initialized")
		}
		return d.walletService, nil
	}

	cmd.AddCommand(
		newCreateCommand(walletServiceFn),
		newTransferCommand(walletServiceFn),
		newTransferTokenCommand(walletServiceFn),
		newInfoCommand(walletServiceFn),
		newCallCommand(walletServiceFn),
		newWriteCommand(walletServiceFn),
	)

	return cmd
}