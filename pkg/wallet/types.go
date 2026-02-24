package wallet

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// WalletInfo stores wallet metadata
type WalletInfo struct {
	Address   common.Address `json:"address"`
	CreatedAt time.Time      `json:"created_at"`
	Encrypted bool           `json:"encrypted"`
}

// PINStore stores the wallet PIN (4 digits)
type PINStore struct {
	PIN       string    `json:"pin"`
	CreatedAt time.Time `json:"created_at"`
}
