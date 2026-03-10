package commands

import (
	"context"
	"fmt"
	"strings"

	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/wallet"
)

func walletCommand() Definition {
	return Definition{
		Name:        "wallet",
		Description: "Manage Ethereum wallets and blockchain operations",
		Usage:       "/wallet [create|transfer|transfer_token|info|chain|switch|call|write] [arguments]",
		Handler: func(ctx context.Context, req Request, rt *Runtime) error {
			// Parse command arguments
			args := strings.SplitN(strings.TrimSpace(req.Text), " ", 2)
			var subcommand string
			var subargs string
			if len(args) > 1 {
				subcommand = strings.TrimSpace(args[1])
				if strings.Contains(subcommand, " ") {
					parts := strings.SplitN(subcommand, " ", 2)
					subcommand = parts[0]
					subargs = parts[1]
				}
			}

			switch subcommand {
			case "create":
				return handleWalletCreate(ctx, req, rt, subargs)
			case "transfer":
				return handleWalletTransfer(ctx, req, rt, subargs)
			case "transfer_token":
				return handleWalletTransferToken(ctx, req, rt, subargs)
			case "info":
				return handleWalletInfo(ctx, req, rt)
			case "chain":
				return handleWalletChain(ctx, req, rt, subargs)
			case "switch":
				return handleWalletSwitchChain(ctx, req, rt, subargs)
			case "call":
				return handleWalletCall(ctx, req, rt, subargs)
			case "write":
				return handleWalletWrite(ctx, req, rt, subargs)
			default:
				return req.Reply(`Usage: /wallet [command] [arguments]

Commands:
  create [password]          Create a new Ethereum wallet
  transfer [to] [amount] [password]  Transfer ETH to another address
  transfer_token [to] [amount] [password]  Transfer ERC20 token to another address
  info                       Show wallet information and balance
  chain                      List available chains
  switch [chain_name/ID]     Switch active chain
  call <contract> <abi> <method> [args]  Call contract read function
  write <contract> <abi> <method> <value> <password> [args]  Execute contract write function

Note: You must configure your wallet in config.json first.`)
			}
		},
	}
}

func handleWalletCreate(ctx context.Context, req Request, rt *Runtime, password string) error {
	password = strings.TrimSpace(password)
	if password == "" {
		return req.Reply("Usage: /wallet create [password]")
	}

	// Create wallet using wallet service
	w := wallet.NewService(wallet.Config{
		Enabled: rt.Config.Wallet.Enabled,
		Chains:  convertChainConfigs(rt.Config.Wallet.Chains),
	}, rt.Config.Agents.Defaults.Workspace)
	if err := w.Initialize(ctx); err != nil {
		return req.Reply(fmt.Sprintf("Error initializing wallet service: %v", err))
	}
	defer w.Close()

	info, err := w.CreateWallet(ctx, password)
	if err != nil {
		return req.Reply(fmt.Sprintf("Error creating wallet: %v", err))
	}

	return req.Reply(fmt.Sprintf("✅ Wallet created successfully!\n\nAddress: %s\n\nKeystore saved to your workspace.", info.Address))
}

