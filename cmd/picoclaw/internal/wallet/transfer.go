package wallet

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
	"github.com/sipeed/picoclaw/pkg/wallet"
)

func newTransferCommand(walletServiceFn func() (*wallet.Service, error)) *cobra.Command {
	var (
		chainID int
	)

	cmd := &cobra.Command{
		Use:   "transfer [from] [to] [amount] [password]",
		Short: "Transfer ETH or ERC20 tokens to an address",
		Long: `Transfer ETH or ERC20 tokens to an address.
Automatically uses ERC20 transfer for non-native chains.`,
		Args: cobra.ExactArgs(4),
		Example: `# Transfer ETH on Ethereum
picoclaw wallet transfer 0x123... 0x456... 1.5 mypassword

# Transfer on specific chain (ClawSwift)
picoclaw wallet transfer 0x123... 0x456... 100 mypassword --chain-id 7441`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fromAddress := strings.TrimSpace(args[0])
			toAddress := strings.TrimSpace(args[1])
			amountStr := strings.TrimSpace(args[2])
			password := strings.TrimSpace(args[3])

			// Validate addresses
			if !common.IsHexAddress(fromAddress) {
				return fmt.Errorf("invalid from address: %s", fromAddress)
			}
			if !common.IsHexAddress(toAddress) {
				return fmt.Errorf("invalid to address: %s", toAddress)
			}

			from := common.HexToAddress(fromAddress)
			to := common.HexToAddress(toAddress)

			walletService, err := walletServiceFn()
			if err != nil {
				return fmt.Errorf("wallet service not available: %w", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			// Parse amount
			amount := new(big.Float)
			if _, ok := amount.SetString(amountStr); !ok {
				return fmt.Errorf("invalid amount: %s", amountStr)
			}

			// Convert to wei (we'll handle decimals in the service)
			amountInt := new(big.Int)
			amount.Int(amountInt)

			// Get chain configuration
			cfg := walletService.GetConfig()
			var chain *wallet.ChainConfig
			if chainID == 0 {
				// Use default chain (first configured)
				if len(cfg.Chains) == 0 {
					return fmt.Errorf("no chains configured")
				}
				chain = &cfg.Chains[0]
			} else {
				chain, err = walletService.GetChainByID(chainID)
				if err != nil {
					return fmt.Errorf("chain not found: %w", err)
				}
			}

			// Convert amount to proper decimals
			amountFloat, _ := amount.Float64()
			amountWei := chain.ToWei(amountFloat)

			// Perform transfer
			tx, err := walletService.Transfer(ctx, from, to, amountWei, password, chain.ChainID)
			if err != nil {
				return fmt.Errorf("transfer failed: %w", err)
			}

			fmt.Printf("✅ Transfer successful!\n")
			fmt.Printf("📍 Transaction Hash: %s\n", tx.Hash().Hex())
			fmt.Printf("🔗 From: %s\n", from.Hex())
			fmt.Printf("🔗 To: %s\n", to.Hex())
			fmt.Printf("💰 Amount: %s %s\n", amountStr, chain.Currency)
			fmt.Printf("⛓️  Chain: %s (ID: %d)\n", chain.Name, chain.ChainID)

			if !chain.IsNative {
				fmt.Printf("🪙 Token: %s (%s)\n", chain.GasTokenName, chain.GasToken)
			}

			fmt.Printf("\n📊 View transaction: %s/tx/%s\n", chain.Explorer, tx.Hash().Hex())

			return nil
		},
	}

	cmd.Flags().IntVar(&chainID, "chain-id", 0, "Chain ID (uses default chain if not specified)")

	return cmd
}