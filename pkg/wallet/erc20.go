package wallet

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ERC20 ABI for standard token functions
const erc20ABI = `[
	{
		"constant": true,
		"inputs": [{"name": "_owner", "type": "address"}],
		"name": "balanceOf",
		"outputs": [{"name": "balance", "type": "uint256"}],
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{"name": "_to", "type": "address"},
			{"name": "_value", "type": "uint256"}
		],
		"name": "transfer",
		"outputs": [{"name": "", "type": "bool"}],
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "decimals",
		"outputs": [{"name": "", "type": "uint8"}],
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "symbol",
		"outputs": [{"name": "", "type": "string"}],
		"type": "function"
	}
]`

var (
	erc20ABIJson abi.ABI
)

func init() {
	var err error
	erc20ABIJson, err = abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		panic(fmt.Sprintf("failed to parse ERC20 ABI: %v", err))
	}
}

// getERC20Balance gets the balance of an ERC20 token
func (s *Service) getERC20Balance(ctx context.Context, client *ethclient.Client, address common.Address, tokenAddress common.Address) (*big.Float, error) {
	// Create the call data
	data, err := erc20ABIJson.Pack("balanceOf", address)
	if err != nil {
		return nil, fmt.Errorf("failed to pack balanceOf call: %w", err)
	}

	// Call the contract
	result, err := client.CallContract(ctx, ethereum.CallMsg{
		To:   &tokenAddress,
		Data: data,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call balanceOf: %w", err)
	}

	// Unpack the result
	var balance *big.Int
	err = erc20ABIJson.UnpackIntoInterface(&balance, "balanceOf", result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack balanceOf result: %w", err)
	}

	// Convert to float with proper decimals
	return new(big.Float).SetInt(balance), nil
}

// buildERC20TransferData builds the data for an ERC20 transfer
func buildERC20TransferData(to common.Address, amount *big.Int) ([]byte, error) {
	return erc20ABIJson.Pack("transfer", to, amount)
}

// getERC20Decimals gets the decimals of an ERC20 token
func getERC20Decimals(ctx context.Context, client *ethclient.Client, tokenAddress common.Address) (uint8, error) {
	// Create the call data
	data, err := erc20ABIJson.Pack("decimals")
	if err != nil {
		return 0, fmt.Errorf("failed to pack decimals call: %w", err)
	}

	// Call the contract
	result, err := client.CallContract(ctx, ethereum.CallMsg{
		To:   &tokenAddress,
		Data: data,
	}, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to call decimals: %w", err)
	}

	// Unpack the result
	var decimals uint8
	err = erc20ABIJson.UnpackIntoInterface(&decimals, "decimals", result)
	if err != nil {
		return 0, fmt.Errorf("failed to unpack decimals result: %w", err)
	}

	return decimals, nil
}

// getERC20Symbol gets the symbol of an ERC20 token
func getERC20Symbol(ctx context.Context, client *ethclient.Client, tokenAddress common.Address) (string, error) {
	// Create the call data
	data, err := erc20ABIJson.Pack("symbol")
	if err != nil {
		return "", fmt.Errorf("failed to pack symbol call: %w", err)
	}

	// Call the contract
	result, err := client.CallContract(ctx, ethereum.CallMsg{
		To:   &tokenAddress,
		Data: data,
	}, nil)
	if err != nil {
		return "", fmt.Errorf("failed to call symbol: %w", err)
	}

	// Unpack the result
	var symbol string
	err = erc20ABIJson.UnpackIntoInterface(&symbol, "symbol", result)
	if err != nil {
		return "", fmt.Errorf("failed to unpack symbol result: %w", err)
	}

	return symbol, nil
}

// ERC20Token represents ERC20 token information
type ERC20Token struct {
	Address  common.Address
	Symbol   string
	Decimals uint8
}