package wallet

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
	"github.com/sipeed/picoclaw/pkg/wallet"
)

func newInfoCommand(walletServiceFn func() (*wallet.Service, error)) *cobra.Command {
	var (
		address string
		chainID int
	)

	cmd := &cobra.Command{
		Use:   "info",
		Short: "Display wallet information and balance",
		Long:  `Display wallet address, balance, and other information for a specific address.`,
		Example: `# Show info for default wallet (first account)
picoclaw wallet info

# Show info for specific address
picoclaw wallet info --address 0x123...

# Show info for specific chain
picoclaw wallet info --address 0x123... --chain-id 7441`,
		RunE: func(cmd *cobra.Command, args []string) error {
			walletService, err := walletServiceFn()
			if err != nil {
				return fmt.Errorf("wallet service not available: %w", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			var walletInfo *wallet.WalletInfo

			if address != "" {
				// Validate address
				if !common.IsHexAddress(address) {
					return fmt.Errorf("invalid address: %s", address)
				}

				addr := common.HexToAddress(address)

				if chainID != 0 {
					// Get info for specific chain
				chain, err := walletService.GetChainByID(chainID)
				if err != nil {
					return fmt.Errorf("chain not found: %w", err)
				}
				walletInfo, err = walletService.GetWalletInfoForChain(ctx, addr, chainID)
				if err != nil {
					return fmt.Errorf("failed to get wallet info for chain %s: %w", chain.Name, err)
				}
				} else {
					// Get info for default chain
					walletInfo, err = walletService.GetWalletInfo(ctx, addr)
					if err != nil {
						return fmt.Errorf("failed to get wallet info: %w", err)
					}
				}
			} else {
				// Get info for first account (default)
				accounts := walletService.GetAccounts()
				if len(accounts) == 0 {
					return fmt.Errorf("no wallets found")
				}

				if chainID != 0 {
					walletInfo, err = walletService.GetWalletInfoForChain(ctx, accounts[0].Address, chainID)
					if err != nil {
						return fmt.Errorf("failed to get wallet info for chain %d: %w", chainID, err)
					}
				} else {
					walletInfo, err = walletService.GetWalletInfo(ctx, accounts[0].Address)
					if err != nil {
						return fmt.Errorf("failed to get wallet info: %w", err)
					}
				}
			}

			// Display wallet information
			fmt.Printf("💼 Wallet Information\n")
			fmt.Printf("===================\n")
			fmt.Printf("📍 Address: %s\n", walletInfo.Address)
			
			if walletInfo.Balance != nil {
				fmt.Printf("💰 Balance: %s %s\n", walletInfo.Balance.String(), walletInfo.Chain)
			}
			
			if walletInfo.Chain != "" {
				fmt.Printf("⛓️  Chain: %s (ID: %d)\n", walletInfo.Chain, walletInfo.ChainID)
			}

			if walletInfo.Path != "" {
				fmt.Printf("🔐 Keystore: %s\n", walletInfo.Path)
			}

			// Add explorer link if available
			if walletInfo.ChainID != 0 {
				if chain, err := walletService.GetChainByID(walletInfo.ChainID); err == nil && chain.Explorer != "" {
					fmt.Printf("\n📊 View on explorer: %s/address/%s\n", chain.Explorer, walletInfo.Address)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&address, "address", "", "Wallet address (uses first account if not specified)")
	cmd.Flags().IntVar(&chainID, "chain-id", 0, "Chain ID (uses default chain if not specified)")

	return cmd
}