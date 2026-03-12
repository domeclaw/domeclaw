package wallet

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sipeed/picoclaw/pkg/wallet"
	"github.com/spf13/cobra"
)

func newCreateCommand(walletServiceFn func() (*wallet.Service, error)) *cobra.Command {
	var (
		chainID int
	)

	cmd := &cobra.Command{
		Use:   "create [password]",
		Short: "Create a new Ethereum wallet",
		Long:  "Create a new Ethereum wallet with password-protected JSON keystore",
		Args:  cobra.ExactArgs(1),
		Example: `# Create a new wallet with password
picoclaw wallet create mypassword

# Create a new wallet for specific chain
picoclaw wallet create mypassword --chain-id 7441`,
		RunE: func(cmd *cobra.Command, args []string) error {
			password := args[0]

			walletService, err := walletServiceFn()
			if err != nil {
				return fmt.Errorf("wallet service not available: %w", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			walletInfo, err := walletService.CreateWallet(ctx, password)
			if err != nil {
				return fmt.Errorf("failed to create wallet: %w", err)
			}

			fmt.Printf("✅ Wallet created successfully!\n")
			fmt.Printf("📍 Address: %s\n", walletInfo.Address)
			fmt.Printf("🔐 Keystore: %s\n", walletInfo.Path)
			fmt.Printf("\n⚠️  Important: Save your password securely. It cannot be recovered if lost.\n")
			fmt.Printf("⚠️  Backup your keystore file: %s\n", walletInfo.Path)

			// Save password to pin.json
			pinFilePath := filepath.Join(walletService.GetWorkspace(), "wallets", "pin.json")

			// Create wallets directory if not exists
			if err := os.MkdirAll(filepath.Dir(pinFilePath), 0700); err != nil {
				return fmt.Errorf("failed to create wallets directory: %w", err)
			}

			// Create pin data struct
			pinData := struct {
				Password string `json:"password"`
			}{Password: password}

			// Marshal to JSON
			pinJson, err := json.MarshalIndent(pinData, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal pin data: %w", err)
			}

			// Write to pin.json
			if err := os.WriteFile(pinFilePath, pinJson, 0600); err != nil {
				return fmt.Errorf("failed to write pin.json: %w", err)
			}

			// Verify pin.json exists
			if _, err := os.Stat(pinFilePath); err != nil {
				return fmt.Errorf("pin.json not found after creation: %w", err)
			}

			fmt.Printf("🔒 Password saved to: %s\n", pinFilePath)

			return nil
		},
	}

	cmd.Flags().IntVar(&chainID, "chain-id", 0, "Chain ID for the wallet (uses default chain if not specified)")

	return cmd
}