func handleWalletTransfer(ctx context.Context, req Request, rt *Runtime, args string) error {
	// Split arguments: to, amount, password
	parts := strings.SplitN(strings.TrimSpace(args), " ", 3)
	if len(parts) < 3 {
		return req.Reply("Usage: /wallet transfer [to] [amount] [password]")
	}

	to := common.HexToAddress(parts[0])
	amountStr := strings.TrimSpace(parts[1])
	amount, err := parseAmount(amountStr)
	if err != nil {
		return req.Reply(fmt.Sprintf("Invalid amount: %v", err))
	}
	password := parts[2]

	w := wallet.NewService(wallet.Config{
		Enabled: rt.Config.Wallet.Enabled,
		Chains:  convertChainConfigs(rt.Config.Wallet.Chains),
	}, rt.Config.Agents.Defaults.Workspace)
	if err := w.Initialize(ctx); err != nil {
		return req.Reply(fmt.Sprintf("Error initializing wallet service: %v", err))
	}
	defer w.Close()

	// Get first account from keystore
	accounts := w.GetAccounts()
	if len(accounts) == 0 {
		return req.Reply("No wallets found. Create a wallet first with /wallet create [password]")
	}
	from := accounts[0].Address

	// Use first configured chain
	if len(rt.Config.Wallet.Chains) == 0 {
		return req.Reply("No chains configured")
	}
	chainID := rt.Config.Wallet.Chains[0].ChainID

	// Convert amount to wei based on chain decimal places
	chain, err := w.GetChainByID(chainID)
	if err != nil {
		return req.Reply(fmt.Sprintf("Error getting chain configuration: %v", err))
	}
	weiAmount := convertToWei(amount, chain.Decimal)

	tx, err := w.Transfer(ctx, from, to, weiAmount, password, chainID)
	if err != nil {
		return req.Reply(fmt.Sprintf("Error transferring: %v", err))
	}

	return req.Reply(fmt.Sprintf("✅ Transfer initiated!\n\nTransaction hash: %s", tx.Hash().Hex()))
}

func handleWalletTransferToken(ctx context.Context, req Request, rt *Runtime, args string) error {
	// Split arguments: to, amount, password
	parts := strings.SplitN(strings.TrimSpace(args), " ", 3)
	if len(parts) < 3 {
		return req.Reply("Usage: /wallet transfer_token [to] [amount] [password]")
	}

	to := common.HexToAddress(parts[0])
	amountStr := strings.TrimSpace(parts[1])
	amount, err := parseAmount(amountStr)
	if err != nil {
		return req.Reply(fmt.Sprintf("Invalid amount: %v", err))
	}
	password := parts[2]

	w := wallet.NewService(wallet.Config{
		Enabled: rt.Config.Wallet.Enabled,
		Chains:  convertChainConfigs(rt.Config.Wallet.Chains),
	}, rt.Config.Agents.Defaults.Workspace)
	if err := w.Initialize(ctx); err != nil {
		return req.Reply(fmt.Sprintf("Error initializing wallet service: %v", err))
	}
	defer w.Close()

	// Get first account from keystore
	accounts := w.GetAccounts()
	if len(accounts) == 0 {
		return req.Reply("No wallets found. Create a wallet first with /wallet create [password]")
	}
	from := accounts[0].Address

	// Use first configured chain
	if len(rt.Config.Wallet.Chains) == 0 {
		return req.Reply("No chains configured")
	}
	chainID := rt.Config.Wallet.Chains[0].ChainID

	// Convert amount to token units based on chain decimal places
	chain, err := w.GetChainByID(chainID)
	if err != nil {
		return req.Reply(fmt.Sprintf("Error getting chain configuration: %v", err))
	}
	tokenAmount := convertToWei(amount, chain.Decimal)

	var tx *types.Transaction

	if chain.IsNative {
		// For native chains, use Transfer method
		tx, err = w.Transfer(ctx, from, to, tokenAmount, password, chainID)
	} else {
		// For non-native chains, use TransferToken with gas token
		tx, err = w.TransferToken(ctx, from, to, tokenAmount, password, chainID, common.HexToAddress(chain.GasToken))
	}

	if err != nil {
		return req.Reply(fmt.Sprintf("Error transferring token: %v", err))
	}

	return req.Reply(fmt.Sprintf("✅ Token transfer initiated!\n\nTransaction hash: %s", tx.Hash().Hex()))
}

