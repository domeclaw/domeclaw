package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sipeed/domeclaw/pkg/blockchain"
	"github.com/sipeed/domeclaw/pkg/config"
	"github.com/sipeed/domeclaw/pkg/logger"
)

// WalletContractCallTool allows AI to call contract read functions
type WalletContractCallTool struct {
	workspace string
	cfg       *config.Config
}

// NewWalletContractCallTool creates a new contract call tool
func NewWalletContractCallTool(workspace string, cfg *config.Config) *WalletContractCallTool {
	return &WalletContractCallTool{
		workspace: workspace,
		cfg:       cfg,
	}
}

func (t *WalletContractCallTool) Name() string {
	return "query_contract_call"
}

func (t *WalletContractCallTool) Description() string {
	return "Call a read-only function on a smart contract. " +
		"Use this when user asks to query contract data like 'เช็ค balanceOf', 'ดู totalSupply', 'call contract', etc. " +
		"This does not require PIN as it's a read operation. " +
		"ABI must be uploaded first via /wallet abiupload."
}

func (t *WalletContractCallTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"contract_address": map[string]any{
				"type":        "string",
				"description": "Smart contract address (0x...)",
			},
			"abi_name": map[string]any{
				"type":        "string",
				"description": "Name of the uploaded ABI (e.g., 'erc20', 'uniswap-v2')",
			},
			"method": map[string]any{
				"type":        "string",
				"description": "Method name to call (e.g., 'balanceOf', 'totalSupply', 'symbol')",
			},
			"args": map[string]any{
				"type":        "array",
				"description": "Optional arguments for the method",
				"items": map[string]any{
					"type": "string",
				},
			},
		},
		"required": []string{"contract_address", "abi_name", "method"},
	}
}

func (t *WalletContractCallTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	contractAddress, _ := args["contract_address"].(string)
	abiName, _ := args["abi_name"].(string)
	method, _ := args["method"].(string)

	var callArgs []interface{}
	if argsList, ok := args["args"].([]interface{}); ok {
		for _, arg := range argsList {
			if s, ok := arg.(string); ok {
				// Try to parse as address or number
				if len(s) == 42 && strings.HasPrefix(s, "0x") {
					callArgs = append(callArgs, common.HexToAddress(s))
				} else if num, ok := new(big.Int).SetString(s, 10); ok {
					callArgs = append(callArgs, num)
				} else {
					callArgs = append(callArgs, s)
				}
			}
		}
	}

	logger.InfoCF("wallet_contract", "Contract call", map[string]any{
		"contract": contractAddress,
		"abi":      abiName,
		"method":   method,
	})

	// Validate address
	if len(contractAddress) != 42 || contractAddress[:2] != "0x" {
		return ErrorResult("Invalid contract address format")
	}

	// Initialize blockchain
	bcClient := blockchain.NewClient()
	var chainID int64 = 7441
	if t.cfg != nil && len(t.cfg.Wallet.Chains) > 0 {
		chain := &t.cfg.Wallet.Chains[0]
		chainID = chain.ChainID
		if err := bcClient.AddChain(chain); err != nil {
			return ErrorResult(fmt.Sprintf("Blockchain connection failed: %v", err))
		}
	}

	// Initialize ABI manager and contract service
	abiManager, err := blockchain.NewABIManager(t.workspace)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to initialize ABI manager: %v", err))
	}

	contractService := blockchain.NewContractService(bcClient, abiManager)
	contract := common.HexToAddress(contractAddress)

	// Execute call
	result, err := contractService.CallContract(ctx, chainID, contract, abiName, method, callArgs)
	if err != nil {
		logger.ErrorCF("wallet_contract", "Call failed", map[string]any{"error": err.Error()})
		return ErrorResult(fmt.Sprintf("Contract call failed: %v", err))
	}

	resultStr := fmt.Sprintf("%v", result)
	if result == nil {
		resultStr = "(no return value)"
	}

	output := fmt.Sprintf("📤 Contract Call Result\n\nContract: `%s`\nMethod: `%s`\nABI: `%s`\n\nResult: `%s`",
		contractAddress, method, abiName, resultStr)

	return UserResult(output)
}

// WalletContractWriteTool allows AI to execute contract write functions
type WalletContractWriteTool struct {
	workspace string
	cfg       *config.Config
}

// NewWalletContractWriteTool creates a new contract write tool
func NewWalletContractWriteTool(workspace string, cfg *config.Config) *WalletContractWriteTool {
	return &WalletContractWriteTool{
		workspace: workspace,
		cfg:       cfg,
	}
}

func (t *WalletContractWriteTool) Name() string {
	return "execute_contract_write"
}

