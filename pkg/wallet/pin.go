package wallet

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// GeneratePIN generates a random 4-digit PIN
func GeneratePIN() (string, error) {
	max := big.NewInt(10000) // 0000-9999
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%04d", n.Int64()), nil
}

// ValidatePIN checks if PIN is valid 4-digit format
func ValidatePIN(pin string) bool {
	if len(pin) != 4 {
		return false
	}
	for _, c := range pin {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
