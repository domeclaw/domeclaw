package wallet

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sipeed/picoclaw/pkg/wallet"
	"github.com/spf13/cobra"
)

func newTransferCommand(walletServiceFn func() (*wallet.Service, error)) *cobra.Command {
	var (
		chainID int
	)

	cmd := &cobra.Command{
		Use:   "transfer [to] [amount]",
		Short: "Transfer ETH or ERC20 tokens to an address (auto-use only wallet and PIN from pin.json)",
		Long: `Transfer ETH or ERC20 tokens to an address using only wallet and PIN from pin.json.
Automatically uses ERC20 transfer for non-native chains.`,
		Args: cobra.ExactArgs(2),
		Example: `# Transfer on default chain (ClawSwift)
picoclaw wallet transfer 0x456... 0.01

# Transfer on specific chain (ClawSwift)
picoclaw wallet transfer 0x456... 0.01 --chain-id 7441`,
		RunE: func(cmd *cobra.Command, args []string) error {
			toAddress := strings.TrimSpace(args[0])
			amountStr := strings.TrimSpace(args[1])

			// Validate to address
			if !common.IsHexAddress(toAddress) {
				return fmt.Errorf("invalid to address: %s", toAddress)
			}

			to := common.HexToAddress(toAddress)

			walletService, err := walletServiceFn()
			if err != nil {
				return fmt.Errorf("wallet service not available: %w", err)
			}

			// Read password from pin.json
			pinFilePath := filepath.Join(walletService.GetWorkspace(), "wallets", "pin.json")
			pinJson, err := os.ReadFile(pinFilePath)
			if err != nil {
				return fmt.Errorf("failed to read pin.json: %w", err)
			}

			var pinData struct {
				Password string `json:"password"`
			}
			if err := json.Unmarshal(pinJson, &pinData); err != nil {
				return fmt.Errorf("failed to unmarshal pin.json: %w", err)
			}

			// Get default wallet address (only one wallet allowed)
			accounts := walletService.GetAccounts()
			if len(accounts) == 0 {
				return fmt.Errorf("no wallet found in keystore")
			}
			if len(accounts) > 1 {
				return fmt.Errorf("multiple wallets found - system only allows one wallet")
			}
			from := accounts[0].Address

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

			// Convert amount to proper decimals
			amountFloat, _ := amount.Float64()
			amountWei := chain.ToWei(amountFloat)

			// Perform transfer
			tx, err := walletService.Transfer(ctx, from, to, amountWei, pinData.Password, chain.ChainID)
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
