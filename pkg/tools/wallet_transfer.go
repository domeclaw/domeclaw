package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/logger"
	"github.com/sipeed/picoclaw/pkg/wallet"
)

// WalletTransferTool allows AI to transfer tokens using stored PIN
type WalletTransferTool struct {
	workspace string
	cfg       *config.Config
}

// NewWalletTransferTool creates a new wallet transfer tool
func NewWalletTransferTool(workspace string, cfg *config.Config) *WalletTransferTool {
	return &WalletTransferTool{
		workspace: workspace,
		cfg:       cfg,
	}
}

func (t *WalletTransferTool) Name() string {
	return "wallet_transfer"
}

func (t *WalletTransferTool) Description() string {
	return "[HOTWALLET - NO MANUAL PIN REQUIRED] Transfer tokens using the configured wallet. " +
		"PIN is read automatically from workspace storage. " +
		"This tool allows AI to execute token transfers without user intervention. " +
		"Use this when user asks to send/transfer tokens like 'ส่ง 0.01 CLAW ให้ 0xABC', 'โอนเงิน', 'transfer tokens'."
}

func (t *WalletTransferTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"to_address": map[string]any{
				"type":        "string",
				"description": "Recipient address (0x...)",
			},
			"amount": map[string]any{
				"type":        "string",
				"description": "Amount to transfer in token units (e.g., '0.01', '100'). Supports decimals.",
			},
			"token_address": map[string]any{
				"type":        "string",
				"description": "Optional: ERC20 token contract address. If not provided, transfers native chain token (CLAW)",
			},
		},
		"required": []string{"to_address", "amount"},
	}
}

// Execute performs the token transfer
func (t *WalletTransferTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	toAddress, _ := args["to_address"].(string)
	amountStr, _ := args["amount"].(string)
	tokenAddress, _ := args["token_address"].(string)

	logger.InfoCF("wallet_transfer", "Transfer initiated by AI", map[string]any{
		"to":     toAddress,
		"amount": amountStr,
		"token":  tokenAddress,
	})

	// Validate addresses
	if len(toAddress) != 42 || toAddress[:2] != "0x" {
		return ErrorResult("Invalid recipient address format - must be 42 chars starting with 0x")
	}

	// Read PIN automatically from workspace
	pin, err := t.readPIN()
	if err != nil {
		logger.ErrorCF("wallet_transfer", "Failed to read PIN from workspace", map[string]any{"error": err.Error()})
		return ErrorResult(fmt.Sprintf("Failed to read PIN: %v. Please create wallet with /wallet create first.", err))
	}

	// Parse amount
	amountFloat := new(big.Float)
	if _, ok := amountFloat.SetString(amountStr); !ok {
		return ErrorResult("Invalid amount format - must be a valid number")
	}

	// Create wallet service
	w := wallet.NewService(wallet.Config{
		Enabled: t.cfg.Wallet.Enabled,
		Chains:  convertChainConfigs(t.cfg.Wallet.Chains),
	}, t.workspace)
	if err := w.Initialize(ctx); err != nil {
		return ErrorResult(fmt.Sprintf("Failed to initialize wallet service: %v", err))
	}
	defer w.Close()

	// Get accounts from keystore
	accounts := w.GetAccounts()
	if len(accounts) == 0 {
		return ErrorResult("No wallet found in keystore")
	}
	fromAddr := accounts[0].Address

	// Use first configured chain
	if len(t.cfg.Wallet.Chains) == 0 {
		return ErrorResult("No chains configured")
	}
	chainID := t.cfg.Wallet.Chains[0].ChainID

	// Get chain info
	chain, err := w.GetChainByID(chainID)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to get chain info: %v", err))
	}

	// Convert amount to wei
	multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(chain.Decimal)), nil)
	amount := new(big.Float).Mul(amountFloat, new(big.Float).SetInt(multiplier))
	amountInt, _ := amount.Int(nil)

	if amountInt.Sign() <= 0 {
		return ErrorResult("Amount must be greater than 0")
	}

	logger.InfoCF("wallet_transfer", "Executing transfer", map[string]any{
		"from":   fromAddr.Hex(),
		"to":     toAddress,
		"amount": amountStr,
		"value":  amountInt.String(),
	})

	// Execute transfer
	toAddr := common.HexToAddress(toAddress)
	var txHash string

	if tokenAddress != "" {
		// ERC20 transfer
		tokenAddr := common.HexToAddress(tokenAddress)
		tx, err := w.TransferToken(ctx, fromAddr, toAddr, amountInt, pin, chainID, tokenAddr)
		if err != nil {
			logger.ErrorCF("wallet_transfer", "Transfer failed", map[string]any{"error": err.Error()})
			return ErrorResult(fmt.Sprintf("Transfer failed: %v", err))
		}
		txHash = tx.Hash().Hex()
	} else {
		// Native token transfer
		tx, err := w.Transfer(ctx, fromAddr, toAddr, amountInt, pin, chainID)
		if err != nil {
			logger.ErrorCF("wallet_transfer", "Transfer failed", map[string]any{"error": err.Error()})
			return ErrorResult(fmt.Sprintf("Transfer failed: %v", err))
		}
		txHash = tx.Hash().Hex()
	}

	logger.InfoCF("wallet_transfer", "Transfer successful", map[string]any{
		"tx_hash": txHash,
	})

	// Get explorer URL from config
	explorerURL := chain.Explorer
	if explorerURL == "" {
		explorerURL = "https://exp.clawswift.net"
	}

	result := fmt.Sprintf("✅ **Transfer Successful!**\n\n"+
		"📤 **From:** `%s`\n"+
		"📥 **To:** `%s`\n"+
		"💵 **Amount:** %s %s\n\n"+
		"🔗 **Transaction Hash:**\n`%s`\n\n"+
		"🔍 [View on Explorer](%s/tx/%s)",
		fromAddr.Hex(),
		toAddress,
		amountStr,
		chain.Currency,
		txHash,
		explorerURL,
		txHash,
	)

	return UserResult(result)
}

// readPIN reads the PIN from workspace wallets directory
func (t *WalletTransferTool) readPIN() (string, error) {
	pinFile := filepath.Join(t.workspace, "wallets", "pin.json")
	data, err := os.ReadFile(pinFile)
	if err != nil {
		return "", fmt.Errorf("failed to read pin.json from %s: %w", pinFile, err)
	}

	var pinData struct {
		Password string `json:"password"`
	}
	if err := json.Unmarshal(data, &pinData); err != nil {
		return "", fmt.Errorf("failed to parse pin.json: %w", err)
	}

	return pinData.Password, nil
}
