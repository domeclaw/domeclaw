package blockchain

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// GetNativeBalance gets native token balance for an address
func (c *Client) GetNativeBalance(ctx context.Context, chainID int64, address common.Address) (*big.Int, error) {
	client, ok := c.GetClient(chainID)
	if !ok {
		return nil, fmt.Errorf("chain %d not found", chainID)
	}

	balance, err := client.BalanceAt(ctx, address, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return balance, nil
}

// GetTokenDecimals gets the decimals of an ERC20 token
func (c *Client) GetTokenDecimals(ctx context.Context, chainID int64, tokenAddress common.Address) (int32, error) {
	client, ok := c.GetClient(chainID)
	if !ok {
		return 18, fmt.Errorf("chain %d not found", chainID)
	}

	// decimals() function signature: 0x313ce567
	decimalsSig := []byte{0x31, 0x3c, 0xe5, 0x67}

	// Use direct eth_call
	var resultHex string
	err := client.Client().Call(&resultHex, "eth_call", map[string]interface{}{
		"to":   tokenAddress.Hex(),
		"data": common.Bytes2Hex(decimalsSig),
	}, "latest")

	if err != nil || len(resultHex) < 2 {
		return 18, fmt.Errorf("eth_call decimals failed: %w", err)
	}

	// Parse result - should be 32 bytes, decimals is the last byte
	result := common.FromHex(resultHex)
	if len(result) < 32 {
		return 18, fmt.Errorf("invalid decimals result length: %d", len(result))
	}

	// Decimals is the last byte
	decimals := int32(result[31])
	return decimals, nil
}

// GetERC20Balance gets ERC20 token balance for an address
func (c *Client) GetERC20Balance(ctx context.Context, chainID int64, tokenAddress, walletAddress common.Address) (*big.Int, error) {
	client, ok := c.GetClient(chainID)
	if !ok {
		return nil, fmt.Errorf("chain %d not found", chainID)
	}

	// ERC20 balanceOf function signature: balanceOf(address) returns (uint256)
	balanceOfSig := []byte{0x70, 0xa0, 0x82, 0x31}

	// Prepare call data: function selector + padded address
	callData := append(balanceOfSig, common.LeftPadBytes(walletAddress.Bytes(), 32)...)

	// Use direct eth_call instead of CallContract to avoid execution reverted
	var resultHex string
	err := client.Client().Call(&resultHex, "eth_call", map[string]interface{}{
		"to":   tokenAddress.Hex(),
		"data": common.Bytes2Hex(callData),
	}, "latest")

	if err != nil {
		return nil, fmt.Errorf("eth_call balanceOf failed: %w", err)
	}

	// Parse hex result
	result := common.FromHex(resultHex)
	if len(result) == 0 {
		return big.NewInt(0), nil
	}

	balance := new(big.Int).SetBytes(result)
	return balance, nil
}

// GetBalance gets balance (native or ERC20) based on chain config
func (c *Client) GetBalance(ctx context.Context, chainID int64, address common.Address) (*BalanceInfo, error) {
	chain, ok := c.GetChain(chainID)
	if !ok {
		return nil, fmt.Errorf("chain %d not found", chainID)
	}

	balanceInfo := &BalanceInfo{
		ChainID:   chainID,
		ChainName: chain.Name,
		Address:   address,
	}

	var balance *big.Int
	var err error

	if chain.IsNative {
		// Native token balance
		balance, err = c.GetNativeBalance(ctx, chainID, address)
		if err != nil {
			return nil, err
		}

		balanceInfo.Balance = balance
		balanceInfo.TokenName = chain.Currency
		balanceInfo.IsNative = true
		balanceInfo.Decimals = 18 // Default for native tokens
	} else {
		// ERC20 token balance
		if chain.GasToken == "" {
			return nil, fmt.Errorf("gas token not specified for chain %s", chain.Name)
		}

		tokenAddress := common.HexToAddress(chain.GasToken)

		// Get decimals from contract
		decimals, err := c.GetTokenDecimals(ctx, chainID, tokenAddress)
		if err != nil {
			// Fallback to 18 if can't get decimals
			decimals = 18
		}

		balance, err = c.GetERC20Balance(ctx, chainID, tokenAddress, address)
		if err != nil {
			return nil, err
		}

		balanceInfo.Balance = balance
		balanceInfo.TokenName = chain.GasTokenName
		balanceInfo.IsNative = false
		balanceInfo.TokenAddress = tokenAddress
		balanceInfo.Decimals = decimals
	}

	return balanceInfo, nil
}

// GetAllBalances returns balances for all configured chains
func (c *Client) GetAllBalances(ctx context.Context, address common.Address) ([]*BalanceInfo, error) {
	c.mu.RLock()
	chainIDs := make([]int64, 0, len(c.chains))
	for chainID := range c.chains {
		chainIDs = append(chainIDs, chainID)
	}
	c.mu.RUnlock()

	balances := make([]*BalanceInfo, 0, len(chainIDs))

	for _, chainID := range chainIDs {
		balance, err := c.GetBalance(ctx, chainID, address)
		if err != nil {
			// Log error but continue with other chains
			continue
		}
		balances = append(balances, balance)
	}

	return balances, nil
}

// BalanceInfo contains balance information
type BalanceInfo struct {
	ChainID      int64
	ChainName    string
	Address      common.Address
	Balance      *big.Int
	TokenName    string
	IsNative     bool
	TokenAddress common.Address
	Decimals     int32
}

// FormattedBalance returns balance as a formatted string
func (b *BalanceInfo) FormattedBalance() string {
	if b.Balance == nil {
		return "0"
	}

	// Convert to float with decimals
	balanceFloat := new(big.Float).SetInt(b.Balance)
	divisor := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(b.Decimals)), nil))
	balanceFloat.Quo(balanceFloat, divisor)

	return balanceFloat.String()
}