func handleWalletInfo(ctx context.Context, req Request, rt *Runtime) error {
	w := wallet.NewService(wallet.Config{
		Enabled: rt.Config.Wallet.Enabled,
		Chains:  convertChainConfigs(rt.Config.Wallet.Chains),
	}, rt.Config.Agents.Defaults.Workspace)
	if err := w.Initialize(ctx); err != nil {
		return req.Reply(fmt.Sprintf("Error initializing wallet service: %v", err))
	}
	defer w.Close()

	// Get first account from keystore
	accounts := w.GetAccounts()
	if len(accounts) == 0 {
		return req.Reply("No wallets found. Create a wallet first with /wallet create [password]")
	}
	address := accounts[0].Address

	// Use first configured chain
	if len(rt.Config.Wallet.Chains) == 0 {
		return req.Reply("No chains configured")
	}
	chainID := rt.Config.Wallet.Chains[0].ChainID

	info, err := w.GetWalletInfoForChain(ctx, address, chainID)
	if err != nil {
		return req.Reply(fmt.Sprintf("Error getting wallet info: %v", err))
	}

	// Format the response
	var resp strings.Builder
	resp.WriteString(fmt.Sprintf("Wallet Address: %s\n", info.Address))
	resp.WriteString(fmt.Sprintf("Chain: %s (ID: %d)\n", info.Chain, info.ChainID))

	// Format balance in human-readable format
	chain, err := w.GetChainByID(chainID)
	if err == nil {
		// Convert wei to token units
		divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(chain.Decimal)), nil)
		balance := new(big.Float).Quo(info.Balance, new(big.Float).SetInt(divisor))

		// Format to show appropriate decimal places
		formatStr := fmt.Sprintf("%%.%df", chain.Decimal)
		resp.WriteString(fmt.Sprintf("Balance: %s %s\n", fmt.Sprintf(formatStr, balance), chain.Currency))
	} else {
		resp.WriteString(fmt.Sprintf("Balance: %s Wei\n", info.Balance.String()))
	}

	return req.Reply(resp.String())
}

func handleWalletChain(ctx context.Context, req Request, rt *Runtime, args string) error {
	// List available chains
	if len(rt.Config.Wallet.Chains) == 0 {
		return req.Reply("No chains configured")
	}

	var resp strings.Builder
	resp.WriteString("Available Chains:\n")
	for i, chain := range rt.Config.Wallet.Chains {
		prefix := "  "
		if i == 0 {
			prefix = "✅ "
		}
		resp.WriteString(fmt.Sprintf("%s%s (ID: %d)\n", prefix, chain.Name, chain.ChainID))
		if !chain.IsNative {
			resp.WriteString(fmt.Sprintf("     Gas Token: %s (%s)\n", chain.GasToken, chain.GasTokenName))
		}
	}

	return req.Reply(resp.String())
}

func handleWalletSwitchChain(ctx context.Context, req Request, rt *Runtime, args string) error {
	// Switch active chain - currently we just use first chain, this is for future enhancement
	return req.Reply("Chain switching is not implemented yet. Currently using the first configured chain.")
}

func convertChainConfigs(configChains []config.ChainConfig) []wallet.ChainConfig {
	var walletChains []wallet.ChainConfig
	for _, chain := range configChains {
		walletChains = append(walletChains, wallet.ChainConfig{
			Name:         chain.Name,
			ChainID:      chain.ChainID,
			RPC:          chain.RPC,
			Explorer:     chain.Explorer,
			Currency:     chain.Currency,
			IsNative:     chain.IsNative,
			GasToken:     chain.GasToken,
			GasTokenName: chain.GasTokenName,
			Decimal:      chain.Decimal,
		})
	}
	return walletChains
}

func parseAmount(amountStr string) (float64, error) {
	return strconv.ParseFloat(amountStr, 64)
}

