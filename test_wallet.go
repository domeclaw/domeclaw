package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sipeed/domeclaw/pkg/config"
	"github.com/sipeed/domeclaw/pkg/wallet"
)

func main() {
	home, _ := os.UserHomeDir()
	workspace := filepath.Join(home, ".domeclaw", "workspace")

	fmt.Println("🧪 Testing Wallet Service...")
	fmt.Printf("Workspace: %s\n\n", workspace)

	walletCfg := &config.WalletConfig{
		Enabled: true,
		Chains: []config.EVMChain{
			{
				Name:       "ClawSwift",
				ChainID:    7441,
				RPC:        "https://exp.clawswift.net/rpc",
				Explorer:   "https://exp.clawswift.net",
				Currency:   "CLAW",
				IsNative:   false,
				GasToken:   "0x20c0000000000000000000000000000000000000",
				GasTokenName: "CLAW",
			},
		},
	}
	ws := wallet.NewWalletService(workspace, walletCfg)

	// Test 1: Check if wallet exists
	fmt.Println("Test 1: Check wallet exists")
	exists := ws.WalletExists()
	fmt.Printf("Wallet exists: %v\n\n", exists)

	if exists {
		fmt.Println("Wallet already exists, skipping creation test")
		addr, _ := ws.GetAddress()
		fmt.Printf("Address: %s\n", addr.Hex())
		return
	}

	// Test 2: Create wallet with PIN
	fmt.Println("Test 2: Create wallet with PIN 1234")
	address, err := ws.CreateWallet("1234")
	if err != nil {
		fmt.Printf("❌ Error: %v\n\n", err)
	} else {
		fmt.Printf("✅ Success! Address: %s\n\n", address.Hex())
	}

	// Test 3: Try to create again (should fail)
	fmt.Println("Test 3: Try to create again (should fail)")
	_, err = ws.CreateWallet("5678")
	if err != nil {
		fmt.Printf("✅ Expected error: %v\n\n", err)
	}

	// Test 4: Unlock with correct PIN
	fmt.Println("Test 4: Unlock with correct PIN (1234)")
	err = ws.Unlock("1234")
	if err != nil {
		fmt.Printf("❌ Error: %v\n\n", err)
	} else {
		fmt.Println("✅ Unlocked successfully!\n")
		ws.Lock()
	}

	// Test 5: Unlock with wrong PIN
	fmt.Println("Test 5: Unlock with wrong PIN (9999)")
	err = ws.Unlock("9999")
	if err != nil {
		fmt.Printf("✅ Expected error: %v\n\n", err)
	}

	// Test 6: Validate PIN
	fmt.Println("Test 6: Validate PIN formats")
	testPINs := []string{"1234", "0000", "9999", "123", "12345", "abcd"}
	for _, pin := range testPINs {
		valid := wallet.ValidatePIN(pin)
		fmt.Printf("  PIN '%s': %v\n", pin, valid)
	}
}
