package wallet

import (
	"context"
	"math/big"
	"os"
	"testing"
)

func TestWalletService(t *testing.T) {
	// Create a temporary workspace for testing
	tempDir, err := os.MkdirTemp("", "picoclaw-wallet-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test configuration
	config := Config{
		Enabled: true,
		Chains: []ChainConfig{
			{
				Name:     "TestChain",
				ChainID:  1337,
				RPC:      "http://localhost:8545", // This won't work, but we can test the config
				Explorer: "https://test-explorer.com",
				Currency: "TEST",
				IsNative: true,
				Decimal:  18,
			},
		},
	}

	// Create wallet service
	service := NewService(config, tempDir)

	// Test configuration validation
	t.Run("ValidateConfig", func(t *testing.T) {
		if err := config.Validate(); err != nil {
			t.Errorf("Config validation failed: %v", err)
		}
	})

	// Test wallet creation (this will work even without blockchain connection)
	t.Run("CreateWallet", func(t *testing.T) {
		ctx := context.Background()
		walletInfo, err := service.CreateWallet(ctx, "testpassword")
		if err != nil {
			t.Fatalf("Failed to create wallet: %v", err)
		}

		if walletInfo.Address == "" {
			t.Error("Wallet address should not be empty")
		}

		if walletInfo.Path == "" {
			t.Error("Wallet path should not be empty")
		}

		t.Logf("Created wallet: %s", walletInfo.Address)
	})

	// Test chain configuration
	t.Run("ChainConfig", func(t *testing.T) {
		chain, err := config.GetChainByID(1337)
		if err != nil {
			t.Errorf("Failed to get chain by ID: %v", err)
		}

		if chain.Name != "TestChain" {
			t.Errorf("Expected chain name 'TestChain', got '%s'", chain.Name)
		}

		// Test wei conversion
		amount := 1.5
		wei := chain.ToWei(amount)
		if wei.Cmp(big.NewInt(0)) <= 0 {
			t.Error("Wei conversion should produce positive value")
		}

		// Test from wei conversion
		convertedBack := chain.FromWei(wei)
		if convertedBack <= 0 {
			t.Error("From wei conversion should produce positive value")
		}
	})

	// Test error handling
	t.Run("ErrorHandling", func(t *testing.T) {
		// Test invalid password
		_, err := service.CreateWallet(context.Background(), "")
		if err == nil {
			t.Error("Expected error for empty password")
		}

		// Test invalid chain ID
		_, err = config.GetChainByID(999)
		if err == nil {
			t.Error("Expected error for invalid chain ID")
		}
	})
}

func TestChainConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  ChainConfig
		wantErr bool
	}{
		{
			name: "Valid native chain",
			config: ChainConfig{
				Name:     "Ethereum",
				ChainID:  1,
				RPC:      "https://mainnet.infura.io",
				Explorer: "https://etherscan.io",
				Currency: "ETH",
				IsNative: true,
				Decimal:  18,
			},
			wantErr: false,
		},
		{
			name: "Valid ERC20 chain",
			config: ChainConfig{
				Name:         "ClawSwift",
				ChainID:      7441,
				RPC:          "https://exp.clawswift.net/rpc",
				Explorer:     "https://exp.clawswift.net",
				Currency:     "CLAW",
				IsNative:     false,
				GasToken:     "0x20c0000000000000000000000000000000000000",
				GasTokenName: "CLAW",
				Decimal:      16,
			},
			wantErr: false,
		},
		{
			name: "Invalid - missing name",
			config: ChainConfig{
				ChainID:  1,
				RPC:      "https://mainnet.infura.io",
				Explorer: "https://etherscan.io",
				Currency: "ETH",
				IsNative: true,
				Decimal:  18,
			},
			wantErr: true,
		},
		{
			name: "Invalid - missing RPC",
			config: ChainConfig{
				Name:     "Ethereum",
				ChainID:  1,
				Explorer: "https://etherscan.io",
				Currency: "ETH",
				IsNative: true,
				Decimal:  18,
			},
			wantErr: true,
		},
		{
			name: "Invalid - non-native without gas token",
			config: ChainConfig{
				Name:     "ClawSwift",
				ChainID:  7441,
				RPC:      "https://exp.clawswift.net/rpc",
				Explorer: "https://exp.clawswift.net",
				Currency: "CLAW",
				IsNative: false,
				Decimal:  16,
			},
			wantErr: true,
		},
		{
			name: "Invalid - negative decimal",
			config: ChainConfig{
				Name:     "Ethereum",
				ChainID:  1,
				RPC:      "https://mainnet.infura.io",
				Explorer: "https://etherscan.io",
				Currency: "ETH",
				IsNative: true,
				Decimal:  -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}