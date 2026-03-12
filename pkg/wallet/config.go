package wallet

import (
	"fmt"
	"math/big"
)

// Config represents wallet configuration
type Config struct {
	Enabled bool         `json:"enabled"`
	Chains  []ChainConfig `json:"chains"`
}

// ChainConfig represents configuration for a blockchain network
type ChainConfig struct {
	Name         string `json:"name"`
	ChainID      int    `json:"chain_id"`
	RPC          string `json:"rpc"`
	Explorer     string `json:"explorer"`
	Currency     string `json:"currency"`
	IsNative     bool   `json:"is_native"`
	GasToken     string `json:"gas_token,omitempty"`
	GasTokenName string `json:"gas_token_name,omitempty"`
	Decimal      int    `json:"decimal"`
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if !c.Enabled {
		return nil // Skip validation if wallet is disabled
	}

	if len(c.Chains) == 0 {
		return fmt.Errorf("at least one chain must be configured")
	}

	for i, chain := range c.Chains {
		if err := chain.Validate(); err != nil {
			return fmt.Errorf("chain %d (%s): %w", i, chain.Name, err)
		}
	}

	return nil
}

// Validate checks if the chain configuration is valid
func (c *ChainConfig) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("chain name is required")
	}

	if c.ChainID <= 0 {
		return fmt.Errorf("chain_id must be positive")
	}

	if c.RPC == "" {
		return fmt.Errorf("rpc endpoint is required")
	}

	if c.Currency == "" {
		return fmt.Errorf("currency symbol is required")
	}

	if c.Decimal < 0 {
		return fmt.Errorf("decimal must be non-negative")
	}

	// For non-native chains, gas token is required
	if !c.IsNative && c.GasToken == "" {
		return fmt.Errorf("gas_token is required for non-native chains")
	}

	return nil
}

// GetChainByID returns chain configuration by chain ID
func (c *Config) GetChainByID(chainID int) (*ChainConfig, error) {
	for _, chain := range c.Chains {
		if chain.ChainID == chainID {
			return &chain, nil
		}
	}
	return nil, fmt.Errorf("chain with ID %d not found", chainID)
}

// GetChainByName returns chain configuration by name
func (c *Config) GetChainByName(name string) (*ChainConfig, error) {
	for _, chain := range c.Chains {
		if chain.Name == name {
			return &chain, nil
		}
	}
	return nil, fmt.Errorf("chain with name %s not found", name)
}

// GetDecimals returns the decimal places for the chain
func (c *ChainConfig) GetDecimals() int {
	if c.Decimal == 0 {
		return 18 // Default to 18 decimals (standard for ETH)
	}
	return c.Decimal
}

// ToWei converts token amount to wei units
func (c *ChainConfig) ToWei(amount float64) *big.Int {
	decimals := c.GetDecimals()
	multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	
	// Convert float to big.Int with proper precision
	weiAmount := new(big.Float).Mul(big.NewFloat(amount), new(big.Float).SetInt(multiplier))
	result := new(big.Int)
	weiAmount.Int(result)
	return result
}

// FromWei converts wei units to token amount
func (c *ChainConfig) FromWei(wei *big.Int) float64 {
	decimals := c.GetDecimals()
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	
	// Convert big.Int to float with proper precision
	amount := new(big.Float).Quo(new(big.Float).SetInt(wei), new(big.Float).SetInt(divisor))
	result, _ := amount.Float64()
	return result
}