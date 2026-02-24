package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sipeed/domeclaw/pkg/blockchain"
	"github.com/sipeed/domeclaw/pkg/config"
	"github.com/sipeed/domeclaw/pkg/logger"
)

// WalletQueryTool allows AI to query wallet balance directly from blockchain
type WalletQueryTool struct {
	workspace string
	cfg       *config.Config
}

// NewWalletQueryTool creates a new wallet query tool
func NewWalletQueryTool(workspace string, cfg *config.Config) *WalletQueryTool {
	return &WalletQueryTool{
		workspace: workspace,
		cfg:       cfg,
	}
}

func (t *WalletQueryTool) Name() string {
	return "query_wallet_balance"
}

func (t *WalletQueryTool) Description() string {
	return "Query wallet balance directly from blockchain. " +
		"This tool reads the wallet address from workspace and queries balance from ClawSwift chain. " +
		"Use this when user asks about their balance in natural language like 'เรามี balance เท่าไหร่' or 'เช็คยอดเงิน'. " +
		"Returns formatted balance with symbol and decimals."
}

func (t *WalletQueryTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"token_address": map[string]any{
				"type":        "string",
				"description": "Optional: ERC20 token contract address. If not provided, defaults to CLAW token (0x20c0000000000000000000000000000000000000)",
			},
		},
		"required": []string{},
	}
}

// Execute queries the wallet balance directly
func (t *WalletQueryTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	var tokenAddress string
	if ta, ok := args["token_address"].(string); ok && ta != "" {
		tokenAddress = ta
	} else {
		// Default to CLAW token
		tokenAddress = "0x20c0000000000000000000000000000000000000"
	}

	logger.InfoCF("wallet_query", "Querying balance", map[string]any{
		"token": tokenAddress,
	})

	// Get wallet address from workspace
	walletAddr, err := t.getWalletAddress()
	if err != nil {
		logger.ErrorCF("wallet_query", "Failed to get wallet address", map[string]any{"error": err.Error()})
		return ErrorResult(fmt.Sprintf("No wallet found. Please create one with /wallet create [PIN]"))
	}

	// Initialize blockchain client
	bcClient := blockchain.NewClient()
	var chainID int64 = 7441 // default ClawSwift
	
	if t.cfg != nil && t.cfg.Wallet.Enabled && len(t.cfg.Wallet.Chains) > 0 {
		chain := &t.cfg.Wallet.Chains[0]
		chainID = chain.ChainID
		if err := bcClient.AddChain(chain); err != nil {
			logger.ErrorCF("wallet_query", "Failed to connect to blockchain", map[string]any{"error": err.Error()})
			return ErrorResult(fmt.Sprintf("Blockchain connection failed: %v", err))
		}
	}

	tokenAddr := common.HexToAddress(tokenAddress)

	// Get token decimals
	decimals, err := bcClient.GetTokenDecimals(ctx, chainID, tokenAddr)
	if err != nil {
		decimals = 18 // default
		logger.WarnCF("wallet_query", "Failed to get decimals, using default", map[string]any{"error": err.Error()})
	}

	// Get token symbol
	symbol, err := bcClient.GetTokenSymbol(ctx, chainID, tokenAddr)
	if err != nil {
		symbol = "???"
		logger.WarnCF("wallet_query", "Failed to get symbol", map[string]any{"error": err.Error()})
	}

	// Get balance
	balance, err := bcClient.GetERC20Balance(ctx, chainID, tokenAddr, walletAddr)
	if err != nil {
		logger.ErrorCF("wallet_query", "Failed to get balance", map[string]any{"error": err.Error()})
		return ErrorResult(fmt.Sprintf("Failed to query balance: %v", err))
	}

	// Format balance
	balanceFloat := new(big.Float).SetInt(balance)
	divisor := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil))
	balanceFloat.Quo(balanceFloat, divisor)
	formatted := balanceFloat.Text('f', 4)

	logger.InfoCF("wallet_query", "Balance retrieved", map[string]any{
		"wallet":  walletAddr.Hex(),
		"balance": formatted,
		"symbol":  symbol,
	})

	result := fmt.Sprintf("💰 Wallet Balance\n\n👛 Address: `%s`\n🪙 Token: `%s`\n🏷️ Symbol: **%s**\n📊 Decimals: %d\n\n💵 Balance: `%s %s`",
		walletAddr.Hex(),
		tokenAddress,
		symbol,
		decimals,
		formatted,
		symbol,
	)

	return UserResult(result)
}

// getWalletAddress reads the wallet address from wallet.json
func (t *WalletQueryTool) getWalletAddress() (common.Address, error) {
	walletFile := filepath.Join(t.workspace, "wallet", "wallet.json")
	data, err := os.ReadFile(walletFile)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to read wallet.json: %w", err)
	}

	var walletData struct {
		Address string `json:"address"`
	}
	if err := json.Unmarshal(data, &walletData); err != nil {
		return common.Address{}, fmt.Errorf("failed to parse wallet.json: %w", err)
	}

	return common.HexToAddress(walletData.Address), nil
}
