package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sipeed/domeclaw/pkg/blockchain"
	"github.com/sipeed/domeclaw/pkg/config"
	"github.com/sipeed/domeclaw/pkg/logger"
)

// WalletAutoTool allows AI to automatically transfer tokens using stored PIN
type WalletAutoTool struct {
	workspace string
	cfg       *config.Config
}

// NewWalletAutoTool creates a new wallet auto tool
func NewWalletAutoTool(workspace string, cfg *config.Config) *WalletAutoTool {
	return &WalletAutoTool{
		workspace: workspace,
		cfg:       cfg,
	}
}

func (t *WalletAutoTool) Name() string {
	return "wallet_auto_transfer"
}

func (t *WalletAutoTool) Description() string {
	return "[HOTWALLET - NO MANUAL PIN REQUIRED] Automatically transfer ERC20 tokens using the configured wallet. " +
		"PIN is read automatically from workspace storage. " +
		"This tool allows AI to execute token transfers without user intervention. " +
		"Example: transfer CLAW tokens to another address immediately."
}

func (t *WalletAutoTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"token_address": map[string]any{
				"type":        "string",
				"description": "ERC20 token contract address (e.g., 0x20c0000000000000000000000000000000000000 for CLAW)",
			},
			"to_address": map[string]any{
				"type":        "string",
				"description": "Recipient address (0x...)",
			},
			"amount": map[string]any{
				"type":        "string",
				"description": "Amount to transfer in token units (e.g., '0.01', '100'). Supports decimals.",
			},
		},
		"required": []string{"token_address", "to_address", "amount"},
	}
}

// Execute performs the token transfer automatically
func (t *WalletAutoTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	tokenAddress, _ := args["token_address"].(string)
	toAddress, _ := args["to_address"].(string)
	amountStr, _ := args["amount"].(string)

	logger.InfoCF("wallet_auto", "Auto transfer initiated by AI", map[string]any{
		"token":  tokenAddress,
		"to":     toAddress,
		"amount": amountStr,
	})

	// Read PIN automatically from workspace
	pin, err := t.readPIN()
	if err != nil {
		logger.ErrorCF("wallet_auto", "Failed to read PIN from workspace", map[string]any{"error": err.Error()})
		return ErrorResult(fmt.Sprintf("Failed to read PIN: %v", err))
	}

	logger.InfoCF("wallet_auto", "PIN retrieved automatically", nil)

	// Validate addresses
	if len(tokenAddress) != 42 || tokenAddress[:2] != "0x" {
		return ErrorResult("Invalid token address format - must be 42 chars starting with 0x")
	}
	if len(toAddress) != 42 || toAddress[:2] != "0x" {
		return ErrorResult("Invalid recipient address format - must be 42 chars starting with 0x")
	}

	// Parse amount
	amountFloat := new(big.Float)
	if _, ok := amountFloat.SetString(amountStr); !ok {
		return ErrorResult("Invalid amount format - must be a valid number")
	}

	// Initialize keystore
	walletDir := filepath.Join(t.workspace, "wallet")
	ks := keystore.NewKeyStore(walletDir, keystore.StandardScryptN, keystore.StandardScryptP)

	accounts := ks.Accounts()
	if len(accounts) == 0 {
		return ErrorResult("No wallet found in keystore")
	}

	account := accounts[0]

	// Unlock account
	if err := ks.Unlock(account, pin); err != nil {
		logger.ErrorCF("wallet_auto", "Failed to unlock wallet", map[string]any{"error": err.Error()})
		return ErrorResult(fmt.Sprintf("Failed to unlock wallet with auto-retrieved PIN: %v", err))
	}
	defer ks.Lock(account.Address)

	logger.InfoCF("wallet_auto", "Wallet unlocked successfully", map[string]any{
		"address": account.Address.Hex(),
	})

	// Initialize blockchain
	bcClient := blockchain.NewClient()
	var chainID int64 = 7441 // default
	if t.cfg != nil && len(t.cfg.Wallet.Chains) > 0 {
		chain := &t.cfg.Wallet.Chains[0]
		chainID = chain.ChainID
		if err := bcClient.AddChain(chain); err != nil {
			logger.ErrorCF("wallet_auto", "Failed to connect to blockchain", map[string]any{"error": err.Error()})
			return ErrorResult(fmt.Sprintf("Blockchain connection failed: %v", err))
		}
	}

	// Get token decimals and convert amount
	decimals := int32(18)
	tokenAddr := common.HexToAddress(tokenAddress)

	// Try to get actual decimals from contract
	if dec, err := bcClient.GetTokenDecimals(ctx, chainID, tokenAddr); err == nil {
		decimals = dec
		logger.InfoCF("wallet_auto", "Token decimals detected", map[string]any{"decimals": decimals})
	}

	// Convert amount to wei
	multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	amount := new(big.Float).Mul(amountFloat, new(big.Float).SetInt(multiplier))
	amountInt, _ := amount.Int(nil)

	if amountInt.Sign() <= 0 {
		return ErrorResult("Amount must be greater than 0")
	}

	// Create transfer service
	transferService := blockchain.NewTransferService(bcClient)

	// Create signer
	signer := func(ctx context.Context, chainID int64, tx *types.Transaction) (*types.Transaction, error) {
		chainIDBig := big.NewInt(chainID)
		return ks.SignTx(account, tx, chainIDBig)
	}

	// Execute transfer
	toAddr := common.HexToAddress(toAddress)

	logger.InfoCF("wallet_auto", "Executing transfer", map[string]any{
		"from":   account.Address.Hex(),
		"to":     toAddress,
		"token":  tokenAddress,
		"amount": amountStr,
		"value":  amountInt.String(),
	})

	txHash, err := transferService.TransferERC20(
		ctx,
		chainID,
		account.Address,
		tokenAddr,
		toAddr,
		amountInt,
		signer,
	)

	if err != nil {
		logger.ErrorCF("wallet_auto", "Transfer failed", map[string]any{"error": err.Error()})
		return ErrorResult(fmt.Sprintf("Transfer failed: %v", err))
	}

	logger.InfoCF("wallet_auto", "Transfer successful", map[string]any{
		"tx_hash": txHash.Hex(),
	})

	return UserResult(fmt.Sprintf("✅ Auto-Transfer Successful!\n\nFrom: %s\nTo: %s\nAmount: %s\nToken: %s\n\nTransaction Hash: %s",
		account.Address.Hex(),
		toAddress,
		amountStr,
		tokenAddress,
		txHash.Hex(),
	))
}

// readPIN reads the PIN from workspace wallet directory
func (t *WalletAutoTool) readPIN() (string, error) {
	pinFile := filepath.Join(t.workspace, "wallet", "pin.json")
	data, err := os.ReadFile(pinFile)
	if err != nil {
		return "", fmt.Errorf("failed to read pin.json from %s: %w", pinFile, err)
	}

	var pinData struct {
		PIN string `json:"pin"`
	}
	if err := json.Unmarshal(data, &pinData); err != nil {
		return "", fmt.Errorf("failed to parse pin.json: %w", err)
	}

	return pinData.PIN, nil
}