func convertToWei(amount float64, decimals int) *big.Int {
	multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	amountBig := new(big.Float).SetFloat64(amount)
	multiplierBig := new(big.Float).SetInt(multiplier)
	resultBig := new(big.Float).Mul(amountBig, multiplierBig)
	result := new(big.Int)
	resultBig.Int(result)
	return result
}
func handleWalletCall(ctx context.Context, req Request, rt *Runtime, args string) error {
	parts := strings.SplitN(strings.TrimSpace(args), " ", 3)
	if len(parts) < 3 {
		return req.Reply("Usage: /wallet call <contract_address> <abi_type> <method> [parameters]")
	}

	contractAddr := common.HexToAddress(parts[0])
	abiType := parts[1]
	method := parts[2]

	var params []interface{}
	if len(strings.Split(strings.TrimSpace(args), " ")) > 3 {
		paramsPart := strings.SplitN(args, " ", 4)[3]
		paramStrs := strings.SplitN(strings.TrimSpace(paramsPart), " ", -1)
		for _, p := range paramStrs {
			if strings.TrimSpace(p) != "" {
				params = append(params, strings.TrimSpace(p))
			}
		}
	}

	// Create wallet service
	w := wallet.NewService(wallet.Config{
		Enabled: rt.Config.Wallet.Enabled,
		Chains:  convertChainConfigs(rt.Config.Wallet.Chains),
	}, rt.Config.Agents.Defaults.Workspace)
	if err := w.Initialize(ctx); err != nil {
		return req.Reply(fmt.Sprintf("Error initializing wallet service: %v", err))
	}
	defer w.Close()

	// Convert params to appropriate types
	var convertedParams []interface{}
	for _, p := range params {
		strP := strings.TrimSpace(p.(string))
		if strings.HasPrefix(strings.ToLower(strP), "0x") && len(strP) == 42 {
			convertedParams = append(convertedParams, common.HexToAddress(strP))
		} else {
			bigInt := new(big.Int)
			if _, ok := bigInt.SetString(strP, 10); ok {
				convertedParams = append(convertedParams, bigInt)
			} else {
				convertedParams = append(convertedParams, strP)
			}
		}
	}

	// Call contract method
	result, err := w.CallContractMethod(ctx, contractAddr, abiType, method, convertedParams...)
	if err != nil {
		return req.Reply(fmt.Sprintf("Error calling contract method: %v", err))
	}

	return req.Reply(fmt.Sprintf("Result: %v", result))
}

func handleWalletWrite(ctx context.Context, req Request, rt *Runtime, args string) error {
	parts := strings.SplitN(strings.TrimSpace(args), " ", 5)
	if len(parts) < 5 {
		return req.Reply("Usage: /wallet write <contract_address> <abi_type> <method> <value> <password> [parameters]")
	}

	contractAddr := common.HexToAddress(parts[0])
	abiType := parts[1]
	method := parts[2]

	value := new(big.Int)
	value.SetString(strings.TrimSpace(parts[3]), 10)

	password := parts[4]

	var params []interface{}
	if len(strings.Split(strings.TrimSpace(args), " ")) > 5 {
		paramsPart := strings.SplitN(args, " ", 6)[5]
		paramStrs := strings.SplitN(strings.TrimSpace(paramsPart), " ", -1)
		for _, p := range paramStrs {
			if strings.TrimSpace(p) != "" {
				params = append(params, strings.TrimSpace(p))
			}
		}
	}

	// Create wallet service
	w := wallet.NewService(wallet.Config{
		Enabled: rt.Config.Wallet.Enabled,
		Chains:  convertChainConfigs(rt.Config.Wallet.Chains),
	}, rt.Config.Agents.Defaults.Workspace)
	if err := w.Initialize(ctx); err != nil {
		return req.Reply(fmt.Sprintf("Error initializing wallet service: %v", err))
	}
	defer w.Close()

	// Get accounts from keystore
	accounts := w.GetAccounts()
	if len(accounts) == 0 {
		return req.Reply("No wallets found")
	}
	from := accounts[0].Address

	// Convert params to appropriate types
	var convertedParams []interface{}
	for _, p := range params {
		strP := strings.TrimSpace(p.(string))
		if strings.HasPrefix(strings.ToLower(strP), "0x") && len(strP) == 42 {
			convertedParams = append(convertedParams, common.HexToAddress(strP))
		} else {
			bigInt := new(big.Int)
			if _, ok := bigInt.SetString(strP, 10); ok {
				convertedParams = append(convertedParams, bigInt)
			} else {
				convertedParams = append(convertedParams, strP)
			}
		}
	}

	// Execute contract method
	tx, err := w.ExecuteContractMethod(ctx, from, contractAddr, abiType, method, value, password, convertedParams...)
	if err != nil {
		return req.Reply(fmt.Sprintf("Error executing contract method: %v", err))
	}

	return req.Reply(fmt.Sprintf("Transaction sent! Hash: %s", tx.Hash().Hex()))
}
