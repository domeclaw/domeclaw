package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/wallet"
)

// QueryContractCallTool queries a smart contract function
type QueryContractCallTool struct {
	config *config.Config
}

func NewQueryContractCallTool(cfg *config.Config) *QueryContractCallTool {
	return &QueryContractCallTool{
		config: cfg,
	}
}

func (t *QueryContractCallTool) Name() string {
	return "query_contract_call"
}

func (t *QueryContractCallTool) Description() string {
	return "Queries a read-only smart contract function on ClawSwift network. " +
		"Use this when user asks to query contract data like 'เช็ค balanceOf', 'ดู totalSupply', 'call contract', etc."
}

func (t *QueryContractCallTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"contract_address": map[string]any{
				"type":        "string",
				"description": "Smart contract address (0x...)",
			},
			"abi_type": map[string]any{
				"type":        "string",
				"description": "ABI type/name (e.g., 'erc20', 'erc721')",
			},
			"method": map[string]any{
				"type":        "string",
				"description": "Method name to call (e.g., 'balanceOf', 'totalSupply')",
			},
			"params": map[string]any{
				"type":        "array",
				"description": "Optional method arguments",
				"items": map[string]any{
					"type": "string",
				},
			},
		},
		"required": []string{"contract_address", "abi_type", "method"},
	}
}

func (t *QueryContractCallTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	contractAddr, ok := args["contract_address"].(string)
	if !ok || contractAddr == "" {
		return ErrorResult("contract_address is required").WithError(fmt.Errorf("missing contract address"))
	}

	abiType, ok := args["abi_type"].(string)
	if !ok || abiType == "" {
		return ErrorResult("abi_type is required").WithError(fmt.Errorf("missing ABI type"))
	}

	method, ok := args["method"].(string)
	if !ok || method == "" {
		return ErrorResult("method is required").WithError(fmt.Errorf("missing method name"))
	}

	params, ok := args["params"].([]interface{})
	if !ok {
		params = []interface{}{}
	}

	// Create wallet service
	w := wallet.NewService(wallet.Config{
		Enabled: t.config.Wallet.Enabled,
		Chains:  convertChainConfigs(t.config.Wallet.Chains),
	}, t.config.Agents.Defaults.Workspace)
	if err := w.Initialize(ctx); err != nil {
		return ErrorResult(fmt.Sprintf("Failed to initialize wallet service: %v", err)).WithError(err)
	}
	defer w.Close()

	// Call contract method
	result, err := w.CallContractMethod(ctx, common.HexToAddress(contractAddr), abiType, method, params...)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to call contract method: %v", err)).WithError(err)
	}

	// Convert result to JSON string for LLM
	var jsonResult string
	switch v := result.(type) {
	case *big.Int:
		jsonResult = v.String()
	case string:
		jsonResult = v
	case int:
		jsonResult = fmt.Sprintf("%d", v)
	case uint8:
		jsonResult = fmt.Sprintf("%d", v)
	default:
		data, err := json.Marshal(v)
		if err != nil {
			jsonResult = fmt.Sprintf("%v", v)
		} else {
			jsonResult = string(data)
		}
	}

	return NewToolResult(jsonResult)
}

// ExecuteContractWriteTool executes a smart contract function that writes to blockchain
type ExecuteContractWriteTool struct {
	config *config.Config
}

func NewExecuteContractWriteTool(cfg *config.Config) *ExecuteContractWriteTool {
	return &ExecuteContractWriteTool{
		config: cfg,
	}
}

func (t *ExecuteContractWriteTool) Name() string {
	return "execute_contract_write"
}

func (t *ExecuteContractWriteTool) Description() string {
	return "Executes a write smart contract function on ClawSwift network. " +
		"Use this when user asks to write to contract like 'transfer tokens', 'approve', 'write contract', etc. " +
		"PIN is read automatically from workspace."
}

func (t *ExecuteContractWriteTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"contract_address": map[string]any{
				"type":        "string",
				"description": "Smart contract address (0x...)",
			},
			"abi_type": map[string]any{
				"type":        "string",
				"description": "ABI type/name",
			},
			"method": map[string]any{
				"type":        "string",
				"description": "Method name to call (e.g., 'transfer', 'approve')",
			},
			"value": map[string]any{
				"type":        "string",
				"description": "ETH value to send (use '0' for token transfers)",
			},
			"params": map[string]any{
				"type":        "array",
				"description": "Method arguments",
				"items": map[string]any{
					"type": "string",
				},
			},
		},
		"required": []string{"contract_address", "abi_type", "method", "value"},
	}
}

func (t *ExecuteContractWriteTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	contractAddr, ok := args["contract_address"].(string)
	if !ok || contractAddr == "" {
		return ErrorResult("contract_address is required").WithError(fmt.Errorf("missing contract address"))
	}

	abiType, ok := args["abi_type"].(string)
	if !ok || abiType == "" {
		return ErrorResult("abi_type is required").WithError(fmt.Errorf("missing ABI type"))
	}

	method, ok := args["method"].(string)
	if !ok || method == "" {
		return ErrorResult("method is required").WithError(fmt.Errorf("missing method name"))
	}

	password, ok := args["password"].(string)
	if !ok || password == "" {
		return ErrorResult("password is required").WithError(fmt.Errorf("missing password"))
	}

	valueStr, ok := args["value"].(string)
	var value *big.Int
	if ok && valueStr != "" {
		value = new(big.Int)
		value.SetString(valueStr, 10)
	} else {
		value = big.NewInt(0)
	}

	params, ok := args["params"].([]interface{})
	if !ok {
		params = []interface{}{}
	}

	// Create wallet service
	w := wallet.NewService(wallet.Config{
		Enabled: t.config.Wallet.Enabled,
		Chains:  convertChainConfigs(t.config.Wallet.Chains),
	}, t.config.Agents.Defaults.Workspace)
	if err := w.Initialize(ctx); err != nil {
		return ErrorResult(fmt.Sprintf("Failed to initialize wallet service: %v", err)).WithError(err)
	}
	defer w.Close()

	// Get accounts from keystore
	accounts := w.GetAccounts()
	if len(accounts) == 0 {
		return ErrorResult("No wallets found").WithError(fmt.Errorf("no wallets in keystore"))
	}

	// Use first account
	from := accounts[0].Address

	// Convert params to appropriate types
	var convertedParams []interface{}
	for _, p := range params {
		switch v := p.(type) {
		case string:
			// Handle hex addresses
			if strings.HasPrefix(strings.ToLower(v), "0x") && len(v) == 42 {
				convertedParams = append(convertedParams, common.HexToAddress(v))
			} else {
				// Try to parse as number
				bigInt := new(big.Int)
				if _, ok := bigInt.SetString(v, 10); ok {
					convertedParams = append(convertedParams, bigInt)
				} else {
					convertedParams = append(convertedParams, v)
				}
			}
		case float64:
			convertedParams = append(convertedParams, big.NewInt(int64(v)))
		default:
			convertedParams = append(convertedParams, v)
		}
	}

	// Execute contract method
	tx, err := w.ExecuteContractMethod(ctx, from, common.HexToAddress(contractAddr), abiType, method, value, password, convertedParams...)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to execute contract method: %v", err)).WithError(err)
	}

	return NewToolResult(tx.Hash().Hex())
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