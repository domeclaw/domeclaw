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

func newTransferTokenCommand(walletServiceFn func() (*wallet.Service, error)) *cobra.Command {
	var (
		chainID      int
		tokenAddress string
	)

	cmd := &cobra.Command{
		Use:   "transfer_token [from] [to] [amount] [password]",
		Short: "Transfer ERC20 tokens to an address",
		Long: `Transfer ERC20 tokens to an address.
Use this command for explicit ERC20 token transfers.`,
		Args: cobra.ExactArgs(4),
		Example: `# Transfer ERC20 tokens on Ethereum
picoclaw wallet transfer_token 0x123... 0x456... 100 mypassword --token 0xA0b86a33E6441e0E1f0a8c9C3a1F7d8E2B4c9D6E

# Transfer CLAW tokens on ClawSwift
picoclaw wallet transfer_token 0x123... 0x456... 50 mypassword --chain-id 7441`,
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
			if tokenAddress != "" && !common.IsHexAddress(tokenAddress) {
				return fmt.Errorf("invalid token address: %s", tokenAddress)
			}

			from := common.HexToAddress(fromAddress)
			to := common.HexToAddress(toAddress)
			var tokenAddr common.Address
			if tokenAddress != "" {
				tokenAddr = common.HexToAddress(tokenAddress)
			}

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

			// If no token address specified, use the chain's gas token
			if tokenAddr == (common.Address{}) {
				if chain.GasToken == "" {
					return fmt.Errorf("no token address specified and chain has no gas token configured")
				}
				tokenAddr = common.HexToAddress(chain.GasToken)
			}

			// Convert amount to proper decimals
			amountFloat, _ := amount.Float64()
			amountWei := chain.ToWei(amountFloat)

			// Perform token transfer
			tx, err := walletService.TransferToken(ctx, from, to, amountWei, password, chain.ChainID, tokenAddr)
			if err != nil {
				return fmt.Errorf("token transfer failed: %w", err)
			}

			fmt.Printf("✅ Token transfer successful!\n")
			fmt.Printf("📍 Transaction Hash: %s\n", tx.Hash().Hex())
			fmt.Printf("🔗 From: %s\n", from.Hex())
			fmt.Printf("🔗 To: %s\n", to.Hex())
			fmt.Printf("💰 Amount: %s\n", amountStr)
			fmt.Printf("🪙 Token: %s\n", tokenAddr.Hex())
			fmt.Printf("⛓️  Chain: %s (ID: %d)\n", chain.Name, chain.ChainID)

			fmt.Printf("\n📊 View transaction: %s/tx/%s\n", chain.Explorer, tx.Hash().Hex())

			return nil
		},
	}

	cmd.Flags().IntVar(&chainID, "chain-id", 0, "Chain ID (uses default chain if not specified)")
	cmd.Flags().StringVar(&tokenAddress, "token", "", "Token contract address (uses chain gas token if not specified)")

	return cmd
}