func (t *WalletContractWriteTool) Description() string {
	return "Execute a state-changing function on a smart contract. " +
		"Use this when user asks to write to contract like 'transfer tokens', 'approve', 'write contract', etc. " +
		"This requires the wallet to be unlocked with PIN. " +
		"ABI must be uploaded first via /wallet abiupload."
}

func (t *WalletContractWriteTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"contract_address": map[string]any{
				"type":        "string",
				"description": "Smart contract address (0x...)",
			},
			"abi_name": map[string]any{
				"type":        "string",
				"description": "Name of the uploaded ABI",
			},
			"method": map[string]any{
				"type":        "string",
				"description": "Method name to call (e.g., 'transfer', 'approve')",
			},
			"value": map[string]any{
				"type":        "string",
				"description": "ETH value to send with transaction (use '0' for token transfers)",
			},
			"args": map[string]any{
				"type":        "array",
				"description": "Arguments for the method",
				"items": map[string]any{
					"type": "string",
				},
			},
		},
		"required": []string{"contract_address", "abi_name", "method", "value"},
	}
}

func (t *WalletContractWriteTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	contractAddress, _ := args["contract_address"].(string)
	abiName, _ := args["abi_name"].(string)
	method, _ := args["method"].(string)
	valueStr, _ := args["value"].(string)

	var callArgs []interface{}
	if argsList, ok := args["args"].([]interface{}); ok {
		for _, arg := range argsList {
			if s, ok := arg.(string); ok {
				if len(s) == 42 && strings.HasPrefix(s, "0x") {
					callArgs = append(callArgs, common.HexToAddress(s))
				} else if num, ok := new(big.Int).SetString(s, 10); ok {
					callArgs = append(callArgs, num)
				} else {
					callArgs = append(callArgs, s)
				}
			}
		}
	}

	logger.InfoCF("wallet_contract", "Contract write", map[string]any{
		"contract": contractAddress,
		"abi":      abiName,
		"method":   method,
	})

	// Validate address
	if len(contractAddress) != 42 || contractAddress[:2] != "0x" {
		return ErrorResult("Invalid contract address format")
	}

	// Read PIN
	pin, err := t.readPIN()
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to read PIN: %v", err))
	}

	// Parse value
	value := big.NewInt(0)
	if valueStr != "" && valueStr != "0" {
		if v, ok := new(big.Int).SetString(valueStr, 10); ok {
			value = v
		}
	}

	// Initialize keystore
	walletDir := filepath.Join(t.workspace, "wallet")
	ks := keystore.NewKeyStore(walletDir, keystore.StandardScryptN, keystore.StandardScryptP)

	accounts := ks.Accounts()
	if len(accounts) == 0 {
		return ErrorResult("No wallet found")
	}
	account := accounts[0]

	// Unlock
	if err := ks.Unlock(account, pin); err != nil {
		return ErrorResult(fmt.Sprintf("Failed to unlock wallet: %v", err))
	}
	defer ks.Lock(account.Address)

	// Initialize blockchain
	bcClient := blockchain.NewClient()
	var chainID int64 = 7441
	if t.cfg != nil && len(t.cfg.Wallet.Chains) > 0 {
		chain := &t.cfg.Wallet.Chains[0]
		chainID = chain.ChainID
		if err := bcClient.AddChain(chain); err != nil {
			return ErrorResult(fmt.Sprintf("Blockchain connection failed: %v", err))
		}
	}

	// Initialize services
	abiManager, err := blockchain.NewABIManager(t.workspace)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to initialize ABI manager: %v", err))
	}

	contractService := blockchain.NewContractService(bcClient, abiManager)
	contract := common.HexToAddress(contractAddress)

	// Create signer
	signer := func(ctx context.Context, chainID int64, tx *types.Transaction) (*types.Transaction, error) {
		chainIDBig := big.NewInt(chainID)
		return ks.SignTx(account, tx, chainIDBig)
	}

	// Execute write
	txHash, err := contractService.WriteContract(ctx, chainID, account.Address, contract, abiName, method, callArgs, value, signer)
	if err != nil {
		logger.ErrorCF("wallet_contract", "Write failed", map[string]any{"error": err.Error()})
		return ErrorResult(fmt.Sprintf("Contract write failed: %v", err))
	}

	output := fmt.Sprintf("✅ Transaction Sent!\n\n📤 Transaction Hash:\n`%s`", txHash.Hex())
	return UserResult(output)
}

func (t *WalletContractWriteTool) readPIN() (string, error) {
	pinFile := filepath.Join(t.workspace, "wallet", "pin.json")
	data, err := os.ReadFile(pinFile)
	if err != nil {
		return "", fmt.Errorf("failed to read pin.json: %w", err)
	}

	var pinData struct {
		PIN string `json:"pin"`
	}
	if err := json.Unmarshal(data, &pinData); err != nil {
		return "", fmt.Errorf("failed to parse pin.json: %w", err)
	}

	return pinData.PIN, nil
}
