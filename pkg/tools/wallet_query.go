package tools

import (
	"context"
	"fmt"
	"math/big"

	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/logger"
	"github.com/sipeed/picoclaw/pkg/wallet"
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
		"Use this when user asks about their balance in natural language like 'เรามี balance เท่าไหร่', 'เช็คยอดเงิน', 'wallet มีเหรียญอะไรบ้าง', 'มีกี่เหรียญ', 'ดู wallet'. " +
		"Returns formatted balance with symbol and decimals."
}

func (t *WalletQueryTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"token_address": map[string]any{
				"type":        "string",
				"description": "Optional: ERC20 token contract address. If not provided, defaults to native chain token (CLAW)",
			},
		},
		"required": []string{},
	}
}

// Execute queries the wallet balance directly
func (t *WalletQueryTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	logger.InfoCF("wallet_query", "Querying balance", nil)

	// Create wallet service
	w := wallet.NewService(wallet.Config{
		Enabled: t.cfg.Wallet.Enabled,
		Chains:  convertChainConfigs(t.cfg.Wallet.Chains),
	}, t.workspace)
	if err := w.Initialize(ctx); err != nil {
		logger.ErrorCF("wallet_query", "Failed to initialize wallet service", map[string]any{"error": err.Error()})
		return ErrorResult(fmt.Sprintf("Failed to initialize wallet service: %v", err))
	}
	defer w.Close()

	// Get accounts from keystore
	accounts := w.GetAccounts()
	if len(accounts) == 0 {
		return ErrorResult("No wallet found. Please create one with /wallet create [password]")
	}
	walletAddr := accounts[0].Address

	// Use first configured chain
	if len(t.cfg.Wallet.Chains) == 0 {
		return ErrorResult("No chains configured")
	}
	chainID := t.cfg.Wallet.Chains[0].ChainID

	// Get chain info for explorer URL
	chain, err := w.GetChainByID(chainID)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to get chain info: %v", err))
	}

	// Get wallet info
	info, err := w.GetWalletInfoForChain(ctx, walletAddr, chainID)
	if err != nil {
		logger.ErrorCF("wallet_query", "Failed to get wallet info", map[string]any{"error": err.Error()})
		return ErrorResult(fmt.Sprintf("Failed to query wallet info: %v", err))
	}

	// Format balance
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(chain.Decimal)), nil)
	balanceFloat := new(big.Float).Quo(info.Balance, new(big.Float).SetInt(divisor))
	formatted := balanceFloat.Text('f', int(chain.Decimal))

	// Trim trailing zeros
	formatted = trimTrailingZeros(formatted)

	logger.InfoCF("wallet_query", "Balance retrieved", map[string]any{
		"wallet":  walletAddr.Hex(),
		"balance": formatted,
		"symbol":  chain.Currency,
	})

	// Build explorer link
	explorerURL := chain.Explorer
	if explorerURL == "" {
		explorerURL = "https://exp.clawswift.net"
	}
	addressLink := fmt.Sprintf("%s/address/%s", explorerURL, walletAddr.Hex())

	result := fmt.Sprintf("💰 **Wallet Balance**\n\n"+
		"👛 **Address:** `%s`\n"+
		"🔗 [View on Explorer](%s)\n\n"+
		"⛓️ **Chain:** %s (ID: %d)\n"+
		"🪙 **Symbol:** %s\n\n"+
		"💵 **Balance:** `%s %s`",
		walletAddr.Hex(),
		addressLink,
		chain.Name,
		chainID,
		chain.Currency,
		formatted,
		chain.Currency,
	)

	return UserResult(result)
}

// trimTrailingZeros removes trailing zeros from decimal string
func trimTrailingZeros(s string) string {
	if idx := len(s); idx > 0 {
		for idx > 0 && s[idx-1] == '0' {
			idx--
		}
		if idx > 0 && s[idx-1] == '.' {
			idx--
		}
		return s[:idx]
	}
	return s
}
